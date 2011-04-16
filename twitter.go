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
	"github.com/garyburd/twister/oauth"
	"github.com/garyburd/twister/web"
	"http"
	"io/ioutil"
	"json"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	TWITTER_API_BASE    = "http://api.twitter.com/1"
	TWITTER_GET_TIMEOUT = 10 // seconds.
)

var oauthClient = oauth.Client{
	Credentials:                   oauth.Credentials{clientToken, clientSecret},
	TemporaryCredentialRequestURI: "http://api.twitter.com/oauth/request_token",
	ResourceOwnerAuthorizationURI: "http://api.twitter.com/oauth/authenticate",
	TokenRequestURI:               "http://api.twitter.com/oauth/access_token",
}

type twitterClient struct {
	twitterToken *oauth.Credentials
}

func newTwitterClient() *twitterClient {
	return &twitterClient{twitterToken: &oauth.Credentials{accessToken, accessTokenSecret}}
}

func (tw *twitterClient) twitterGet(url string, param web.ParamMap) (p []byte, err os.Error) {
	oauthClient.SignParam(tw.twitterToken, "GET", url, param)
	url = url + "?" + param.FormEncodedString()
	var resp *http.Response
	done := make(chan bool, 1)
	go func() {
		resp, _, err = http.Get(url)
		done <- true
	}()

	timeout := time.After(TWITTER_GET_TIMEOUT * 1e9) //
	select {
	case <-done:
		break
	case <-timeout:
		return nil, os.NewError("http Get timed out - " + url)
	}
	if resp == nil {
		panic("oops")
	}
	return readHttpResponse(resp, err)
}

// Data in param must be URL escaped already.
func (tw *twitterClient) twitterPost(url string, param web.ParamMap) (p []byte, err os.Error) {
	oauthClient.SignParam(tw.twitterToken, "POST", url, param)

	// TODO: remove this dupe.
	var resp *http.Response
	done := make(chan bool, 1)
	go func() {
		resp, err = http.PostForm(url, param.StringMap())
		done <- true
	}()

	timeout := time.After(TWITTER_GET_TIMEOUT * 1e9) // post in this case.
	select {
	case <-done:
		break
	case <-timeout:
		return nil, os.NewError("http POST timed out - " + url)
	}
	if resp == nil {
		panic("oops")
	}
	return readHttpResponse(resp, err)
}


func (tw *twitterClient) Update(status string) (err os.Error) {
	if len(status) > 140 {
		return os.NewError("Message exceeds twitter char limit: " + status)
	}
	url := TWITTER_API_BASE + "/statuses/update.json"
	param := make(web.ParamMap)
	param.Set("status", status)
	var p []byte
	if p, err = tw.twitterPost(url, param); err != nil {
		log.Println("Update error:", err)
		log.Printf("response: %q", p)
	} else {
		log.Printf("Status update: %v", status)
	}
	return
}

func parseResponseError(p []byte) string {
	var r map[string]string
	if err := json.Unmarshal(p, &r); err != nil {
		log.Printf("parseResponseError json.Unmarshal error: %v", err)
		return ""
	}
	e, ok := r["error"]
	if !ok {
		return ""
	}
	return e

}

func readHttpResponse(resp *http.Response, httpErr os.Error) (p []byte, err os.Error) {
	err = httpErr
	if err != nil {
		log.Println(err.String())
		return nil, err
	}
	p, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	rateLimitStats(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		e := parseResponseError(p)
		if e == "" {
			e = "unknown"
		}
		err = os.NewError(fmt.Sprintf("Server Error code: %d; msg: %v", resp.StatusCode, e))
		return nil, err
	}
	return p, nil

}

func rateLimitStats(resp *http.Response) {
	if resp == nil {
		return
	}
	curr := time.Seconds()
	reset, _ := strconv.Atoi64(resp.GetHeader("X-RateLimit-Reset"))
	remaining, _ := strconv.Atoi64(resp.GetHeader("X-RateLimit-Remaining"))
	if remaining < 1 && reset-curr > 0 {
		log.Printf("Twitter API limits exceeded. Sleeping for %d seconds.\n", reset-curr)
		time.Sleep((reset - curr) * 1e9)
	}
}
