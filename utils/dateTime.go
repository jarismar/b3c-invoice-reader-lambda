package utils

import "time"

func GetDateObject(dateTime string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateTime)
}

func ToFirstDayOfMonth(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.Local)
}
