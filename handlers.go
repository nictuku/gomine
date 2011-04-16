package main

import (
	"fmt"
	"regexp"
	"time"
)

type eventHandler struct {
	re *regexp.Regexp
	f  func(*event)
}

type event struct {
	t *time.Time
	m string   "message"
	p []string // Parsed strings after regexp extraction with a matching eventHandler.
}

func (ev *event) Resolve() {
	for _, h := range eventHandlers {
		p := h.re.FindStringSubmatch(ev.m)
		if p == nil {
			continue
		}
		// Save the matched substrings into the event.
		ev.p = p
		h.f(ev)
		return
	}
	// XXX
	fmt.Printf(".. no event handler matched '%q'\n", ev.m)
}

var eventHandlers = []eventHandler{
	handleLogin,
}

// handleLogin
//2011-04-15 20:18:29 [INFO] nictuku [/84.72.7.79:56179] logged in with entity id 125
var handleLogin = eventHandler{
	regexp.MustCompile(`([^ ]+) \[([^\]]+)\] logged in with entity id [0-9]+`),
	func(ev *event) {
		fmt.Printf("User %v logged in", ev.p[2])
	},
}
