package cmd

import (
	"testing"
)

func TestBumpVersion_Patch(t *testing.T) {
	testCases := []struct {
		name     string
		current  string
		expected string
	}{
		{"simple patch", "v1.2.3", "v1.2.4"},
		{"zero patch", "v2.0.0", "v2.0.1"},
		{"high numbers", "v10.20.99", "v10.20.100"},
		{"without v prefix", "1.2.3", "v1.2.4"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bumpVersion(tc.current, "patch")
			if err != nil {
				t.Fatalf("bumpVersion failed: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestBumpVersion_Minor(t *testing.T) {
	testCases := []struct {
		name     string
		current  string
		expected string
	}{
		{"simple minor", "v1.2.3", "v1.3.0"},
		{"resets patch", "v2.5.99", "v2.6.0"},
		{"zero minor", "v3.0.0", "v3.1.0"},
		{"without v prefix", "1.2.3", "v1.3.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bumpVersion(tc.current, "minor")
			if err != nil {
				t.Fatalf("bumpVersion failed: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestBumpVersion_Major(t *testing.T) {
	testCases := []struct {
		name     string
		current  string
		expected string
	}{
		{"simple major", "v1.2.3", "v2.0.0"},
		{"resets minor and patch", "v5.10.20", "v6.0.0"},
		{"from v0", "v0.9.9", "v1.0.0"},
		{"without v prefix", "1.2.3", "v2.0.0"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bumpVersion(tc.current, "major")
			if err != nil {
				t.Fatalf("bumpVersion failed: %v", err)
			}
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestBumpVersion_Invalid(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		bump    string
	}{
		{"invalid version format", "abc", "patch"},
		{"invalid bump type", "v1.2.3", "invalid"},
		{"empty version", "", "patch"},
		{"missing parts", "v1.2", "patch"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := bumpVersion(tc.version, tc.bump)
			if err == nil {
				t.Errorf("Expected error for %s with %s, but got none", tc.version, tc.bump)
			}
		})
	}
}

func TestDetectInstallMethod(t *testing.T) {
	// This test just ensures the function doesn't panic
	// Actual detection depends on runtime environment
	method := detectInstallMethod()

	validMethods := []string{"homebrew", "go install", "binary", ""}
	isValid := false
	for _, valid := range validMethods {
		if method == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		t.Errorf("Unexpected install method: %s", method)
	}
}
