package cmd

import (
	"testing"
)

func TestInitialProfitModel(t *testing.T) {
	model := initialProfitModel()

	if len(model.tabs) != 4 {
		t.Errorf("Expected 4 tabs, got %d", len(model.tabs))
	}

	if model.selectedTab != 0 {
		t.Errorf("Expected initial tab 0, got %d", model.selectedTab)
	}

	if !model.loading {
		t.Errorf("Expected initial loading state to be true")
	}

	// Check tab names
	expectedTabs := []string{"Overview", "Clients", "Expenses", "Trends"}
	for i, expected := range expectedTabs {
		if model.tabs[i] != expected {
			t.Errorf("Tab %d: expected %s, got %s", i, expected, model.tabs[i])
		}
	}
}

func TestDashboardData(t *testing.T) {
	data := &dashboardData{
		currentRevenue:  5000,
		currentExpenses: 2000,
		currentProfit:   3000,
		prevRevenue:     4000,
		prevExpenses:    1500,
		prevProfit:      2500,
		yearRevenue:     50000,
		yearExpenses:    20000,
		yearProfit:      30000,
		expensesByType:  make(map[string]float64),
	}

	// Verify calculations
	if data.currentProfit != data.currentRevenue-data.currentExpenses {
		t.Errorf("Current profit calculation incorrect")
	}

	if data.yearProfit != data.yearRevenue-data.yearExpenses {
		t.Errorf("Year profit calculation incorrect")
	}

	// Test profit margin
	margin := (data.yearProfit / data.yearRevenue) * 100
	if margin != 60 {
		t.Errorf("Expected 60%% margin, got %f%%", margin)
	}
}

func TestMonthData(t *testing.T) {
	md := monthData{
		month:    "Jan",
		revenue:  10000,
		expenses: 4000,
		profit:   6000,
	}

	if md.profit != md.revenue-md.expenses {
		t.Errorf("Month profit should equal revenue - expenses")
	}
}

func TestClientRevenue(t *testing.T) {
	client := clientRevenue{
		name:    "Test Client",
		revenue: 5000,
	}

	if client.name != "Test Client" {
		t.Errorf("Expected name 'Test Client', got %s", client.name)
	}

	if client.revenue != 5000 {
		t.Errorf("Expected revenue 5000, got %f", client.revenue)
	}
}

func TestAbsFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{5.0, 5.0},
		{-5.0, 5.0},
		{0, 0},
		{-0.5, 0.5},
	}

	for _, test := range tests {
		result := absFloat(test.input)
		if result != test.expected {
			t.Errorf("absFloat(%f) = %f; want %f", test.input, result, test.expected)
		}
	}
}

func TestProfitModelRenderProgressBar(t *testing.T) {
	model := initialProfitModel()

	tests := []struct {
		percent float64
		width   int
	}{
		{0, 10},
		{50, 10},
		{100, 10},
		{150, 10}, // over 100%
	}

	for _, test := range tests {
		bar := model.renderProgressBar(test.percent, test.width)

		// Bar should have correct length (width + 2 for borders)
		expectedLen := test.width + 2
		actualLen := len([]rune(bar))
		if actualLen != expectedLen {
			t.Errorf("Progress bar width for %f%% = %d; want %d", test.percent, actualLen, expectedLen)
		}
	}
}

func TestGoalProgress(t *testing.T) {
	data := &dashboardData{
		currentRevenue: 3000,
		monthlyGoal:    5000,
	}

	progress := (data.currentRevenue / data.monthlyGoal) * 100
	if progress != 60 {
		t.Errorf("Expected 60%% progress, got %f%%", progress)
	}
}
