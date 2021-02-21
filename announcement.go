package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Announcement defines an announcement for the bot to make.
type Announcement struct {
	Offset          int    `yaml:"offset"`
	Message         string `yaml:"message"`
	CheckAttendance bool   `yaml:"check_attendance"`
}

// OffsetHours returns number of hours the offset represents.
func (a Announcement) OffsetHours() int {
	return int(math.Floor(float64(a.Offset / 3600)))
}

// Duration returns offset as time.Duration type.
func (a Announcement) Duration() time.Duration {
	return time.Second * time.Duration(a.Offset)
}

// String returns announcement as human readable string.
func (a Announcement) String() string {
	offsetSuffix := "prior"
	if a.Offset > 0 {
		offsetSuffix = "after"
	}
	hours := a.Duration().Hours()
	return fmt.Sprintf(
		"%d hour(s) %s the event say...\n\t%s",
		int(math.Abs(hours)),
		offsetSuffix,
		a.Message,
	)
}

// MessageFilterTime display message with given time replacing all {TIME}.
func (a Announcement) MessageFilterTime(t StringTime) string {
	return strings.ReplaceAll(
		a.Message,
		"{TIME}",
		strings.Join(t.DisplayAllTimeZones(), " / "),
	)
}
