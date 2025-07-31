package time

import (
	"testing"
	stdtime "time"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentDate(t *testing.T) {
	t.Run("should_return_current_date_in_correct_format_when_called", func(t *testing.T) {
		result := GetCurrentDate()

		assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, result)

		parsed, err := stdtime.Parse("2006-01-02", result)
		assert.NoError(t, err)
		assert.True(t, parsed.Year() >= 2020)
	})
}

func TestGetCurrentMonthYear(t *testing.T) {
	t.Run("should_return_current_month_year_in_correct_format_when_called", func(t *testing.T) {
		result := GetCurrentMonthYear()

		assert.Regexp(t, `^\d{4}-\d{2}$`, result)

		parsed, err := stdtime.Parse("2006-01", result)
		assert.NoError(t, err)
		assert.True(t, parsed.Year() >= 2020)
	})
}

func TestGetNextMonthStart(t *testing.T) {
	t.Run("should_return_first_day_of_next_month_at_midnight_when_called", func(t *testing.T) {
		result := GetNextMonthStart()

		now := stdtime.Now().UTC()
		assert.True(t, result.After(now))
		assert.Equal(t, 1, result.Day())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
	})
}

func TestGetFirstDayOfMonth(t *testing.T) {
	t.Run("should_return_first_day_of_current_month_at_midnight_when_called", func(t *testing.T) {
		result := GetFirstDayOfMonth()

		assert.Equal(t, 1, result.Day())
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())
	})
}

func TestGetLastDayOfMonth(t *testing.T) {
	t.Run("should_return_last_day_of_current_month_when_called", func(t *testing.T) {
		result := GetLastDayOfMonth()

		assert.True(t, result.Day() >= 28)
		assert.True(t, result.Day() <= 31)
		assert.Equal(t, 0, result.Hour())
	})
}

func TestFormatDate(t *testing.T) {
	t.Run("should_format_date_correctly_when_valid_time_provided", func(t *testing.T) {
		testTime := stdtime.Date(2024, 3, 15, 10, 30, 0, 0, stdtime.UTC)
		result := FormatDate(testTime)
		assert.Equal(t, "2024-03-15", result)
	})
}

func TestFormatMonthYear(t *testing.T) {
	t.Run("should_format_month_year_correctly_when_valid_time_provided", func(t *testing.T) {
		testTime := stdtime.Date(2024, 3, 15, 10, 30, 0, 0, stdtime.UTC)
		result := FormatMonthYear(testTime)
		assert.Equal(t, "2024-03", result)
	})
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name          string
		dateStr       string
		expectError   bool
		expectedYear  int
		expectedMonth stdtime.Month
		expectedDay   int
	}{
		{
			name:          "should_parse_valid_date_when_correct_format",
			dateStr:       "2024-03-15",
			expectError:   false,
			expectedYear:  2024,
			expectedMonth: stdtime.March,
			expectedDay:   15,
		},
		{
			name:        "should_return_error_when_invalid_format",
			dateStr:     "2024/03/15",
			expectError: true,
		},
		{
			name:        "should_return_error_when_invalid_date",
			dateStr:     "2024-13-32",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.dateStr)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedYear, result.Year())
				assert.Equal(t, tt.expectedMonth, result.Month())
				assert.Equal(t, tt.expectedDay, result.Day())
			}
		})
	}
}

func TestParseMonthYear(t *testing.T) {
	tests := []struct {
		name          string
		monthYearStr  string
		expectError   bool
		expectedYear  int
		expectedMonth stdtime.Month
	}{
		{
			name:          "should_parse_valid_month_year_when_correct_format",
			monthYearStr:  "2024-03",
			expectError:   false,
			expectedYear:  2024,
			expectedMonth: stdtime.March,
		},
		{
			name:         "should_return_error_when_invalid_format",
			monthYearStr: "2024/03",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMonthYear(tt.monthYearStr)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedYear, result.Year())
				assert.Equal(t, tt.expectedMonth, result.Month())
			}
		})
	}
}

func TestGetDaysInMonth(t *testing.T) {
	t.Run("should_return_days_in_current_month_when_called", func(t *testing.T) {
		result := GetDaysInMonth()
		assert.True(t, result >= 28)
		assert.True(t, result <= 31)
	})
}

func TestGetDaysInMonthForDate(t *testing.T) {
	tests := []struct {
		name     string
		date     stdtime.Time
		expected int
	}{
		{
			name:     "should_return_31_when_january",
			date:     stdtime.Date(2024, 1, 15, 0, 0, 0, 0, stdtime.UTC),
			expected: 31,
		},
		{
			name:     "should_return_29_when_february_leap_year",
			date:     stdtime.Date(2024, 2, 15, 0, 0, 0, 0, stdtime.UTC),
			expected: 29,
		},
		{
			name:     "should_return_28_when_february_non_leap_year",
			date:     stdtime.Date(2023, 2, 15, 0, 0, 0, 0, stdtime.UTC),
			expected: 28,
		},
		{
			name:     "should_return_30_when_november",
			date:     stdtime.Date(2024, 11, 15, 0, 0, 0, 0, stdtime.UTC),
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDaysInMonthForDate(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRemainingDaysInMonth(t *testing.T) {
	t.Run("should_return_remaining_days_including_today_when_called", func(t *testing.T) {
		result := GetRemainingDaysInMonth()
		assert.True(t, result >= 1)
		assert.True(t, result <= 31)
	})
}

func TestIsNewMonth(t *testing.T) {
	now := stdtime.Now().UTC()

	tests := []struct {
		name     string
		date     stdtime.Time
		expected bool
	}{
		{
			name:     "should_return_false_when_same_month",
			date:     now,
			expected: false,
		},
		{
			name:     "should_return_true_when_different_month",
			date:     now.AddDate(0, 1, 0),
			expected: true,
		},
		{
			name:     "should_return_true_when_different_year",
			date:     now.AddDate(1, 0, 0),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewMonth(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsNewDay(t *testing.T) {
	now := stdtime.Now().UTC()

	tests := []struct {
		name     string
		date     stdtime.Time
		expected bool
	}{
		{
			name:     "should_return_false_when_same_day",
			date:     now,
			expected: false,
		},
		{
			name:     "should_return_true_when_different_day",
			date:     now.AddDate(0, 0, 1),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNewDay(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSameDay(t *testing.T) {
	tests := []struct {
		name     string
		t1       stdtime.Time
		t2       stdtime.Time
		expected bool
	}{
		{
			name:     "should_return_true_when_same_day",
			t1:       stdtime.Date(2024, 3, 15, 10, 30, 0, 0, stdtime.UTC),
			t2:       stdtime.Date(2024, 3, 15, 14, 45, 0, 0, stdtime.UTC),
			expected: true,
		},
		{
			name:     "should_return_false_when_different_day",
			t1:       stdtime.Date(2024, 3, 15, 10, 30, 0, 0, stdtime.UTC),
			t2:       stdtime.Date(2024, 3, 16, 10, 30, 0, 0, stdtime.UTC),
			expected: false,
		},
		{
			name:     "should_return_false_when_different_month",
			t1:       stdtime.Date(2024, 3, 15, 10, 30, 0, 0, stdtime.UTC),
			t2:       stdtime.Date(2024, 4, 15, 10, 30, 0, 0, stdtime.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSameDay(tt.t1, tt.t2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSameMonth(t *testing.T) {
	tests := []struct {
		name     string
		t1       stdtime.Time
		t2       stdtime.Time
		expected bool
	}{
		{
			name:     "should_return_true_when_same_month",
			t1:       stdtime.Date(2024, 3, 15, 0, 0, 0, 0, stdtime.UTC),
			t2:       stdtime.Date(2024, 3, 25, 0, 0, 0, 0, stdtime.UTC),
			expected: true,
		},
		{
			name:     "should_return_false_when_different_month",
			t1:       stdtime.Date(2024, 3, 15, 0, 0, 0, 0, stdtime.UTC),
			t2:       stdtime.Date(2024, 4, 15, 0, 0, 0, 0, stdtime.UTC),
			expected: false,
		},
		{
			name:     "should_return_false_when_same_month_different_year",
			t1:       stdtime.Date(2024, 3, 15, 0, 0, 0, 0, stdtime.UTC),
			t2:       stdtime.Date(2023, 3, 15, 0, 0, 0, 0, stdtime.UTC),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSameMonth(tt.t1, tt.t2)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMonthStartEnd(t *testing.T) {
	t.Run("should_return_first_and_last_day_of_current_month_when_called", func(t *testing.T) {
		start, end := GetMonthStartEnd()

		assert.Equal(t, 1, start.Day())
		assert.Equal(t, 0, start.Hour())

		assert.True(t, end.Day() >= 28)
		assert.True(t, end.Day() <= 31)
		assert.Equal(t, start.Month(), end.Month())
		assert.Equal(t, start.Year(), end.Year())
	})
}

func TestGetTomorrowStart(t *testing.T) {
	t.Run("should_return_midnight_of_tomorrow_when_called", func(t *testing.T) {
		result := GetTomorrowStart()
		now := stdtime.Now().UTC()

		assert.True(t, result.After(now))
		assert.Equal(t, 0, result.Hour())
		assert.Equal(t, 0, result.Minute())
		assert.Equal(t, 0, result.Second())

		daysDiff := result.Day() - now.Day()
		if daysDiff < 0 {
			daysDiff = 1
		}
		assert.True(t, daysDiff == 1 || daysDiff == -30 || daysDiff == -29 || daysDiff == -27)
	})
}
