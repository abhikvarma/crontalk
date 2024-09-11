package cronutil

import "time"
import "github.com/robfig/cron/v3"

func GetNextRunTimes(expression string, count int) ([]time.Time, error) {
	schedule, err := cron.ParseStandard(expression)
	if err != nil {
		return nil, err
	}

	var times []time.Time
	now := time.Now()
	next := schedule.Next(now)

	for i := 0; i < 5; i++ {
		times = append(times, next)
		next = schedule.Next(next)
	}

	return times, nil
}
