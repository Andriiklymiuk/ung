package cmd

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/models"
)

func TestImportDefaults(t *testing.T) {
	// Test default values
	if importFile != "" {
		t.Errorf("Default import file should be empty, got %s", importFile)
	}

	if importType != "" {
		t.Errorf("Default import type should be empty, got %s", importType)
	}

	if importDryRun {
		t.Errorf("Default importDryRun should be false")
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"2024-01-15", "2024-01-15", false},
		{"01/15/2024", "2024-01-15", false},
		{"1/5/2024", "2024-01-05", false},
		{"invalid", "", true},
	}

	for _, test := range tests {
		result, err := parseDate(test.input)
		if test.hasError && err == nil {
			t.Errorf("Expected error for input %s", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for input %s: %v", test.input, err)
		}
		if !test.hasError && result.Format("2006-01-02") != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result.Format("2006-01-02"))
		}
	}
}

func TestParseCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected models.ExpenseCategory
	}{
		{"software", models.ExpenseCategorySoftware},
		{"Software Subscription", models.ExpenseCategorySoftware},
		{"hardware", models.ExpenseCategoryHardware},
		{"travel", models.ExpenseCategoryTravel},
		{"meals", models.ExpenseCategoryMeals},
		{"office supplies", models.ExpenseCategoryOfficeSupplies},
		{"utilities", models.ExpenseCategoryUtilities},
		{"marketing", models.ExpenseCategoryMarketing},
		{"random stuff", models.ExpenseCategoryOther},
	}

	for _, test := range tests {
		result := parseCategory(test.input)
		if result != test.expected {
			t.Errorf("For input %s, expected %s, got %s", test.input, test.expected, result)
		}
	}
}

func TestSkipRowsDefault(t *testing.T) {
	// Default should be 1 (skip header row)
	if importSkipRows != 1 {
		t.Errorf("Default skip rows should be 1, got %d", importSkipRows)
	}
}

func TestParseDateFormats(t *testing.T) {
	// Test various date formats that should be supported
	validDates := []string{
		"2024-01-15",
		"01/15/2024",
		"2024/01/15",
	}

	for _, d := range validDates {
		_, err := parseDate(d)
		if err != nil {
			t.Errorf("Should be able to parse: %s", d)
		}
	}
}

func TestParseDateReturnsCorrectTime(t *testing.T) {
	result, err := parseDate("2024-06-15")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Year() != 2024 {
		t.Errorf("Expected year 2024, got %d", result.Year())
	}
	if result.Month() != time.June {
		t.Errorf("Expected month June, got %v", result.Month())
	}
	if result.Day() != 15 {
		t.Errorf("Expected day 15, got %d", result.Day())
	}
}
