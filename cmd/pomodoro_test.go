package cmd

import (
	"testing"
)

func TestPomodoroDefaults(t *testing.T) {
	// Test default values
	if pomodoroWorkMinutes != 25 {
		t.Errorf("Default work minutes should be 25, got %d", pomodoroWorkMinutes)
	}

	if pomodoroBreakMinutes != 5 {
		t.Errorf("Default break minutes should be 5, got %d", pomodoroBreakMinutes)
	}

	if pomodoroLongBreak != 15 {
		t.Errorf("Default long break should be 15, got %d", pomodoroLongBreak)
	}

	if pomodoroSessions != 4 {
		t.Errorf("Default sessions should be 4, got %d", pomodoroSessions)
	}
}

func TestPlayNotification(t *testing.T) {
	// Should not panic
	playNotification()
}
