package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "https://discord.com/api/v10"

type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a new Discord REST API client.
func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{},
	}
}

// SendMessage sends a message to a channel.
func (c *Client) SendMessage(ctx context.Context, channelID, content string) (*Message, error) {
	body := MessageCreate{Content: content}
	endpoint := fmt.Sprintf("/channels/%s/messages", channelID)

	var msg Message
	if err := c.do(ctx, http.MethodPost, endpoint, body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// GetChannel retrieves a channel by ID.
func (c *Client) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	endpoint := fmt.Sprintf("/channels/%s", channelID)

	var ch Channel
	if err := c.do(ctx, http.MethodGet, endpoint, nil, &ch); err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetGuildChannels retrieves all channels in a guild.
func (c *Client) GetGuildChannels(ctx context.Context, guildID string) ([]Channel, error) {
	endpoint := fmt.Sprintf("/guilds/%s/channels", guildID)

	var channels []Channel
	if err := c.do(ctx, http.MethodGet, endpoint, nil, &channels); err != nil {
		return nil, err
	}
	return channels, nil
}

// CreateScheduledEvent creates a new scheduled event in a guild.
func (c *Client) CreateScheduledEvent(ctx context.Context, guildID string, event *GuildScheduledEventCreate) (*GuildScheduledEvent, error) {
	endpoint := fmt.Sprintf("/guilds/%s/scheduled-events", guildID)

	var result GuildScheduledEvent
	if err := c.do(ctx, http.MethodPost, endpoint, event, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) do(ctx context.Context, method, endpoint string, body, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bot "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "DiscordBot (https://github.com/kenesparta/discord-cnayp-bots, 1.0.0)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord api error (status %d): %s", resp.StatusCode, respBody)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
