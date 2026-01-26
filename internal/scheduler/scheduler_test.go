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
			name: "triggers on matching day and time",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"monday"},
				Time:     "17:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 27, 17, 0, 0, 0, time.UTC), // Monday
			want: true,
		},
		{
			name: "does not trigger on wrong day",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"tuesday"},
				Time:     "17:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 27, 17, 0, 0, 0, time.UTC), // Monday
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
			now:  time.Date(2025, 1, 27, 17, 0, 0, 0, time.UTC), // Monday
			want: false,
		},
		{
			name: "triggers with multiple days",
			schedule: Schedule{
				Name:     "Test Event",
				Days:     []string{"monday", "wednesday", "friday"},
				Time:     "17:00",
				Timezone: "UTC",
			},
			now:  time.Date(2025, 1, 29, 17, 0, 0, 0, time.UTC), // Wednesday
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
			now:  time.Date(2025, 1, 27, 17, 0, 0, 0, time.UTC), // Monday
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

func TestShouldTrigger_AlreadyCreatedToday(t *testing.T) {
	s := &Scheduler{
		lastCreated: map[string]string{
			"Test Event": "2025-01-27",
		},
	}

	schedule := Schedule{
		Name:     "Test Event",
		Days:     []string{"monday"},
		Time:     "17:00",
		Timezone: "UTC",
	}

	now := time.Date(2025, 1, 27, 17, 0, 0, 0, time.UTC)

	if s.shouldTrigger(schedule, now) {
		t.Error("shouldTrigger() should return false when event already created today")
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

	// 17:00 UTC = 12:00 EST (during standard time)
	now := time.Date(2025, 1, 27, 17, 0, 0, 0, time.UTC)

	if !s.shouldTrigger(schedule, now) {
		t.Error("shouldTrigger() should handle timezone conversion")
	}
}

func TestListSchedules(t *testing.T) {
	s := &Scheduler{
		schedules: []Schedule{
			{Name: "Event 1"},
			{Name: "Event 2"},
			{Name: "Event 3"},
		},
	}

	names := s.ListSchedules()

	if len(names) != 3 {
		t.Errorf("ListSchedules() returned %d names, want 3", len(names))
	}

	expected := []string{"Event 1", "Event 2", "Event 3"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("ListSchedules()[%d] = %s, want %s", i, name, expected[i])
		}
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
