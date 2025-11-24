package services

import (
	"sync"
	"testing"
	"time"
	"ung-telegram/internal/models"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()

	if sm == nil {
		t.Fatal("NewSessionManager returned nil")
	}

	if sm.sessions == nil {
		t.Error("sessions map not initialized")
	}

	if sm.users == nil {
		t.Error("users map not initialized")
	}
}

func TestSetAndGetSession(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(123456)

	session := &models.Session{
		TelegramID: telegramID,
		State:      string(models.StateInvoiceSelectClient),
		Data:       make(map[string]interface{}),
	}

	sm.SetSession(session)

	retrieved := sm.GetSession(telegramID)
	if retrieved == nil {
		t.Fatal("GetSession returned nil")
	}

	if retrieved.TelegramID != telegramID {
		t.Errorf("Expected TelegramID %d, got %d", telegramID, retrieved.TelegramID)
	}

	if retrieved.State != string(models.StateInvoiceSelectClient) {
		t.Errorf("Expected State %s, got %s", models.StateInvoiceSelectClient, retrieved.State)
	}
}

func TestGetSession_NotExists(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(999999)

	session := sm.GetSession(telegramID)
	if session != nil {
		t.Error("Expected nil for non-existent session")
	}
}

func TestClearSession(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(123456)

	session := &models.Session{
		TelegramID: telegramID,
		State:      string(models.StateInvoiceAmount),
		Data:       make(map[string]interface{}),
	}

	sm.SetSession(session)

	// Verify session exists
	if sm.GetSession(telegramID) == nil {
		t.Fatal("Session should exist before clearing")
	}

	sm.ClearSession(telegramID)

	// Verify session is cleared
	if sm.GetSession(telegramID) != nil {
		t.Error("Session should be nil after clearing")
	}
}

func TestSetAndGetUser(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(789012)

	user := &models.User{
		TelegramID: telegramID,
		Email:      "test@example.com",
		APIToken:   "test-token",
	}

	sm.SetUser(user)

	retrieved := sm.GetUser(telegramID)
	if retrieved == nil {
		t.Fatal("GetUser returned nil")
	}

	if retrieved.TelegramID != telegramID {
		t.Errorf("Expected TelegramID %d, got %d", telegramID, retrieved.TelegramID)
	}

	if retrieved.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, retrieved.Email)
	}

	if retrieved.APIToken != user.APIToken {
		t.Errorf("Expected APIToken %s, got %s", user.APIToken, retrieved.APIToken)
	}
}

func TestGetUser_NotExists(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(999999)

	user := sm.GetUser(telegramID)
	if user != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestIsAuthenticated_True(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(345678)

	user := &models.User{
		TelegramID: telegramID,
		Email:      "auth@example.com",
		APIToken:   "valid-token",
	}

	sm.SetUser(user)

	if !sm.IsAuthenticated(telegramID) {
		t.Error("User should be authenticated")
	}
}

func TestIsAuthenticated_False_NoUser(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(999999)

	if sm.IsAuthenticated(telegramID) {
		t.Error("Non-existent user should not be authenticated")
	}
}

func TestIsAuthenticated_False_EmptyToken(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(456789)

	user := &models.User{
		TelegramID: telegramID,
		Email:      "notoken@example.com",
		APIToken:   "",
	}

	sm.SetUser(user)

	if sm.IsAuthenticated(telegramID) {
		t.Error("User with empty token should not be authenticated")
	}
}

func TestSessionUpdateTime(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(567890)

	session := &models.Session{
		TelegramID: telegramID,
		State:      string(models.StateInvoiceAmount),
		Data:       make(map[string]interface{}),
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}

	oldTime := session.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	sm.SetSession(session)

	retrieved := sm.GetSession(telegramID)
	if retrieved.UpdatedAt.Before(oldTime) || retrieved.UpdatedAt.Equal(oldTime) {
		t.Error("UpdatedAt should be updated when setting session")
	}
}

func TestSessionDataStorage(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(678901)

	session := &models.Session{
		TelegramID: telegramID,
		State:      string(models.StateInvoiceDescription),
		Data: map[string]interface{}{
			"client_id": uint(42),
			"amount":    1500.50,
			"description": "Test invoice",
		},
	}

	sm.SetSession(session)

	retrieved := sm.GetSession(telegramID)
	if retrieved.Data["client_id"] != uint(42) {
		t.Error("client_id not stored correctly")
	}

	if retrieved.Data["amount"] != 1500.50 {
		t.Error("amount not stored correctly")
	}

	if retrieved.Data["description"] != "Test invoice" {
		t.Error("description not stored correctly")
	}
}

func TestConcurrentAccess(t *testing.T) {
	sm := NewSessionManager()
	telegramID := int64(111111)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			session := &models.Session{
				TelegramID: telegramID,
				State:      string(models.StateInvoiceAmount),
				Data: map[string]interface{}{
					"iteration": idx,
				},
			}
			sm.SetSession(session)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.GetSession(telegramID)
		}()
	}

	// Concurrent authentication checks
	user := &models.User{
		TelegramID: telegramID,
		Email:      "concurrent@example.com",
		APIToken:   "token",
	}
	sm.SetUser(user)

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.IsAuthenticated(telegramID)
		}()
	}

	wg.Wait()

	// Verify data integrity after concurrent operations
	session := sm.GetSession(telegramID)
	if session == nil {
		t.Fatal("Session should exist after concurrent writes")
	}

	if !sm.IsAuthenticated(telegramID) {
		t.Error("User should be authenticated after concurrent operations")
	}
}

func TestMultipleUsers(t *testing.T) {
	sm := NewSessionManager()

	users := []int64{100, 200, 300, 400, 500}

	// Create sessions for multiple users
	for _, id := range users {
		session := &models.Session{
			TelegramID: id,
			State:      string(models.StateInvoiceSelectClient),
			Data:       map[string]interface{}{"user_id": id},
		}
		sm.SetSession(session)

		user := &models.User{
			TelegramID: id,
			Email:      "user" + string(rune(id)) + "@example.com",
			APIToken:   "token-" + string(rune(id)),
		}
		sm.SetUser(user)
	}

	// Verify all sessions exist independently
	for _, id := range users {
		session := sm.GetSession(id)
		if session == nil {
			t.Errorf("Session for user %d not found", id)
			continue
		}

		if session.TelegramID != id {
			t.Errorf("Expected TelegramID %d, got %d", id, session.TelegramID)
		}

		if !sm.IsAuthenticated(id) {
			t.Errorf("User %d should be authenticated", id)
		}
	}

	// Clear one session and verify others are unaffected
	sm.ClearSession(users[2])

	if sm.GetSession(users[2]) != nil {
		t.Error("Cleared session should be nil")
	}

	if sm.GetSession(users[0]) == nil || sm.GetSession(users[1]) == nil {
		t.Error("Other sessions should remain intact")
	}
}
