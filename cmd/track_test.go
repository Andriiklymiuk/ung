package cmd

import (
	"testing"
	"time"

	"github.com/Andriiklymiuk/ung/internal/db"
)

func TestTrackingSessionCreate(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Create a tracking session
	result, err := db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, billable, notes)
		VALUES ('Test Project', ?, 1, 'Test notes')
	`, time.Now())
	if err != nil {
		t.Fatalf("Failed to create tracking session: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get last insert ID: %v", err)
	}

	if id < 1 {
		t.Errorf("Expected session ID >= 1, got %d", id)
	}

	// Verify the session was created
	var projectName string
	var billable bool
	err = db.DB.QueryRow("SELECT project_name, billable FROM tracking_sessions WHERE id = ?", id).
		Scan(&projectName, &billable)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}

	if projectName != "Test Project" {
		t.Errorf("Expected project name 'Test Project', got '%s'", projectName)
	}
	if !billable {
		t.Errorf("Expected billable to be true")
	}
}

func TestTrackingSessionStartStop(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	startTime := time.Now().Add(-1 * time.Hour)

	// Start a session
	result, err := db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, billable)
		VALUES ('Active Project', ?, 1)
	`, startTime)
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}

	id, _ := result.LastInsertId()

	// Verify session is active (no end_time)
	var endTime *time.Time
	err = db.DB.QueryRow("SELECT end_time FROM tracking_sessions WHERE id = ?", id).Scan(&endTime)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}
	if endTime != nil {
		t.Errorf("Expected end_time to be NULL for active session")
	}

	// Stop the session
	now := time.Now()
	duration := int(now.Sub(startTime).Seconds())
	_, err = db.DB.Exec(`
		UPDATE tracking_sessions SET end_time = ?, duration = ? WHERE id = ?
	`, now, duration, id)
	if err != nil {
		t.Fatalf("Failed to stop session: %v", err)
	}

	// Verify session is stopped
	var stoppedEndTime time.Time
	var stoppedDuration int
	err = db.DB.QueryRow("SELECT end_time, duration FROM tracking_sessions WHERE id = ?", id).
		Scan(&stoppedEndTime, &stoppedDuration)
	if err != nil {
		t.Fatalf("Failed to query stopped session: %v", err)
	}

	if stoppedEndTime.IsZero() {
		t.Errorf("Expected end_time to be set after stopping")
	}
	if stoppedDuration < 3500 || stoppedDuration > 3700 { // ~1 hour in seconds
		t.Errorf("Expected duration around 3600 seconds, got %d", stoppedDuration)
	}
}

func TestTrackingSessionWithClient(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Create a client first
	clientResult, err := db.DB.Exec(`
		INSERT INTO clients (name, email) VALUES ('Test Client', 'client@test.com')
	`)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	clientID, _ := clientResult.LastInsertId()

	// Create a tracking session with client
	result, err := db.DB.Exec(`
		INSERT INTO tracking_sessions (client_id, project_name, start_time, billable)
		VALUES (?, 'Client Project', ?, 1)
	`, clientID, time.Now())
	if err != nil {
		t.Fatalf("Failed to create session with client: %v", err)
	}

	sessionID, _ := result.LastInsertId()

	// Verify the session has the client
	var storedClientID int
	err = db.DB.QueryRow("SELECT client_id FROM tracking_sessions WHERE id = ?", sessionID).
		Scan(&storedClientID)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}

	if int64(storedClientID) != clientID {
		t.Errorf("Expected client ID %d, got %d", clientID, storedClientID)
	}
}

func TestTrackingSessionSoftDelete(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Create a session
	result, err := db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, billable)
		VALUES ('To Delete', ?, 1)
	`, time.Now())
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	id, _ := result.LastInsertId()

	// Soft delete
	_, err = db.DB.Exec("UPDATE tracking_sessions SET deleted_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		t.Fatalf("Failed to soft delete: %v", err)
	}

	// Verify session is soft-deleted (deleted_at is set)
	var deletedAt *time.Time
	err = db.DB.QueryRow("SELECT deleted_at FROM tracking_sessions WHERE id = ?", id).Scan(&deletedAt)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}

	if deletedAt == nil {
		t.Errorf("Expected deleted_at to be set after soft delete")
	}

	// Verify it doesn't appear in non-deleted queries
	var count int
	err = db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE deleted_at IS NULL").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count non-deleted sessions: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 non-deleted sessions, got %d", count)
	}
}

func TestTrackingSessionEdit(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Create a session
	result, err := db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, duration, hours, billable, notes)
		VALUES ('Original Project', ?, 3600, 1.0, 1, 'Original notes')
	`, time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	id, _ := result.LastInsertId()

	// Edit the session
	_, err = db.DB.Exec(`
		UPDATE tracking_sessions SET project_name = ?, hours = ?, duration = ?, notes = ? WHERE id = ?
	`, "Updated Project", 2.5, 9000, "Updated notes", id)
	if err != nil {
		t.Fatalf("Failed to edit session: %v", err)
	}

	// Verify edits
	var projectName, notes string
	var hours float64
	var duration int
	err = db.DB.QueryRow("SELECT project_name, hours, duration, notes FROM tracking_sessions WHERE id = ?", id).
		Scan(&projectName, &hours, &duration, &notes)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}

	if projectName != "Updated Project" {
		t.Errorf("Expected project name 'Updated Project', got '%s'", projectName)
	}
	if hours != 2.5 {
		t.Errorf("Expected hours 2.5, got %f", hours)
	}
	if duration != 9000 {
		t.Errorf("Expected duration 9000, got %d", duration)
	}
	if notes != "Updated notes" {
		t.Errorf("Expected notes 'Updated notes', got '%s'", notes)
	}
}

func TestTrackingSessionUnbilledFilter(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Create billable session (unbilled)
	_, err := db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, end_time, duration, billable)
		VALUES ('Unbilled Task', ?, ?, 3600, 1)
	`, time.Now().Add(-2*time.Hour), time.Now().Add(-1*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create unbilled session: %v", err)
	}

	// Create non-billable session
	_, err = db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, end_time, duration, billable)
		VALUES ('Non-billable Task', ?, ?, 1800, 0)
	`, time.Now().Add(-3*time.Hour), time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create non-billable session: %v", err)
	}

	// Create billed session (has [Invoiced: marker in notes)
	_, err = db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, end_time, duration, billable, notes)
		VALUES ('Billed Task', ?, ?, 7200, 1, '[Invoiced: INV-001]')
	`, time.Now().Add(-4*time.Hour), time.Now().Add(-3*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create billed session: %v", err)
	}

	// Query unbilled sessions only
	var unbilledCount int
	err = db.DB.QueryRow(`
		SELECT COUNT(*) FROM tracking_sessions
		WHERE billable = 1
		AND deleted_at IS NULL
		AND (notes NOT LIKE '%[Invoiced:%' OR notes IS NULL OR notes = '')
	`).Scan(&unbilledCount)
	if err != nil {
		t.Fatalf("Failed to count unbilled sessions: %v", err)
	}

	if unbilledCount != 1 {
		t.Errorf("Expected 1 unbilled session, got %d", unbilledCount)
	}
}

func TestTrackingSessionLogManualHours(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Create client first
	clientResult, err := db.DB.Exec("INSERT INTO clients (name, email) VALUES ('Manual Log Client', 'manual@test.com')")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	clientID, _ := clientResult.LastInsertId()

	// Create contract with all required fields
	contractResult, err := db.DB.Exec(`
		INSERT INTO contracts (client_id, name, contract_type, contract_num, hourly_rate, currency, active, start_date)
		VALUES (?, 'Manual Contract', 'hourly', 'C-001', 50.0, 'USD', 1, ?)
	`, clientID, time.Now())
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}
	contractID, _ := contractResult.LastInsertId()

	// Log manual hours
	hours := 2.5
	durationSecs := int(hours * 3600)
	now := time.Now()

	sessionResult, err := db.DB.Exec(`
		INSERT INTO tracking_sessions
		(client_id, contract_id, project_name, start_time, end_time, duration, hours, billable, notes)
		VALUES (?, ?, 'Manual Task', ?, ?, ?, ?, 1, 'Logged manually')
	`, clientID, contractID, now, now.Add(time.Duration(durationSecs)*time.Second), durationSecs, hours)
	if err != nil {
		t.Fatalf("Failed to log manual hours: %v", err)
	}
	sessionID, _ := sessionResult.LastInsertId()

	// Verify
	var storedHours float64
	var storedDuration int
	err = db.DB.QueryRow("SELECT hours, duration FROM tracking_sessions WHERE id = ?", sessionID).
		Scan(&storedHours, &storedDuration)
	if err != nil {
		t.Fatalf("Failed to query session: %v", err)
	}

	if storedHours != hours {
		t.Errorf("Expected hours %f, got %f", hours, storedHours)
	}
	if storedDuration != durationSecs {
		t.Errorf("Expected duration %d, got %d", durationSecs, storedDuration)
	}
}

func TestTrackingActiveSessionCount(t *testing.T) {
	setupTestDB(t)
	defer db.Close()

	// Initially no active sessions
	var count int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE end_time IS NULL").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count active sessions: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 active sessions initially, got %d", count)
	}

	// Start an active session
	_, err = db.DB.Exec(`
		INSERT INTO tracking_sessions (project_name, start_time, billable)
		VALUES ('Active Session', ?, 1)
	`, time.Now())
	if err != nil {
		t.Fatalf("Failed to create active session: %v", err)
	}

	// Check active count
	err = db.DB.QueryRow("SELECT COUNT(*) FROM tracking_sessions WHERE end_time IS NULL").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count active sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 active session, got %d", count)
	}
}
