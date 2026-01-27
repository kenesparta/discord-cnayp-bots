package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kenesparta/discord-cncf-bots/internal/discord"
)

// Schedule represents a scheduled event configuration.
type Schedule struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	VoiceChannel    string   `json:"voice_channel"`
	NotifyChannel   string   `json:"notify_channel"`
	Days            []string `json:"days"`
	Time            string   `json:"time"`
	Timezone        string   `json:"timezone"`
	DurationMinutes int      `json:"duration_minutes"`
}

// ScheduleConfig is the root configuration for schedules.
type ScheduleConfig struct {
	Schedules       []Schedule `json:"schedules"`
	DigestTime      string     `json:"digest_time"`
	DigestChannel   string     `json:"digest_channel"`
	ReminderMinutes []int      `json:"reminder_minutes"`
}

// Scheduler manages scheduled Discord events.
type Scheduler struct {
	client     *discord.Client
	guildID    string
	configPath string

	schedules       []Schedule
	digestTime      string
	digestChannel   string
	reminderMinutes []int

	channelCache  map[string]string
	lastCreated   map[string]string // schedule name -> date (YYYY-MM-DD)
	lastDigest    string            // date of last digest (YYYY-MM-DD)
	sentReminders map[string]bool   // "scheduleName:date:minutes" -> true
	mu            sync.RWMutex
}

// New creates a new Scheduler instance.
func New(client *discord.Client, guildID, configPath string) *Scheduler {
	return &Scheduler{
		client:          client,
		guildID:         guildID,
		configPath:      configPath,
		channelCache:    make(map[string]string),
		lastCreated:     make(map[string]string),
		sentReminders:   make(map[string]bool),
		reminderMinutes: []int{60, 15}, // default: 1 hour and 15 min before
	}
}

// Load reads and parses the schedule configuration file.
func (s *Scheduler) Load() error {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("read schedule config: %w", err)
	}

	var config ScheduleConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("parse schedule config: %w", err)
	}

	s.mu.Lock()
	s.schedules = config.Schedules
	s.digestTime = config.DigestTime
	s.digestChannel = config.DigestChannel
	if len(config.ReminderMinutes) > 0 {
		s.reminderMinutes = config.ReminderMinutes
	}
	s.mu.Unlock()

	log.Printf("loaded %d schedules from %s", len(config.Schedules), s.configPath)
	return nil
}

// Run starts the scheduler loop. It checks every minute for schedules to trigger.
func (s *Scheduler) Run(ctx context.Context) {
	if err := s.Load(); err != nil {
		log.Printf("failed to load schedules: %v", err)
		return
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return
		case <-ticker.C:
			s.checkAll(ctx)
		}
	}
}

func (s *Scheduler) checkAll(ctx context.Context) {
	s.checkDigest(ctx)
	s.checkReminders(ctx)
	s.checkSchedules(ctx)
}

func (s *Scheduler) checkDigest(ctx context.Context) {
	s.mu.RLock()
	digestTime := s.digestTime
	digestChannel := s.digestChannel
	schedules := s.schedules
	lastDigest := s.lastDigest
	s.mu.RUnlock()

	if digestTime == "" || digestChannel == "" {
		return
	}

	if len(schedules) == 0 {
		return
	}

	loc, err := time.LoadLocation(schedules[0].Timezone)
	if err != nil {
		return
	}

	now := time.Now().In(loc)
	dateKey := now.Format("2006-01-02")

	if lastDigest == dateKey {
		return
	}

	parts := strings.Split(digestTime, ":")
	if len(parts) != 2 {
		return
	}

	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)

	if now.Hour() != hour || now.Minute() != minute {
		return
	}

	s.mu.Lock()
	s.lastDigest = dateKey
	s.mu.Unlock()

	s.sendDailyDigest(ctx, now)
}

func (s *Scheduler) sendDailyDigest(ctx context.Context, now time.Time) {
	s.mu.RLock()
	schedules := s.schedules
	digestChannel := s.digestChannel
	s.mu.RUnlock()

	dayName := strings.ToLower(now.Weekday().String())
	var todayEvents []Schedule

	for _, sch := range schedules {
		for _, d := range sch.Days {
			if strings.ToLower(d) == dayName {
				todayEvents = append(todayEvents, sch)
				break
			}
		}
	}

	channelID, err := s.resolveChannelID(ctx, digestChannel)
	if err != nil {
		log.Printf("failed to resolve digest channel: %v", err)
		return
	}

	var msg string
	if len(todayEvents) == 0 {
		msg = fmt.Sprintf("**Daily Schedule - %s**\n\nNo events scheduled for today.", now.Format("Monday, January 2"))
	} else {
		msg = fmt.Sprintf("**Daily Schedule - %s**\n\n", now.Format("Monday, January 2"))
		for _, evt := range todayEvents {
			msg += fmt.Sprintf("- **%s** at %s (%d min)\n", evt.Name, evt.Time, evt.DurationMinutes)
		}
	}

	if _, err := s.client.SendMessage(ctx, channelID, msg); err != nil {
		log.Printf("failed to send daily digest: %v", err)
	} else {
		log.Println("sent daily digest")
	}
}

func (s *Scheduler) checkReminders(ctx context.Context) {
	s.mu.RLock()
	schedules := s.schedules
	reminderMinutes := s.reminderMinutes
	s.mu.RUnlock()

	now := time.Now()

	for _, schedule := range schedules {
		loc, err := time.LoadLocation(schedule.Timezone)
		if err != nil {
			continue
		}

		localNow := now.In(loc)
		dayName := strings.ToLower(localNow.Weekday().String())

		dayMatch := false
		for _, d := range schedule.Days {
			if strings.ToLower(d) == dayName {
				dayMatch = true
				break
			}
		}
		if !dayMatch {
			continue
		}

		parts := strings.Split(schedule.Time, ":")
		if len(parts) != 2 {
			continue
		}

		var hour, minute int
		fmt.Sscanf(parts[0], "%d", &hour)
		fmt.Sscanf(parts[1], "%d", &minute)

		eventTime := time.Date(
			localNow.Year(), localNow.Month(), localNow.Day(),
			hour, minute, 0, 0, loc,
		)

		for _, mins := range reminderMinutes {
			reminderTime := eventTime.Add(-time.Duration(mins) * time.Minute)
			if localNow.Hour() == reminderTime.Hour() && localNow.Minute() == reminderTime.Minute() {
				dateKey := localNow.Format("2006-01-02")
				reminderKey := fmt.Sprintf("%s:%s:%d", schedule.Name, dateKey, mins)

				s.mu.RLock()
				sent := s.sentReminders[reminderKey]
				s.mu.RUnlock()

				if !sent {
					go s.sendReminder(ctx, schedule, eventTime, mins, reminderKey)
				}
			}
		}
	}
}

func (s *Scheduler) sendReminder(ctx context.Context, schedule Schedule, eventTime time.Time, minutesBefore int, reminderKey string) {
	s.mu.Lock()
	s.sentReminders[reminderKey] = true
	s.mu.Unlock()

	channelID, err := s.resolveChannelID(ctx, schedule.NotifyChannel)
	if err != nil {
		log.Printf("failed to resolve channel for reminder: %v", err)
		return
	}

	voiceChannelID, _ := s.resolveChannelID(ctx, schedule.VoiceChannel)

	var timeText string
	if minutesBefore >= 60 {
		hours := minutesBefore / 60
		if hours == 1 {
			timeText = "1 hour"
		} else {
			timeText = fmt.Sprintf("%d hours", hours)
		}
	} else {
		timeText = fmt.Sprintf("%d minutes", minutesBefore)
	}

	msg := fmt.Sprintf(`⏰ @everyone **Reminder:** %s starts in %s!

⏱️ **Duration:** %d minutes
%s

📍 Join us in <#%s>`,
		schedule.Name,
		timeText,
		schedule.DurationMinutes,
		schedule.Description,
		voiceChannelID,
	)

	if _, err := s.client.SendMessage(ctx, channelID, msg); err != nil {
		log.Printf("failed to send reminder: %v", err)
	} else {
		log.Printf("sent %s reminder for %s", timeText, schedule.Name)
	}
}

func (s *Scheduler) checkSchedules(ctx context.Context) {
	s.mu.RLock()
	schedules := s.schedules
	s.mu.RUnlock()

	now := time.Now()

	for _, schedule := range schedules {
		if s.shouldTrigger(schedule, now) {
			go s.triggerSchedule(ctx, schedule, now)
		}
	}
}

func (s *Scheduler) shouldTrigger(schedule Schedule, now time.Time) bool {
	loc, err := time.LoadLocation(schedule.Timezone)
	if err != nil {
		log.Printf("invalid timezone %s for schedule %s: %v", schedule.Timezone, schedule.Name, err)
		return false
	}

	localNow := now.In(loc)

	// Check if tomorrow matches a scheduled day
	tomorrow := localNow.AddDate(0, 0, 1)
	tomorrowDayName := strings.ToLower(tomorrow.Weekday().String())

	dayMatch := false
	for _, d := range schedule.Days {
		if strings.ToLower(d) == tomorrowDayName {
			dayMatch = true
			break
		}
	}
	if !dayMatch {
		return false
	}

	parts := strings.Split(schedule.Time, ":")
	if len(parts) != 2 {
		log.Printf("invalid time format %s for schedule %s", schedule.Time, schedule.Name)
		return false
	}

	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)

	// Trigger at the same time one day before
	if localNow.Hour() != hour || localNow.Minute() != minute {
		return false
	}

	// Use tomorrow's date as the key to prevent duplicate event creation
	dateKey := tomorrow.Format("2006-01-02")
	s.mu.RLock()
	lastDate := s.lastCreated[schedule.Name]
	s.mu.RUnlock()

	if lastDate == dateKey {
		return false
	}

	return true
}

func (s *Scheduler) triggerSchedule(ctx context.Context, schedule Schedule, now time.Time) {
	log.Printf("triggering schedule: %s", schedule.Name)

	loc, _ := time.LoadLocation(schedule.Timezone)
	localNow := now.In(loc)

	// Event is for tomorrow
	tomorrow := localNow.AddDate(0, 0, 1)
	dateKey := tomorrow.Format("2006-01-02")

	s.mu.Lock()
	s.lastCreated[schedule.Name] = dateKey
	s.mu.Unlock()

	voiceChannelID, err := s.resolveChannelID(ctx, schedule.VoiceChannel)
	if err != nil {
		log.Printf("failed to resolve voice channel %s: %v", schedule.VoiceChannel, err)
		return
	}

	notifyChannelID, err := s.resolveChannelID(ctx, schedule.NotifyChannel)
	if err != nil {
		log.Printf("failed to resolve notify channel %s: %v", schedule.NotifyChannel, err)
		return
	}

	parts := strings.Split(schedule.Time, ":")
	var hour, minute int
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)

	startTime := time.Date(
		tomorrow.Year(), tomorrow.Month(), tomorrow.Day(),
		hour, minute, 0, 0, loc,
	)
	endTime := startTime.Add(time.Duration(schedule.DurationMinutes) * time.Minute)

	event := &discord.GuildScheduledEventCreate{
		ChannelID:          voiceChannelID,
		Name:               schedule.Name,
		Description:        schedule.Description,
		ScheduledStartTime: startTime.UTC().Format(time.RFC3339),
		ScheduledEndTime:   endTime.UTC().Format(time.RFC3339),
		EntityType:         2, // VOICE entity type
		PrivacyLevel:       2, // GUILD_ONLY
	}

	createdEvent, err := s.client.CreateScheduledEvent(ctx, s.guildID, event)
	if err != nil {
		log.Printf("failed to create scheduled event for %s: %v", schedule.Name, err)
		return
	}

	log.Printf("created scheduled event for: %s (starts %s)", schedule.Name, startTime.Format(time.RFC3339))

	// Send notification to the events channel
	notification := fmt.Sprintf(`🎉 Hello @everyone
**New Event Alert!**

📌 **%s**
%s

🗓️ **When:** <t:%d:F> (<t:%d:R>)
🌐 **Timezone:** %s
⏱️ **Duration:** %d minutes
📍 **Where:** <#%s>

See you there! 👋
https://discord.com/events/%s/%s`,
		schedule.Name,
		schedule.Description,
		startTime.Unix(),
		startTime.Unix(),
		schedule.Timezone,
		schedule.DurationMinutes,
		voiceChannelID,
		s.guildID,
		createdEvent.ID,
	)

	if _, err := s.client.SendMessage(ctx, notifyChannelID, notification); err != nil {
		log.Printf("failed to send event notification for %s: %v", schedule.Name, err)
	} else {
		log.Printf("sent event notification for: %s", schedule.Name)
	}
}

func (s *Scheduler) resolveChannelID(ctx context.Context, channelName string) (string, error) {
	s.mu.RLock()
	if id, ok := s.channelCache[channelName]; ok {
		s.mu.RUnlock()
		return id, nil
	}
	s.mu.RUnlock()

	channels, err := s.client.GetGuildChannels(ctx, s.guildID)
	if err != nil {
		return "", fmt.Errorf("get guild channels: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ch := range channels {
		s.channelCache[ch.Name] = ch.ID
	}

	if id, ok := s.channelCache[channelName]; ok {
		return id, nil
	}

	return "", fmt.Errorf("channel not found: %s", channelName)
}
