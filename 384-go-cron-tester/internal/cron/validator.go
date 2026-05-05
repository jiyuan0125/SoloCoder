package cron

func Validate(expr string) error {
	_, err := Parse(expr)
	return err
}
