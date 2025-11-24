package cmd

import (
	"fmt"

	"github.com/Andriiklymiuk/ung/internal/db"
	"github.com/charmbracelet/huh"
)

// clientMatch represents a client search result
type clientMatch struct {
	ID   uint
	Name string
}

// FindClientByName searches for clients by name (case-insensitive partial match)
// Returns the client ID and full name. If multiple matches found, prompts user to select.
func FindClientByName(searchTerm string) (uint, string, error) {
	matches, err := searchClients(searchTerm)
	if err != nil {
		return 0, "", err
	}

	if len(matches) == 0 {
		return 0, "", fmt.Errorf("client '%s' not found", searchTerm)
	}

	// Single match - return immediately
	if len(matches) == 1 {
		return matches[0].ID, matches[0].Name, nil
	}

	// Multiple matches - let user choose
	return promptClientSelection(matches, searchTerm)
}

// searchClients queries the database for clients matching the search term
func searchClients(searchTerm string) ([]clientMatch, error) {
	rows, err := db.DB.Query(`
		SELECT id, name FROM clients
		WHERE LOWER(name) LIKE LOWER(?)
		ORDER BY name
	`, "%"+searchTerm+"%")

	if err != nil {
		return nil, fmt.Errorf("failed to search for clients: %w", err)
	}
	defer rows.Close()

	var matches []clientMatch
	for rows.Next() {
		var m clientMatch
		if err := rows.Scan(&m.ID, &m.Name); err != nil {
			return nil, fmt.Errorf("failed to scan client: %w", err)
		}
		matches = append(matches, m)
	}

	return matches, nil
}

// promptClientSelection shows an interactive UI to select from multiple client matches
func promptClientSelection(matches []clientMatch, searchTerm string) (uint, string, error) {
	clientOptions := make([]huh.Option[uint], len(matches))
	for i, m := range matches {
		clientOptions[i] = huh.NewOption(m.Name, m.ID)
	}

	var selectedClientID uint
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[uint]().
				Title("Multiple clients found").
				Description(fmt.Sprintf("Which '%s' did you mean?", searchTerm)).
				Options(clientOptions...).
				Value(&selectedClientID),
		),
	)

	if err := form.Run(); err != nil {
		return 0, "", fmt.Errorf("selection cancelled: %w", err)
	}

	// Find the selected client name
	var selectedName string
	for _, m := range matches {
		if m.ID == selectedClientID {
			selectedName = m.Name
			break
		}
	}

	return selectedClientID, selectedName, nil
}

// FindActiveContractForClient finds the most recent active contract for a client
func FindActiveContractForClient(clientID uint) (int, error) {
	var contractID int
	err := db.DB.QueryRow(`
		SELECT id FROM contracts
		WHERE client_id = ? AND active = 1
		ORDER BY id DESC
		LIMIT 1
	`, clientID).Scan(&contractID)

	if err != nil {
		return 0, fmt.Errorf("no active contract found for client")
	}

	return contractID, nil
}
