package main

import (
	"fmt"
	"strings"
	"time"
)

// Event defines an event for the bot to announce.
type Event struct {
	Day           StringDay      `json:"day"`
	Time          StringTime     `json:"time"`
	Announcements []Announcement `json:"announcements"`
}

// String returns event as human readable string.
func (e Event) String() string {
	return fmt.Sprintf("%s at %s", e.Day.Weekday().String(), strings.Join(e.Time.DisplayAllTimeZones(), " / "))
}

// Duration returns the time until this event from Sunday.
func (e Event) Duration() time.Duration {
	return e.Day.Duration() + e.Time.Duration()
}

// Empty returns true if event is empty.
func (e Event) Empty() bool {
	return string(e.Day) == "" && string(e.Time) == ""
}

// NextAnnouncement returns next announcement for event.
func (e Event) NextAnnouncement() (*Announcement, time.Duration) {
	if len(e.Announcements) == 0 {
		return nil, time.Duration(0)
	}
	now := time.Now()
	currentTimeDur := time.Second * time.Duration((int(now.Weekday())*86400)+(now.Hour()*3600)+(now.Minute()*60))
	nextDuration := time.Duration(0)
	var nextAn *Announcement
	for _, offset := range []int{604800, 0} {
		for i, an := range e.Announcements {
			dur := e.Duration() + an.Duration() - currentTimeDur + (time.Second * time.Duration(offset))
			if dur < 0 {
				continue
			}
			if nextAn == nil || dur < nextDuration {
				nextAn = &e.Announcements[i]
				nextDuration = dur
			}
		}
	}
	return nextAn, nextDuration
}

// NextEvent returns the next upcoming event and the time until the event occurs.
func NextEvent() (*Event, time.Duration) {
	now := time.Now()
	currentTimeDur := time.Second * time.Duration((int(now.Weekday())*86400)+(now.Hour()*3600)+(now.Minute()*60))
	nextDuration := time.Duration(0)
	var nextEvent *Event
	for _, offset := range []int{604800, 0} {
		for i, event := range config.Events {
			dur := event.Duration() - currentTimeDur + (time.Second * time.Duration(offset))
			if dur < 0 {
				continue
			}
			if nextEvent == nil || dur < nextDuration {
				nextEvent = &config.Events[i]
				nextDuration = dur
			}
		}
	}
	return nextEvent, nextDuration
}

// FireEvent returns the active event or nil if none are active.
func FireEvent() *Event {
	nextEvent, nextDur := NextEvent()
	if nextEvent != nil && nextDur.Seconds() <= 0 && nextDur.Seconds() > -60 {
		return nextEvent
	}
	return nil
}

// FireAnnouncement returns the active anouncement and its coresponding event or nil if none are active.
func FireAnnouncement() (*Event, *Announcement) {
	nextEvent, _ := NextEvent()
	if nextEvent == nil {
		return nil, nil
	}
	nextAn, nextDur := nextEvent.NextAnnouncement()
	if nextAn != nil && nextDur.Seconds() <= 0 && nextDur.Seconds() > -60 {
		return nextEvent, nextAn
	}
	return nil, nil
}
