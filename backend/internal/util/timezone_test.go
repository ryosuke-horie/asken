package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDayRangeInTimezone(t *testing.T) {
	tests := []struct {
		name          string
		dateStr       string
		tz            string
		expectedStart time.Time
		expectedEnd   time.Time
		expectError   bool
	}{
		{
			name:          "UTC timezone",
			dateStr:       "2026-02-03",
			tz:            "UTC",
			expectedStart: time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC),
			expectError:   false,
		},
		{
			name:          "JST timezone (Asia/Tokyo)",
			dateStr:       "2026-02-03",
			tz:            "Asia/Tokyo",
			expectedStart: time.Date(2026, 2, 2, 15, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 2, 3, 15, 0, 0, 0, time.UTC),
			expectError:   false,
		},
		{
			name:          "empty timezone defaults to UTC",
			dateStr:       "2026-02-03",
			tz:            "",
			expectedStart: time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 2, 4, 0, 0, 0, 0, time.UTC),
			expectError:   false,
		},
		{
			name:        "invalid timezone",
			dateStr:     "2026-02-03",
			tz:          "Invalid/Timezone",
			expectError: true,
		},
		{
			name:        "invalid date format",
			dateStr:     "2026/02/03",
			tz:          "UTC",
			expectError: true,
		},
		{
			name:          "US Pacific timezone",
			dateStr:       "2026-02-03",
			tz:            "America/Los_Angeles",
			expectedStart: time.Date(2026, 2, 3, 8, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 2, 4, 8, 0, 0, 0, time.UTC),
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := GetDayRangeInTimezone(tt.dateStr, tt.tz)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStart, start)
			assert.Equal(t, tt.expectedEnd, end)
		})
	}
}

func TestParseDateInTimezone(t *testing.T) {
	tests := []struct {
		name        string
		dateStr     string
		tz          string
		expected    time.Time
		expectError bool
	}{
		{
			name:        "UTC timezone",
			dateStr:     "2026-02-03",
			tz:          "UTC",
			expected:    time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "JST timezone returns UTC equivalent",
			dateStr:     "2026-02-03",
			tz:          "Asia/Tokyo",
			expected:    time.Date(2026, 2, 2, 15, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "empty timezone defaults to UTC",
			dateStr:     "2026-02-03",
			tz:          "",
			expected:    time.Date(2026, 2, 3, 0, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "invalid timezone",
			dateStr:     "2026-02-03",
			tz:          "Invalid/Timezone",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDateInTimezone(tt.dateStr, tt.tz)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
