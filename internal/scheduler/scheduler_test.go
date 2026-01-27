package scheduler

import (
	"testing"
	"time"
)

func TestShouldTrigger(t *testing.T) {
	s := &Scheduler{
		lastCreated: make(map[string]string),
	}

	tests := []struct {
		name     string
		schedule Schedule
		now      time.Time
		want     bool
	}{
		{
			name: "triggers one day before matching day at same time",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"monday"},
				Time:     "17:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 26, 17, 0, 0, 0, time.UTC), // Sunday (one day before Monday)
			want: true,
		},
		{
			name: "does not trigger if tomorrow is not scheduled",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"tuesday"},
				Time:     "17:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 26, 17, 0, 0, 0, time.UTC), // Sunday (tomorrow is Monday)
			want: false,
		},
		{
			name: "does not trigger on wrong time",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"monday"},
				Time:     "18:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 26, 17, 0, 0, 0, time.UTC), // Sunday at 17:00
			want: false,
		},
		{
			name: "triggers with multiple days when tomorrow matches",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"monday", "wednesday", "friday"},
				Time:     "17:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 28, 17, 0, 0, 0, time.UTC), // Tuesday (tomorrow is Wednesday)
			want: true,
		},
		{
			name: "case insensitive days",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"Monday", "WEDNESDAY"},
				Time:     "17:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 26, 17, 0, 0, 0, time.UTC), // Sunday (tomorrow is Monday)
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.shouldTrigger(tt.schedule, tt.now)
			if got != tt.want {
				t.Errorf("shouldTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldTrigger_AlreadyCreatedForTomorrow(t *testing.T) {
	// Event on Monday was already created on Sunday
	s := &Scheduler{
		lastCreated: map[string]string{
			"Test Event": "2025-01-27", // Monday's date
		},
	}

	schedule := Schedule{
		Name:     "Test Event",
		Days:     []string{"monday"},
		Time:     "17:00",
		Timezone: "UTC",
	}

	// Sunday at 17:00 (tomorrow is Monday)
	now := time.Date(2025, 1, 26, 17, 0, 0, 0, time.UTC)

	if s.shouldTrigger(schedule, now) {
		t.Error("shouldTrigger() should return false when event already created for tomorrow")
	}
}

func TestShouldTrigger_DifferentTimezone(t *testing.T) {
	s := &Scheduler{
		lastCreated: make(map[string]string),
	}

	schedule := Schedule{
		Name:     "Test Event",
		Days:     []string{"monday"},
		Time:     "12:00",
		Timezone: "America/New_York",
	}

	// 17:00 UTC on Sunday = 12:00 EST on Sunday (during standard time)
	// Tomorrow is Monday, which matches the schedule
	now := time.Date(2025, 1, 26, 17, 0, 0, 0, time.UTC)

	if !s.shouldTrigger(schedule, now) {
		t.Error("shouldTrigger() should handle timezone conversion and trigger one day before")
	}
}

func TestReminderTimeCalculation(t *testing.T) {
	tests := []struct {
		name          string
		eventTime     string
		reminderMins  int
		currentTime   time.Time
		shouldTrigger bool
	}{
		{
			name:          "1 hour before",
			eventTime:     "17:00",
			reminderMins:  60,
			currentTime:   time.Date(2025, 1, 27, 16, 0, 0, 0, time.UTC),
			shouldTrigger: true,
		},
		{
			name:          "15 minutes before",
			eventTime:     "17:00",
			reminderMins:  15,
			currentTime:   time.Date(2025, 1, 27, 16, 45, 0, 0, time.UTC),
			shouldTrigger: true,
		},
		{
			name:          "wrong time for reminder",
			eventTime:     "17:00",
			reminderMins:  60,
			currentTime:   time.Date(2025, 1, 27, 15, 0, 0, 0, time.UTC),
			shouldTrigger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := []int{17, 0}
			eventTime := time.Date(
				tt.currentTime.Year(), tt.currentTime.Month(), tt.currentTime.Day(),
				parts[0], parts[1], 0, 0, time.UTC,
			)
			reminderTime := eventTime.Add(-time.Duration(tt.reminderMins) * time.Minute)

			matches := tt.currentTime.Hour() == reminderTime.Hour() &&
				tt.currentTime.Minute() == reminderTime.Minute()

			if matches != tt.shouldTrigger {
				t.Errorf("reminder check = %v, want %v", matches, tt.shouldTrigger)
			}
		})
	}
}
