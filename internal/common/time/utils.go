package time

import (
	"time"
)

// GetCurrentDate returns the current date in "2006-01-02" format (UTC)
func GetCurrentDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

// GetCurrentMonthYear returns the current month-year in "2006-01" format (UTC)
func GetCurrentMonthYear() string {
	return time.Now().UTC().Format("2006-01")
}

// GetNextMonthStart returns the start of the next month
func GetNextMonthStart() time.Time {
	now := time.Now().UTC()
	// Get first day of current month
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	// Add one month
	return firstDayOfMonth.AddDate(0, 1, 0)
}

// GetFirstDayOfMonth returns the first day of the current month
func GetFirstDayOfMonth() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// GetLastDayOfMonth returns the last day of the current month
func GetLastDayOfMonth() time.Time {
	now := time.Now().UTC()
	firstDayOfNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	return firstDayOfNextMonth.Add(-24 * time.Hour)
}

// FormatDate formats a time.Time to "2006-01-02" format
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatMonthYear formats a time.Time to "2006-01" format
func FormatMonthYear(t time.Time) string {
	return t.Format("2006-01")
}

// ParseDate parses a date string in "2006-01-02" format
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// ParseMonthYear parses a month-year string in "2006-01" format
func ParseMonthYear(monthYearStr string) (time.Time, error) {
	return time.Parse("2006-01", monthYearStr)
}

// GetDaysInMonth returns the number of days in the current month
func GetDaysInMonth() int {
	now := time.Now().UTC()
	return GetDaysInMonthForDate(now)
}

// GetDaysInMonthForDate returns the number of days in the month of the given date
func GetDaysInMonthForDate(date time.Time) int {
	firstDayOfNextMonth := time.Date(date.Year(), date.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := firstDayOfNextMonth.Add(-24 * time.Hour)
	return lastDayOfMonth.Day()
}

// GetRemainingDaysInMonth returns the number of days remaining in the current month
func GetRemainingDaysInMonth() int {
	now := time.Now().UTC()
	daysInMonth := GetDaysInMonth()
	return daysInMonth - now.Day() + 1 // +1 to include today
}

// IsNewMonth checks if the given date is in a different month than today
func IsNewMonth(date time.Time) bool {
	now := time.Now().UTC()
	return date.Month() != now.Month() || date.Year() != now.Year()
}

// IsNewDay checks if the given date is a different day than today
func IsNewDay(date time.Time) bool {
	now := time.Now().UTC()
	return !IsSameDay(date, now)
}

// IsSameDay checks if two dates are on the same day
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsSameMonth checks if two dates are in the same month
func IsSameMonth(t1, t2 time.Time) bool {
	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()
	return y1 == y2 && m1 == m2
}

// GetMonthStartEnd returns the start and end of the current month
func GetMonthStartEnd() (time.Time, time.Time) {
	start := GetFirstDayOfMonth()
	end := GetLastDayOfMonth()
	return start, end
}

// GetTomorrowStart returns the start of tomorrow (midnight UTC)
func GetTomorrowStart() time.Time {
	now := time.Now().UTC()
	tomorrow := now.AddDate(0, 0, 1)
	return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
}
