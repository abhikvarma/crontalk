package models

import "time"

type Cron struct {
	Expression   string      `json:"expression"`
	Description  string      `json:"description"`
	NextRunTimes []time.Time `json:"next_run_times"`
}

func NewCron(expression, description string, nextRunTimes []time.Time) *Cron {
	return &Cron{
		expression,
		description,
		nextRunTimes,
	}
}
