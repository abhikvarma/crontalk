package cron_internal

import (
	"fmt"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("cron_internal validation error in %s: %s", e.Field, e.Message)
}

type Expression struct {
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
}

func ParseCron(expression string) (*Expression, error) {
	fields := strings.Fields(expression)
	if fields == nil || len(fields) != 5 {
		return nil, &ValidationError{"expression", "must have 5 fields"}
	}

	return &Expression{
		fields[0],
		fields[1],
		fields[2],
		fields[3],
		fields[4],
	}, nil
}

func (c *Expression) Validate() error {
	validators := []struct {
		field    string
		value    string
		min, max int
		names    []string
	}{
		{"minute", c.Minute, 0, 59, nil},
		{"hour", c.Hour, 0, 23, nil},
		{"day of month", c.DayOfMonth, 0, 31, nil},
		{"month", c.Month, 1, 12, []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}},
		{"day of week", c.DayOfWeek, 0, 7, []string{"SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"}},
	}

	for _, v := range validators {
		if err := validateField(v.field, v.value, v.min, v.max, v.names); err != nil {
			return err
		}
	}
	return nil
}

func validateField(field, value string, min, max int, names []string) error {
	if value == "*" {
		return nil
	}

	if strings.Contains(value, "/") {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			return &ValidationError{field, "invalid value"}
		}
		if parts[0] != "*" {
			return &ValidationError{field, "must start with a '*'"}
		}
		if _, err := strconv.Atoi(parts[1]); err != nil {
			return &ValidationError{field, "value after '\\' must be a number"}
		}
		value = parts[1]
	}

	if strings.Contains(value, "-") {
		return validateRange(field, value, min, max, names)
	}

	if strings.Contains(value, ",") {
		if err := validateList(field, value, min, max, names); err != nil {
			return &ValidationError{field, "invalid value in list"}
		}
		return nil
	}

	if value == "?" && (field == "day of month" || field == "day of week") {
		return nil
	}

	if value == "L" && (field == "day of month" || field == "day of week") {
		return nil
	}

	if strings.HasSuffix(value, "W") && field == "day of month" {
		dayStr := strings.TrimSuffix(value, "W")
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 31 {
			return &ValidationError{field, "invalid weekday value"}
		}
		return nil
	}

	if strings.Contains(value, "#") && field == "day of week" {
		parts := strings.Split(value, "#")
		if len(parts) != 2 {
			return &ValidationError{field, "invalid nth weekday of month"}
		}
		weekday, err1 := strconv.Atoi(parts[0])
		n, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || weekday < 1 || weekday > 7 || n < 1 || n > 5 {
			return &ValidationError{field, "invalid nth weekday of month"}
		}
		return nil
	}

	if _, err := strconv.Atoi(value); err == nil {
		return validateNumber(field, value, min, max)
	} else {
		for _, name := range names {
			if strings.EqualFold(value, name) {
				return nil
			}
		}
	}

	return &ValidationError{Field: field, Message: "invalid value"}
}

func validateRange(field, value string, min, max int, names []string) error {
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return &ValidationError{field, "incomplete range"}
	}

	var start, end int
	var err error

	start, err = strconv.Atoi(parts[0])
	if err != nil {
		start = findNameIndex(parts[0], names)
		if start == -1 {
			return &ValidationError{field, "invalid range start"}
		}
		start++ // Adjust index to 1-based for consistency with numeric values
	}

	// Check if end is a number or a name
	end, err = strconv.Atoi(parts[1])
	if err != nil {
		end = findNameIndex(parts[1], names)
		if end == -1 {
			return &ValidationError{field, "invalid range end"}
		}
		end++ // Adjust index to 1-based for consistency with numeric values
	}

	if start < min || start > max || end < min || end > max || end < start {
		return &ValidationError{field, "invalid range"}
	}

	return nil
}

func validateList(field, value string, min, max int, names []string) error {
	parts := strings.Split(value, ",")
	for _, part := range parts {
		if err := validateField(field, part, min, max, names); err != nil {
			return &ValidationError{field, err.Error()}
		}
	}
	return nil
}

func validateNumber(field, value string, min, max int) error {
	num, err := strconv.Atoi(value)
	if err != nil || num < min || num > max {
		return &ValidationError{field, fmt.Sprintf("value must be between %d and %d", min, max)}
	}
	return nil
}

func findNameIndex(name string, names []string) int {
	for i, n := range names {
		if strings.EqualFold(name, n) {
			return i
		}
	}
	return -1
}
