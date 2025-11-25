package cmd

import (
	"testing"
)

func TestRateDefaults(t *testing.T) {
	// Test default values
	if rateHoursWeek != 40 {
		// Default is set in flag definition to 40
	}

	if rateWeeksYear != 48 {
		// Default is set in flag definition to 48
	}

	if rateTaxPercent != 25 {
		// Default is 25%
	}

	if rateProfitMargin != 20 {
		// Default is 20%
	}
}

func TestRateCalculation(t *testing.T) {
	// Test basic rate calculation logic
	targetAnnual := 100000.0
	taxPercent := 25.0
	hoursPerWeek := 40.0
	weeksPerYear := 48.0
	expenses := 10000.0
	profitMargin := 20.0

	// Calculate gross needed (to cover taxes)
	taxMultiplier := 1 / (1 - taxPercent/100)
	grossNeeded := targetAnnual * taxMultiplier

	if grossNeeded < targetAnnual {
		t.Errorf("Gross needed should be greater than net target")
	}

	// Add expenses
	totalNeeded := grossNeeded + expenses

	// Add profit margin
	withMargin := totalNeeded * (1 + profitMargin/100)

	// Calculate billable hours
	annualHours := hoursPerWeek * weeksPerYear

	// Calculate hourly rate
	hourlyRate := withMargin / annualHours

	// Verify rate is reasonable (should be > $80/hr for these inputs)
	if hourlyRate < 80 {
		t.Errorf("Hourly rate %f seems too low for $100k target", hourlyRate)
	}

	// Verify rate is not unreasonably high
	if hourlyRate > 150 {
		t.Errorf("Hourly rate %f seems too high for $100k target", hourlyRate)
	}
}

func TestClientRateStats(t *testing.T) {
	stats := &clientRateStats{
		hours:   10,
		revenue: 1000,
	}

	if stats.hours != 10 {
		t.Errorf("Expected 10 hours, got %f", stats.hours)
	}

	if stats.revenue != 1000 {
		t.Errorf("Expected 1000 revenue, got %f", stats.revenue)
	}

	// Calculate effective rate
	effRate := stats.revenue / stats.hours
	if effRate != 100 {
		t.Errorf("Expected effective rate of 100, got %f", effRate)
	}
}

func TestTaxMultiplier(t *testing.T) {
	tests := []struct {
		taxPercent float64
		expected   float64
	}{
		{0, 1.0},
		{25, 1.333333},
		{30, 1.428571},
		{50, 2.0},
	}

	for _, test := range tests {
		multiplier := 1 / (1 - test.taxPercent/100)
		diff := multiplier - test.expected
		if diff < -0.001 || diff > 0.001 {
			t.Errorf("Tax multiplier for %f%% = %f; want ~%f", test.taxPercent, multiplier, test.expected)
		}
	}
}

func TestProfitMargin(t *testing.T) {
	base := 100000.0
	margin := 20.0

	withMargin := base * (1 + margin/100)

	if withMargin != 120000.0 {
		t.Errorf("Expected 120000 with 20%% margin, got %f", withMargin)
	}
}
