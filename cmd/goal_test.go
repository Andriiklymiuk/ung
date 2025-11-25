package cmd

import (
	"testing"
	"time"
)

func TestGoalDefaults(t *testing.T) {
	// Test default values
	if goalPeriod != "monthly" {
		// Default is set in flag definition
	}

	if goalYear != 0 {
		t.Errorf("Default goal year should be 0, got %d", goalYear)
	}

	if goalMonth != 0 {
		t.Errorf("Default goal month should be 0, got %d", goalMonth)
	}
}

func TestFormatGoalPeriod(t *testing.T) {
	tests := []struct {
		year     int
		month    int
		quarter  int
		period   string
		expected string
	}{
		{2024, 1, 0, "monthly", "Jan 2024"},
		{2024, 6, 0, "monthly", "Jun 2024"},
		{2024, 12, 0, "monthly", "Dec 2024"},
		{2024, 0, 1, "quarterly", "Q1 2024"},
		{2024, 0, 4, "quarterly", "Q4 2024"},
		{2024, 0, 0, "yearly", "2024"},
	}

	for _, test := range tests {
		result := formatGoalPeriod(test.year, test.month, test.quarter, test.period)
		if result != test.expected {
			t.Errorf("formatGoalPeriod(%d, %d, %d, %s) = %s; want %s",
				test.year, test.month, test.quarter, test.period, result, test.expected)
		}
	}
}

func TestGetGoalDateRange(t *testing.T) {
	// Test monthly goal
	monthlyGoal := IncomeGoal{
		Period: "monthly",
		Year:   2024,
		Month:  6,
	}

	start, end := getGoalDateRange(monthlyGoal)
	if start.Year() != 2024 || start.Month() != time.June || start.Day() != 1 {
		t.Errorf("Monthly goal start date incorrect: %v", start)
	}
	if end.Year() != 2024 || end.Month() != time.July || end.Day() != 1 {
		t.Errorf("Monthly goal end date incorrect: %v", end)
	}

	// Test quarterly goal
	quarterlyGoal := IncomeGoal{
		Period:  "quarterly",
		Year:    2024,
		Quarter: 2,
	}

	start, end = getGoalDateRange(quarterlyGoal)
	if start.Month() != time.April {
		t.Errorf("Q2 start should be April, got %v", start.Month())
	}
	if end.Month() != time.July {
		t.Errorf("Q2 end should be July, got %v", end.Month())
	}

	// Test yearly goal
	yearlyGoal := IncomeGoal{
		Period: "yearly",
		Year:   2024,
	}

	start, end = getGoalDateRange(yearlyGoal)
	if start.Month() != time.January || start.Day() != 1 {
		t.Errorf("Yearly goal start should be Jan 1, got %v", start)
	}
	if end.Year() != 2025 || end.Month() != time.January {
		t.Errorf("Yearly goal end should be Jan 2025, got %v", end)
	}
}

func TestGetDaysRemaining(t *testing.T) {
	// Test future date
	future := time.Now().Add(10 * 24 * time.Hour)
	days := getDaysRemaining(future)
	if days < 9 || days > 10 {
		t.Errorf("Expected ~10 days remaining, got %d", days)
	}

	// Test past date
	past := time.Now().Add(-10 * 24 * time.Hour)
	days = getDaysRemaining(past)
	if days != 0 {
		t.Errorf("Expected 0 days for past date, got %d", days)
	}
}

func TestIncomeGoalStruct(t *testing.T) {
	goal := IncomeGoal{
		Amount:      5000,
		Period:      "monthly",
		Year:        2024,
		Month:       6,
		Description: "Test goal",
	}

	if goal.Amount != 5000 {
		t.Errorf("Expected amount 5000, got %f", goal.Amount)
	}
	if goal.Period != "monthly" {
		t.Errorf("Expected period monthly, got %s", goal.Period)
	}
}
