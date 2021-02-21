package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// StringDay represents the day of the week as user inputable string.
type StringDay string

// Weekday returns the day of the week as time.Weekday.
func (s StringDay) Weekday() time.Weekday {
	switch strings.ToLower(strings.TrimSpace(string(s))) {
	case "monday", "mon", "mo", "m":
		{
			return time.Monday
		}
	case "tuesday", "tues", "tue", "tu", "t":
		{
			return time.Tuesday
		}
	case "wednesday", "weds", "wed", "we", "w":
		{
			return time.Wednesday
		}
	case "thursday", "thur", "thu", "th":
		{
			return time.Thursday
		}
	case "friday", "fri", "fr", "f":
		{
			return time.Friday
		}
	case "saturday", "sat", "sa":
		{
			return time.Saturday
		}
	default:
		{
			return time.Sunday
		}
	}
}

// Duration returns the day as a duration since Sunday.
func (s StringDay) Duration() time.Duration {
	return time.Second * time.Duration(int(s.Weekday())*86400)
}

// StringTime represents the time of day as user inputable string.
type StringTime string

// Duration returns the time as duration since midnight.
func (s StringTime) Duration() time.Duration {
	hours := 0
	minutes := 0
	timeSplit := strings.Split(string(s), ":")
	if len(timeSplit) >= 2 {
		hours, _ = strconv.Atoi(timeSplit[0])
		if hours < 0 {
			hours = 0
		} else if hours > 23 {
			hours = 23
		}
		minutes, _ = strconv.Atoi(timeSplit[1])
		if minutes < 0 {
			minutes = 0
		} else if minutes > 59 {
			minutes = 59
		}
	}
	return time.Second * time.Duration((hours*3600)+(minutes*60))
}

// Display returns human readable time of day event takes place for given timezone offset.
func (s StringTime) Display(offset int) string {
	dur := s.Duration() - (time.Second * time.Duration(localOffset-offset))
	hour := int(dur.Hours()) % 24
	min := int(dur.Minutes()) % 60
	ampm := "AM"
	if hour < 0 {
		hour = 24 + hour
	}
	if hour > 12 {
		ampm = "PM"
		hour -= 12
	}
	if hour == 0 {
		hour = 12
	}
	return fmt.Sprintf(
		"%02d:%02d%s",
		hour,
		min,
		ampm,
	)
}

// DisplayAllTimeZones get display for all configured display timezones.
func (s StringTime) DisplayAllTimeZones() []string {
	out := make([]string, 0)
	for _, tzConfName := range config.DisplayTimezones {
		loc, err := time.LoadLocation(tzConfName)
		if err != nil {
			OutputWarning(
				fmt.Sprintf("Tried to load timezone %s, got error, %s", tzConfName, err.Error()),
			)
			continue
		}
		tzName, offset := time.Now().In(loc).Zone()
		out = append(
			out,
			s.Display(offset)+" "+tzName,
		)
	}
	return out
}
