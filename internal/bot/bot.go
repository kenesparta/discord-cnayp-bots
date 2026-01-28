package bot

import (
	"context"
	"encoding/json"
	"log"

	"github.com/kenesparta/discord-cnayp-bots/internal/config"
	"github.com/kenesparta/discord-cnayp-bots/internal/discord"
	"github.com/kenesparta/discord-cnayp-bots/internal/scheduler"
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

	if msg.Content == "!ping" {
		_, err := b.client.SendMessage(ctx, msg.ChannelID, "pong!")
		if err != nil {
			log.Printf("failed to send message: %v", err)
		}
	}
}
