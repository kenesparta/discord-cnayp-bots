package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/kenesparta/discord-cncf-bots/internal/config"
	"github.com/kenesparta/discord-cncf-bots/internal/discord"
	"github.com/kenesparta/discord-cncf-bots/internal/scheduler"
)

type Bot struct {
	client    *discord.Client
	gateway   *discord.Gateway
	scheduler *scheduler.Scheduler
}

// New creates a new Bot instance.
func New(cfg *config.Config) (*Bot, error) {
	intents := discord.IntentGuilds |
		discord.IntentGuildMessages |
		discord.IntentMessageContent

	client := discord.NewClient(cfg.Token)

	return &Bot{
		client:    client,
		gateway:   discord.NewGateway(cfg.Token, intents),
		scheduler: scheduler.New(client, cfg.GuildID, cfg.SchedulePath),
	}, nil
}

// Run starts the bot and blocks until the context is canceled.
func (b *Bot) Run(ctx context.Context) error {
	b.registerHandlers()
	go b.scheduler.Run(ctx)
	log.Println("starting bot...")
	return b.gateway.Connect(ctx)
}

func (b *Bot) registerHandlers() {
	b.gateway.On("MESSAGE_CREATE", b.onMessageCreate)
}

func (b *Bot) onMessageCreate(data json.RawMessage) {
	var msg discord.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("failed to parse message: %v", err)
		return
	}

	if msg.Author != nil && msg.Author.Bot {
		return
	}

	ctx := context.Background()

	switch {
	case msg.Content == "!ping":
		_, err := b.client.SendMessage(ctx, msg.ChannelID, "pong!")
		if err != nil {
			log.Printf("failed to send message: %v", err)
		}

	case msg.Content == "!schedule":
		names := b.scheduler.ListSchedules()
		if len(names) == 0 {
			b.client.SendMessage(ctx, msg.ChannelID, "No schedules configured.")
			return
		}
		reply := "**Available schedules:**\n"
		for i, name := range names {
			reply += fmt.Sprintf("`%d` - %s\n", i+1, name)
		}
		reply += "\nUsage: `!schedule <number>`"
		b.client.SendMessage(ctx, msg.ChannelID, reply)

	case strings.HasPrefix(msg.Content, "!schedule "):
		arg := strings.TrimPrefix(msg.Content, "!schedule ")
		arg = strings.TrimSpace(arg)
		if arg == "" {
			b.client.SendMessage(ctx, msg.ChannelID, "Usage: `!schedule <number>`")
			return
		}

		index, err := strconv.Atoi(arg)
		if err != nil {
			b.client.SendMessage(ctx, msg.ChannelID, "Invalid number. Use `!schedule` to see available options.")
			return
		}

		event, _, err := b.scheduler.CreateEventByIndex(ctx, index)
		if err != nil {
			log.Printf("failed to create event: %v", err)
			b.client.SendMessage(ctx, msg.ChannelID, "Failed to create event: "+err.Error())
			return
		}

		b.client.SendMessage(ctx, msg.ChannelID, "Created scheduled event: **"+event.Name+"**")
	}
}
