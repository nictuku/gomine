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

// Built-in webserver that shows the status of the minecraft server.
package main

import (
	"flag"
	"fmt"
	"http"
	"template"
	"time"
)

var (
	listenAddr = flag.String("http", ":8080", "http listen address")
)

func logResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Printf(
		"%s - - [%s] \"%s %s %s\" - - \"%s\" \"%s\"\n",
		w.RemoteAddr(),
		time.UTC().Format("10/Jan/2006:15:04:05 -0700"),
		r.Method,
		r.URL.Path,
		r.Proto,
		r.Referer,
		r.UserAgent)
}

func Root(w http.ResponseWriter, r *http.Request) {
	logResponse(w, r)
	url := r.FormValue("url")
	if url == "" {
		err := templates["base"].Execute(w, nil)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
		}
		return
	}
	if r.Method != "POST" {
		http.Error(w, "only post is accepted, got "+r.Method, http.StatusInternalServerError)
		return
	}
	return
}

func Static(w http.ResponseWriter, r *http.Request) {
	logResponse(w, r)
	f := r.URL.Path[1:]
	http.ServeFile(w, r, f)
}

var templates = make(map[string]*template.Template)

func init() {
	for _, tmpl := range []string{"base"} {
		templates[tmpl] = template.MustParseFile("templates/"+tmpl+".html", nil)
	}
}
