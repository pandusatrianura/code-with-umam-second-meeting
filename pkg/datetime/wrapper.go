package datetime

import (
	"log"
	"time"
)

func ParseTime(timeString string) (time.Time, error) {
	layout := time.RFC3339

	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Println("LoadLocation err:", err.Error())
		return time.Time{}, err
	}

	parsedTime, err := time.Parse(layout, timeString)
	if err != nil {
		log.Println("parsedTime err:", err.Error())
		return time.Time{}, err
	}

	log.Println("parsedTime:", parsedTime.In(loc))

	return parsedTime.In(loc), nil
}
