package main

import (
	"fmt"
	"strings"
)

// OutputLevel prings a message to the terminal at given indention level.
func OutputLevel(msg string, level int) {
	switch level {
	case 0:
		{
			msg = " - " + msg
			break
		}
	case 1:
		{
			msg = " > " + msg
			break
		}
	default:
		{
			msg = " * " + msg
			break
		}
	}
	fmt.Println(strings.Repeat("  ", level) + msg)
}

// OutputMessage prints a message to the terminal.
func OutputMessage(msg string) {
	OutputLevel(msg, 0)
}

// OutputWarning prints an warning message.
func OutputWarning(msg string) {
	OutputLevel(
		"WARN: "+msg,
		0,
	)
}
