package discord

// User represents a Discord user.
type User struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Bot           bool   `json:"bot,omitempty"`
}

// Guild represents a Discord server.
type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Channel represents a Discord channel.
type Channel struct {
	ID      string `json:"id"`
	GuildID string `json:"guild_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Type    int    `json:"type"`
}

// Message represents a Discord message.
type Message struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	GuildID   string `json:"guild_id,omitempty"`
	Author    *User  `json:"author,omitempty"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
}

// MessageCreate is the payload for creating a new message.
type MessageCreate struct {
	Content string `json:"content"`
}

// GuildScheduledEvent represents a Discord scheduled event.
type GuildScheduledEvent struct {
	ID                 string `json:"id"`
	GuildID            string `json:"guild_id"`
	ChannelID          string `json:"channel_id,omitempty"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	ScheduledStartTime string `json:"scheduled_start_time"`
	ScheduledEndTime   string `json:"scheduled_end_time,omitempty"`
	EntityType         int    `json:"entity_type"`
	Status             int    `json:"status,omitempty"`
}

// GuildScheduledEventCreate is the payload for creating a scheduled event.
type GuildScheduledEventCreate struct {
	ChannelID          string `json:"channel_id,omitempty"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	ScheduledStartTime string `json:"scheduled_start_time"`
	ScheduledEndTime   string `json:"scheduled_end_time,omitempty"`
	EntityType         int    `json:"entity_type"`
	PrivacyLevel       int    `json:"privacy_level"`
}
