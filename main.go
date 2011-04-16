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
	"bufio"
	"flag"
	"fmt"
	"http"
	"os"
	"regexp"
	"time"
)

var consoleMsg = regexp.MustCompile("([^ ]+ [^ ]+) ([^ ]+) (.*)\n")
var timeLayout = "2006-01-02 15:04:05"
var twitter = newTwitterClient()

func parseLine(line string) (e *event, err os.Error) {
	fmt.Print(line)
	matches := consoleMsg.FindStringSubmatch(line)
	if matches == nil {
		return nil, os.NewError("Line format unknown.")
	}
	t, err := time.Parse(timeLayout, matches[1])
	if err != nil {
		return nil, err
	}
	// ignore verbose level.
	return &event{t, matches[3], nil}, nil
}

func main() {
	flag.Parse()
	http.HandleFunc("/", Root)
	http.HandleFunc("/static/", Static)
	go http.ListenAndServe(*listenAddr, nil)
	// The minecraft server actually writes to stderr, but reading from
	// stdin makes things easier since I can use bash and a pipe.
	stdin := bufio.NewReader(os.Stdin)
	for {
		line, err := stdin.ReadString('\n')
		if err != nil && err.String() == "EOF" {
			break
		}
		if err != nil || len(line) <= 1 {
			continue
		}
		ev, err := parseLine(line)
		if err != nil {
			fmt.Println("parseLine error:", err)
			continue
		}
		ev.Resolve()
	}
	os.Exit(0)
}
