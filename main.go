package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bwmarrin/discordgo"
)

// https://discordapp.com/oauth2/authorize?&client_id=812547382441017384&scope=bot&permissions=76864

// Config - configuration for bot
type Config struct {
	Token            string   `yaml:"token"`
	ChannelID        string   `yaml:"channel_id"`
	Timezone         string   `yaml:"timezone"`
	DisplayTimezones []string `yaml:"display_timezones"`
	Events           []Event  `yaml:"events"`
	NextMessage      string   `yaml:"next_message"`
	//RoleID           string   `yaml:"role_id"`
	//GuildID          string   `yaml:"guild_id"`
}

const skipFlagFile = ".skip"

// config
var config Config
var localTZName string
var localOffset int
var lastMessageID string

func main() {

	// load config
	OutputMessage("Load config.yaml.")
	config = Config{}
	configBytes, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
		panic(err)
	}

	// get timezone offset
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		panic(err)
	}
	localTZName, localOffset = time.Now().In(loc).Zone()

	// read out events
	OutputMessage("Event times are...")
	for _, event := range config.Events {
		OutputLevel(event.String(), 1)
		for _, an := range event.Announcements {
			OutputLevel(an.String(), 2)
		}
	}

	// display next event
	nextEvent, _ := NextEvent()
	OutputMessage(fmt.Sprintf("Next event on %s.", nextEvent.String()))

	// get bot token
	if config.Token == "" {
		panic(errors.New("bot token not provided"))
	}

	// init discordgo
	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		panic(err)
	}
	defer dg.Close()

	// add handlers
	dg.AddHandler(ready)
	dg.AddHandler(messageReactionAdd)

	// start
	err = dg.Open()
	if err != nil {
		panic(err)
	}
	OutputMessage("Bot ready.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

}

func setSkip() error {
	OutputMessage("Skip next event flag set.")
	return ioutil.WriteFile(skipFlagFile, []byte("0"), 0655)
}

func resetSkip() error {
	OutputMessage("Skip next event flag removed.")
	return os.Remove(skipFlagFile)
}

func isSkip() bool {
	_, err := os.Stat(skipFlagFile)
	if err != nil {
		if !os.IsNotExist(err) {
			OutputWarning(err.Error())
		}
		return false
	}
	return true
}

func updateStatus(s *discordgo.Session) {
	_, nextDur := NextEvent()
	// calculate days and hours left
	days := int(math.Floor(float64(nextDur.Hours() / 24)))
	hours := int(nextDur.Hours()) % 24
	// build time message
	timeRepStr := "less than an hour"
	if days > 0 || hours > 0 {
		timeRepStr = ""
		if days > 0 {
			// determine if day should be plural
			dayStr := "day"
			if days != 1 {
				dayStr += "s"
			}
			// build day(s) message
			timeRepStr = fmt.Sprintf("%d %s", days, dayStr)
		}
		if hours > 0 {
			// determine if hour should be plural
			hourStr := "hour"
			if hours != 1 {
				hourStr += "s"
			}
			// build hours(s) message
			if timeRepStr != "" {
				timeRepStr += " "
			}
			timeRepStr += fmt.Sprintf("%d %s", hours, hourStr)
		}
	}
	// build message
	nextMsg := strings.ReplaceAll(
		config.NextMessage,
		"{TIME}",
		timeRepStr,
	)
	// update status
	s.UpdateGameStatus(
		0,
		nextMsg,
	)
}

func sendAnnouncement(s *discordgo.Session, e Event, a Announcement) {
	msg := a.MessageFilterTime(e.Time)
	OutputMessage("Send announcement: " + msg)
	dm, err := s.ChannelMessageSend(
		config.ChannelID,
		msg,
	)
	if err != nil {
		OutputWarning(err.Error())
	}
	lastMessageID = dm.ID
	if a.CheckAttendance {
		if err := s.MessageReactionAdd(config.ChannelID, dm.ID, "👍"); err != nil {
			OutputWarning(err.Error())
		}
	}
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	updateStatus(s)
	go func() {
		for range time.Tick(time.Minute * 30) {
			updateStatus(s)
		}
	}()
	for range time.Tick(time.Minute) {
		nEvent, cAnnouncement := FireAnnouncement()
		if cAnnouncement != nil && !isSkip() {
			sendAnnouncement(s, *nEvent, *cAnnouncement)
		}
		if FireEvent() != nil && isSkip() {
			if err := resetSkip(); err != nil {
				OutputWarning(err.Error())
			}
		}
	}
}

func messageReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	log.Println(m.Emoji.ID, m.Emoji.Name)
	switch m.Emoji.Name {
	case "🛑":
		{
			// only accept from last message + don't accept reactions from this bot
			if m.UserID == s.State.User.ID || m.MessageID != lastMessageID {
				return
			}
			OutputMessage("Stop reaction detected.")
			if err := setSkip(); err != nil {
				OutputWarning(err.Error())
			}
			return
		}
	case "👍":
		{
			// only accept from last message + don't accept reactions from this bot
			if m.UserID == s.State.User.ID || m.MessageID != lastMessageID {
				return
			}
			// TODO do more with this, ie let people know if they haven't confirmed attendence
			OutputMessage(fmt.Sprintf(
				"User %s +1'd.", m.UserID,
			))
			return
		}
	case "👎":
		{
			return
		}
	case "✅":
		{
			// TODO reserve checkbox for bot use only, use it to let everyone know that attendence has been fully confirmed
			/*if m.UserID == s.State.User.ID || m.MessageID != lastMessageID {
				return
			}
			if err := s.MessageReactionRemove(m.ChannelID, m.MessageID, m.Emoji.Name, m.UserID); err != nil {
				OutputWarning(err.Error())
			}*/
			return
		}
	}
}
