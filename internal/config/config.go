package config

import (
	"errors"
	"os"
)

type Config struct {
	Token        string
	GuildID      string
	SchedulePath string
}

func Load() (*Config, error) {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		return nil, errors.New("DISCORD_BOT_TOKEN environment variable is required")
	}

	guildID := os.Getenv("DISCORD_GUILD_ID")
	if guildID == "" {
		return nil, errors.New("DISCORD_GUILD_ID environment variable is required")
	}

	schedulePath := os.Getenv("DISCORD_SCHEDULE_PATH")
	if schedulePath == "" {
		schedulePath = "config/schedules.json"
	}

	return &Config{
		Token:        token,
		GuildID:      guildID,
		SchedulePath: schedulePath,
	}, nil
}
