package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	gatewayURL = "wss://gateway.discord.gg/?v=10&encoding=json"

	opDispatch       = 0
	opHeartbeat      = 1
	opIdentify       = 2
	opResume         = 6
	opReconnect      = 7
	opInvalidSession = 9
	opHello          = 10
	opHeartbeatAck   = 11
)

// Intents for Discord Gateway.
const (
	IntentGuilds                = 1 << 0
	IntentGuildMembers          = 1 << 1
	IntentGuildMessages         = 1 << 9
	IntentGuildMessageReactions = 1 << 10
	IntentDirectMessages        = 1 << 12
	IntentMessageContent        = 1 << 15
)

// GatewayPayload is the base structure for gateway messages.
type GatewayPayload struct {
	Op        int             `json:"op"`
	Data      json.RawMessage `json:"d,omitempty"`
	Sequence  *int64          `json:"s,omitempty"`
	EventName string          `json:"t,omitempty"`
}

// HelloData is received after connecting.
type HelloData struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

// IdentifyData is sent to authenticate.
type IdentifyData struct {
	Token      string             `json:"token"`
	Intents    int                `json:"intents"`
	Properties IdentifyProperties `json:"properties"`
}

type IdentifyProperties struct {
	OS      string `json:"os"`
	Browser string `json:"browser"`
	Device  string `json:"device"`
}

// ReadyData is received after successful identification.
type ReadyData struct {
	SessionID string `json:"session_id"`
	User      *User  `json:"user"`
}

// Gateway manages the WebSocket connection to Discord.
type Gateway struct {
	token     string
	intents   int
	conn      *websocket.Conn
	sessionID string
	sequence  int64
	mu        sync.Mutex

	handlers  map[string][]func(json.RawMessage)
	handlerMu sync.RWMutex
}

// NewGateway creates a new Gateway connection manager.
func NewGateway(token string, intents int) *Gateway {
	return &Gateway{
		token:    token,
		intents:  intents,
		handlers: make(map[string][]func(json.RawMessage)),
	}
}

// On registers an event handler.
func (g *Gateway) On(event string, handler func(json.RawMessage)) {
	g.handlerMu.Lock()
	defer g.handlerMu.Unlock()
	g.handlers[event] = append(g.handlers[event], handler)
}

// Connect establishes the WebSocket connection and starts the event loop.
func (g *Gateway) Connect(ctx context.Context) error {
	conn, resp, err := websocket.Dial(ctx, gatewayURL, nil)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}
	g.conn = conn

	go g.readLoop(ctx)

	<-ctx.Done()
	return g.conn.Close(websocket.StatusNormalClosure, "shutting down")
}

func (g *Gateway) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var payload GatewayPayload
			if err := wsjson.Read(ctx, g.conn, &payload); err != nil {
				log.Printf("gateway read error: %v", err)
				return
			}
			g.handlePayload(ctx, &payload)
		}
	}
}

func (g *Gateway) handlePayload(ctx context.Context, p *GatewayPayload) {
	if p.Sequence != nil {
		g.mu.Lock()
		g.sequence = *p.Sequence
		g.mu.Unlock()
	}

	switch p.Op {
	case opHello:
		var hello HelloData
		if err := json.Unmarshal(p.Data, &hello); err != nil {
			log.Printf("failed to parse hello: %v", err)
			return
		}
		go g.heartbeatLoop(ctx, time.Duration(hello.HeartbeatInterval)*time.Millisecond)
		g.identify(ctx)

	case opHeartbeat:
		g.sendHeartbeat(ctx)

	case opHeartbeatAck:
		// Heartbeat acknowledged

	case opReconnect:
		log.Println("gateway requested reconnect")

	case opInvalidSession:
		log.Println("invalid session, re-identifying")
		time.Sleep(time.Second)
		g.identify(ctx)

	case opDispatch:
		g.dispatch(p.EventName, p.Data)
	}
}

func (g *Gateway) dispatch(event string, data json.RawMessage) {
	if event == "READY" {
		var ready ReadyData
		if err := json.Unmarshal(data, &ready); err != nil {
			log.Printf("failed to parse ready: %v", err)
			return
		}
		g.mu.Lock()
		g.sessionID = ready.SessionID
		g.mu.Unlock()
		log.Printf("connected as %s#%s", ready.User.Username, ready.User.Discriminator)
	}

	g.handlerMu.RLock()
	handlers := g.handlers[event]
	g.handlerMu.RUnlock()

	for _, h := range handlers {
		go h(data)
	}
}

func (g *Gateway) identify(ctx context.Context) {
	identify := IdentifyData{
		Token:   g.token,
		Intents: g.intents,
		Properties: IdentifyProperties{
			OS:      "linux",
			Browser: "discord-cncf-bots",
			Device:  "discord-cncf-bots",
		},
	}

	data, err := json.Marshal(identify)
	if err != nil {
		log.Printf("failed to marshal identify: %v", err)
		return
	}

	g.send(ctx, opIdentify, data)
}

func (g *Gateway) heartbeatLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.sendHeartbeat(ctx)
		}
	}
}

func (g *Gateway) sendHeartbeat(ctx context.Context) {
	g.mu.Lock()
	seq := g.sequence
	g.mu.Unlock()

	data, _ := json.Marshal(seq)
	g.send(ctx, opHeartbeat, data)
}

func (g *Gateway) send(ctx context.Context, op int, data json.RawMessage) error {
	payload := GatewayPayload{
		Op:   op,
		Data: data,
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.conn == nil {
		return errors.New("not connected")
	}

	return wsjson.Write(ctx, g.conn, payload)
}
