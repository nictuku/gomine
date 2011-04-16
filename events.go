// Copyright 2011 Yves Junqueira
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

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
	message string
	details []string // Parsed strings after regexp extraction with a matching eventHandler.
}

func (ev *event) Resolve() {
	for _, h := range eventHandlers {
		p := h.re.FindStringSubmatch(ev.message)
		if p == nil {
			continue
		}
		// Save the matched substrings into the event.
		ev.details = p
		h.f(ev)
		return
	}
}

var eventHandlers = []eventHandler{
	handleLogin,
}

// Example:
// 2011-04-15 20:18:29 [INFO] nictuku [/84.72.7.79:56179] logged in with entity id 125
var handleLogin = eventHandler{
	regexp.MustCompile(`([^ ]+) \[([^\]]+)\] logged in with entity id [0-9]+`),
	func(ev *event) {
		msg := fmt.Sprintf("User %v logged in!", ev.details[1])
		fmt.Println("twit:", msg)
		err := twitter.Update(msg)
		if err != nil {
			fmt.Println("Twitter error:", err)
		}
	},
}
