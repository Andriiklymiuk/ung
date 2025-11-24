package db

import (
	"os"
	"testing"
)

func TestPasswordManagement(t *testing.T) {
	// Clean up any cached password
	ClearPasswordCache()

	t.Run("Set and get password", func(t *testing.T) {
		testPassword := "TestPassword123!"
		SetDatabasePassword(testPassword)

		retrieved, err := GetDatabasePassword()
		if err != nil {
			t.Fatalf("Failed to get password: %v", err)
		}

		if retrieved != testPassword {
			t.Fatalf("Expected password %s, got %s", testPassword, retrieved)
		}
	})

	t.Run("Clear password cache", func(t *testing.T) {
		SetDatabasePassword("SomePassword123!")
		ClearPasswordCache()

		// After clearing, should try to get from env or prompt
		// Set env var for testing
		os.Setenv("UNG_DB_PASSWORD", "EnvPassword123!")
		defer os.Unsetenv("UNG_DB_PASSWORD")

		retrieved, err := GetDatabasePassword()
		if err != nil {
			t.Fatalf("Failed to get password from env: %v", err)
		}

		if retrieved != "EnvPassword123!" {
			t.Fatalf("Expected env password, got %s", retrieved)
		}
	})

	t.Run("Get password from environment", func(t *testing.T) {
		ClearPasswordCache()

		envPassword := "EnvironmentPassword123!"
		os.Setenv("UNG_DB_PASSWORD", envPassword)
		defer os.Unsetenv("UNG_DB_PASSWORD")

		retrieved, err := GetDatabasePassword()
		if err != nil {
			t.Fatalf("Failed to get password from env: %v", err)
		}

		if retrieved != envPassword {
			t.Fatalf("Expected %s, got %s", envPassword, retrieved)
		}

		// Should be cached now
		os.Unsetenv("UNG_DB_PASSWORD")

		cached, err := GetDatabasePassword()
		if err != nil {
			t.Fatalf("Failed to get cached password: %v", err)
		}

		if cached != envPassword {
			t.Fatal("Password should be cached after first retrieval")
		}
	})

	t.Run("Password persistence in session", func(t *testing.T) {
		ClearPasswordCache()

		password1 := "FirstPassword123!"
		SetDatabasePassword(password1)

		// Multiple calls should return same password
		for i := 0; i < 5; i++ {
			retrieved, err := GetDatabasePassword()
			if err != nil {
				t.Fatalf("Failed to get password on iteration %d: %v", i, err)
			}

			if retrieved != password1 {
				t.Fatalf("Password mismatch on iteration %d", i)
			}
		}
	})

	// Clean up
	ClearPasswordCache()
	os.Unsetenv("UNG_DB_PASSWORD")
}
