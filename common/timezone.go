package common

import "time"

func ServerLocation() *time.Location {
	loc, err := time.LoadLocation("Local")
	if err != nil || loc == nil {
		return time.UTC
	}
	return loc
}

func ServerNow() time.Time {
	return time.Now().In(ServerLocation())
}

func ParseInServerLocation(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ServerLocation())
}
