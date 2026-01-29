package datetime

import (
	"time"
	_ "time/tzdata"
)

func ParseTime(timeString string) (time.Time, error) {
	layout := time.RFC3339

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.Time{}, err
	}

	parsedTime, err := time.Parse(layout, timeString)
	if err != nil {
		return time.Time{}, err
	}

	return parsedTime.In(loc), nil
}
