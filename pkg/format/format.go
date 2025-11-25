package format

import (
	"fmt"
	"math"
)

// RoundHoursUp rounds hours up to the nearest integer
// Examples: 9.5 → 10, 8.1 → 9, 10.0 → 10
func RoundHoursUp(hours float64) float64 {
	return math.Ceil(hours)
}

// FormatHours formats hours with rounding up
func FormatHours(hours float64) string {
	rounded := RoundHoursUp(hours)
	return fmt.Sprintf("%.0f", rounded)
}

// RoundMoneyUp rounds money up to nearest dollar
func RoundMoneyUp(amount float64) float64 {
	return math.Ceil(amount)
}

// FormatMoney formats money with optional rounding
func FormatMoney(amount float64, roundUp bool) string {
	if roundUp {
		return fmt.Sprintf("$%.0f", RoundMoneyUp(amount))
	}
	return fmt.Sprintf("$%.2f", amount)
}
