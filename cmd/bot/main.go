package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/kenesparta/discord-cnayp-bots/internal/bot"
	"github.com/kenesparta/discord-cnayp-bots/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	b, err := bot.New(cfg)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := b.Run(ctx); err != nil {
		log.Fatalf("bot error: %v", err)
	}

	log.Println("bot shutdown complete")
}
