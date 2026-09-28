package pathresolver

import (
	"os"
	"testing"
	"time"
)

func TestResolve(t *testing.T) {
	// Mock time
	originalNow := Now
	defer func() { Now = originalNow }()

	// Fixed time: 2024-02-29 12:00:00 (Leap year)
	mockTime := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC)
	Now = func() time.Time {
		return mockTime
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	tests := []struct {
		name        string
		pattern     string
		expected    string
		expectError bool
	}{
		{
			name:        "today default format",
			pattern:     "/backup/{today}",
			expected:    "/backup/2024-02-29",
			expectError: false,
		},
		{
			name:        "today custom format",
			pattern:     "/backup/{today:DD-MM-YYYY}",
			expected:    "/backup/29-02-2024",
			expectError: false,
		},
		{
			name:        "yesterday leap year boundary",
			pattern:     "/backup/{yesterday}",
			expected:    "/backup/2024-02-28",
			expectError: false,
		},
		{
			name:        "offset positive leap year boundary",
			pattern:     "/backup/{offset:1}",
			expected:    "/backup/2024-03-01",
			expectError: false,
		},
		{
			name:        "hostname variable",
			pattern:     "/srv/{hostname}/data",
			expected:    "/srv/" + hostname + "/data",
			expectError: false,
		},
		{
			name:        "multiple variables",
			pattern:     "/{hostname}/{today:YYYY}/{yesterday:MM}/{offset:-2:DD}",
			expected:    "/" + hostname + "/2024/02/27",
			expectError: false,
		},
		{
			name:        "invalid variable",
			pattern:     "/backup/{unknown}",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid offset value",
			pattern:     "/backup/{offset:abc}",
			expected:    "",
			expectError: true,
		},
		{
			name:        "missing offset value",
			pattern:     "/backup/{offset}",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Resolve(tt.pattern)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil. Result: %s", result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestResolve_Boundaries(t *testing.T) {
	originalNow := Now
	defer func() { Now = originalNow }()

	tests := []struct {
		name     string
		mockTime time.Time
		pattern  string
		expected string
	}{
		{
			name:     "Month end (Jan 31 to Feb 1)",
			mockTime: time.Date(2023, 1, 31, 12, 0, 0, 0, time.UTC),
			pattern:  "{offset:1:YYYY-MM-DD}",
			expected: "2023-02-01",
		},
		{
			name:     "Year rollover (Dec 31 to Jan 1)",
			mockTime: time.Date(2023, 12, 31, 12, 0, 0, 0, time.UTC),
			pattern:  "{offset:1:YYYY-MM-DD}",
			expected: "2024-01-01",
		},
		{
			name:     "Year rollover backwards (Jan 1 to Dec 31)",
			mockTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			pattern:  "{yesterday:YYYY-MM-DD}",
			expected: "2023-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Now = func() time.Time { return tt.mockTime }
			result, err := Resolve(tt.pattern)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
