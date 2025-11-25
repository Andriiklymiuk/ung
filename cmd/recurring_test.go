package cmd

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/models"
)

func TestCalculateNextGenerationDate(t *testing.T) {
	tests := []struct {
		name       string
		frequency  models.RecurringFrequency
		dayOfMonth int
		dayOfWeek  int
		wantAfter  time.Time
	}{
		{
			name:       "monthly on 1st",
			frequency:  models.FrequencyMonthly,
			dayOfMonth: 1,
			wantAfter:  time.Now(),
		},
		{
			name:       "monthly on 15th",
			frequency:  models.FrequencyMonthly,
			dayOfMonth: 15,
			wantAfter:  time.Now(),
		},
		{
			name:       "quarterly",
			frequency:  models.FrequencyQuarterly,
			dayOfMonth: 1,
			wantAfter:  time.Now().AddDate(0, 2, 0),
		},
		{
			name:       "yearly",
			frequency:  models.FrequencyYearly,
			dayOfMonth: 1,
			wantAfter:  time.Now().AddDate(0, 11, 0),
		},
		{
			name:      "weekly",
			frequency: models.FrequencyWeekly,
			dayOfWeek: 1, // Monday
			wantAfter: time.Now(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateNextGenerationDate(tt.frequency, tt.dayOfMonth, tt.dayOfWeek)

			// Check that the date is in the future
			if !got.After(time.Now().Add(-24 * time.Hour)) {
				t.Errorf("calculateNextGenerationDate() = %v, want after %v", got, time.Now())
			}

			// Check that the date is after expected minimum
			if !got.After(tt.wantAfter.Add(-24 * time.Hour)) {
				t.Errorf("calculateNextGenerationDate() = %v, want after %v", got, tt.wantAfter)
			}
		})
	}
}

func TestRecurringFrequencyConstants(t *testing.T) {
	// Verify all frequency constants are defined
	frequencies := []models.RecurringFrequency{
		models.FrequencyWeekly,
		models.FrequencyBiweekly,
		models.FrequencyMonthly,
		models.FrequencyQuarterly,
		models.FrequencyYearly,
	}

	for _, f := range frequencies {
		if f == "" {
			t.Errorf("Frequency constant is empty")
		}
	}
}

func TestCalculateNextGenerationDateMonthlyDayOfMonth(t *testing.T) {
	now := time.Now()

	// Test that day 15 returns a date with day 15
	result := calculateNextGenerationDate(models.FrequencyMonthly, 15, 0)
	if result.Day() != 15 {
		t.Errorf("Expected day 15, got day %d", result.Day())
	}

	// Test that the month is in the future
	if result.Month() <= now.Month() && result.Year() == now.Year() {
		t.Errorf("Expected future month, got %v", result)
	}
}

func TestCalculateNextGenerationDateHandlesEndOfMonth(t *testing.T) {
	// Test that day 31 gets normalized to 28
	result := calculateNextGenerationDate(models.FrequencyMonthly, 31, 0)
	if result.Day() > 28 {
		t.Errorf("Expected day <= 28 for safety, got day %d", result.Day())
	}
}
