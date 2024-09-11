package cron_internal

func ValidateCron(expression string) error {
	cronExp, err := ParseCron(expression)
	if err != nil {
		return err
	}
	return cronExp.Validate()
}
