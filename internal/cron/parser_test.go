package cron

import (
	"errors"
	"testing"
)

func TestParseCron(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		wantErr      bool
		errorField   string
		errorMessage string
	}{
		{"Valid expression", "0 5 * * 1-5", false, "", ""},
		{"Invalid number of fields", "0 5 * *", true, "expression", "must have 5 fields"},
		{"Invalid minute", "60 5 * * *", true, "minute", "value must be between 0 and 59"},
		{"Invalid hour", "0 24 * * *", true, "hour", "value must be between 0 and 23"},
		{"Invalid day of month", "0 5 32 * *", true, "day of month", "value must be between 1 and 31"},
		{"Invalid month", "0 5 * 13 *", true, "month", "value must be between 1 and 12"},
		{"Invalid day of week", "0 5 * * 8", true, "day of week", "value must be between 0 and 7"},
		{"Valid with special chars", "*/15 0 1,15 * 1-5", false, "", ""},
		{"Valid with names", "0 5 * JAN-DEC MON-FRI", false, "", ""},
		{"Invalid step value", "*/a 5 * * *", true, "minute", "invalid step value"},
		{"Invalid range", "0 5 15-10 * *", true, "day of month", "invalid range"},
		{"Valid with L", "0 5 L * *", false, "", ""},
		{"Valid with W", "0 5 15W * *", false, "", ""},
		{"Invalid W usage", "0 5 32W * *", true, "day of month", "invalid weekday value"},
		{"Valid with #", "0 5 * * 2#1", false, "", ""},
		{"Invalid # usage", "0 5 * * 8#1", true, "day of week", "invalid nth weekday of month"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cron, err := ParseCron(tt.expression)
			if err != nil {
				var cronErr *ValidationError
				if errors.As(err, &cronErr) {
					if cronErr.Field != tt.errorField || cronErr.Message != tt.errorMessage {
						t.Errorf("ParseCron() error = %v, want field %v, message %v", cronErr, tt.errorField, tt.errorMessage)
					}
				}
				return
			}
			if err := cron.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("CronExpression.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateField(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		min      int
		max      int
		names    []string
		wantErr  bool
		errField string
		errMsg   string
	}{
		{"Valid number", "minute", "30", 0, 59, nil, false, "", ""},
		{"Valid asterisk", "hour", "*", 0, 23, nil, false, "", ""},
		{"Valid step", "minute", "*/15", 0, 59, nil, false, "", ""},
		{"Invalid step", "minute", "*/60", 0, 59, nil, true, "minute", "value must be between 0 and 59"},
		{"Valid range", "hour", "9-17", 0, 23, nil, false, "", ""},
		{"Invalid range", "hour", "17-9", 0, 23, nil, true, "hour", "invalid range"},
		{"Valid list", "day of week", "1,3,5", 0, 7, nil, false, "", ""},
		{"Invalid list item", "day of week", "1,8,5", 0, 7, nil, true, "day of week", "invalid value in list"},
		{"Valid L usage", "day of month", "L", 1, 31, nil, false, "", ""},
		{"Valid W usage", "day of month", "15W", 1, 31, nil, false, "", ""},
		{"Invalid W usage", "day of month", "32W", 1, 31, nil, true, "day of month", "invalid weekday value"},
		{"Valid # usage", "day of week", "2#1", 0, 7, nil, false, "", ""},
		{"Invalid # usage", "day of week", "8#1", 0, 7, nil, true, "day of week", "invalid nth weekday of month"},
		{"Valid month name", "month", "JAN", 1, 12, []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}, false, "", ""},
		{"Invalid month name", "month", "FOO", 1, 12, []string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}, true, "month", "invalid value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateField(tt.field, tt.value, tt.min, tt.max, tt.names)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateField() error in value %s = %v, wantErr %v", tt.value, err, tt.wantErr)
				return
			}
			if err != nil {
				var cronErr *ValidationError
				if errors.As(err, &cronErr) {
					if cronErr.Field != tt.errField || cronErr.Message != tt.errMsg {
						t.Errorf("validateField() error in value %s = %v, want field %v, message %v", tt.value, cronErr, tt.errField, tt.errMsg)
					}
				}
			}
		})
	}
}
