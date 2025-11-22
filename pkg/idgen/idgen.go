package idgen

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// sanitizeName converts a name to a URL-safe format
// Removes special characters, converts to lowercase, replaces spaces with underscores
func sanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces and dashes with underscores
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")

	// Remove everything except letters, numbers, and underscores
	reg := regexp.MustCompile("[^a-z0-9_]+")
	name = reg.ReplaceAllString(name, "")

	// Remove consecutive underscores
	reg = regexp.MustCompile("_+")
	name = reg.ReplaceAllString(name, "_")

	// Trim underscores from start and end
	name = strings.Trim(name, "_")

	// Limit length to 50 characters
	if len(name) > 50 {
		name = name[:50]
	}

	return name
}

// GenerateInvoiceNumber creates a human-readable invoice number
// Format: inv.clientname.YYYY-MM-DD (with _2, _3, etc. if duplicate)
func GenerateInvoiceNumber(db *gorm.DB, clientName string, issuedDate time.Time) (string, error) {
	sanitizedClient := sanitizeName(clientName)
	dateStr := issuedDate.Format("2006-01-02")

	baseNum := fmt.Sprintf("inv.%s.%s", sanitizedClient, dateStr)

	// Check if this number exists
	var count int64
	err := db.Table("invoices").Where("invoice_num = ?", baseNum).Count(&count).Error
	if err != nil {
		return "", fmt.Errorf("failed to check for duplicate invoice number: %w", err)
	}

	if count == 0 {
		return baseNum, nil
	}

	// If exists, try with _2, _3, etc.
	for i := 2; i < 100; i++ {
		numberedInvoice := fmt.Sprintf("%s_%d", baseNum, i)
		err := db.Table("invoices").Where("invoice_num = ?", numberedInvoice).Count(&count).Error
		if err != nil {
			return "", fmt.Errorf("failed to check for duplicate invoice number: %w", err)
		}

		if count == 0 {
			return numberedInvoice, nil
		}
	}

	// Fallback to timestamp if we somehow have 100+ duplicates
	return fmt.Sprintf("%s_%d", baseNum, time.Now().Unix()), nil
}

// GenerateContractNumber creates a human-readable contract number
// Format: contract.clientname.month.year (with _2, _3, etc. if duplicate)
func GenerateContractNumber(db *gorm.DB, clientName string, startDate time.Time) (string, error) {
	sanitizedClient := sanitizeName(clientName)
	month := strings.ToLower(startDate.Format("jan")) // jan, feb, mar, etc.
	year := startDate.Format("2006")

	baseNum := fmt.Sprintf("contract.%s.%s.%s", sanitizedClient, month, year)

	// Check if this number exists
	var count int64
	err := db.Table("contracts").Where("contract_num = ?", baseNum).Count(&count).Error
	if err != nil {
		return "", fmt.Errorf("failed to check for duplicate contract number: %w", err)
	}

	if count == 0 {
		return baseNum, nil
	}

	// If exists, try with _2, _3, etc.
	for i := 2; i < 100; i++ {
		numberedContract := fmt.Sprintf("%s_%d", baseNum, i)
		err := db.Table("contracts").Where("contract_num = ?", numberedContract).Count(&count).Error
		if err != nil {
			return "", fmt.Errorf("failed to check for duplicate contract number: %w", err)
		}

		if count == 0 {
			return numberedContract, nil
		}
	}

	// Fallback to timestamp if we somehow have 100+ duplicates
	return fmt.Sprintf("%s_%d", baseNum, time.Now().Unix()), nil
}
