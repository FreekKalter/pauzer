// Copyright 2013 (c) Freek Kalter. All rights reserved.
// Use of this source code is governed by the "Revised BSD License"
// that can be found in the LICENSE file.

// pauzer is a website to pause a local sabnzb instance without exposing the whole interface.
package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	api_key   = "d2ef95d20181d30d884321fb9cb68cbe"
	api_url   = "https://localhost:9100/sabnzbd/"
	max_speed = 1000
)

// ignore invalid certificates (todo: make it accecpt a valid cert)
var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var client = &http.Client{Transport: tr}

type CountDown struct {
	SetAt                  time.Time
	Duration               time.Duration
	Limit, LimitPercentage int64
}

func (c CountDown) ExpiresAt() (expire time.Time, err error) {
	if c.SetAt.Equal(time.Unix(0, 0)) {
		err = errors.New("timer not running")
	}
	expire = c.SetAt.Add(c.Duration)
	return
}

func (c CountDown) SecondsLeft() (secs int64, err error) {
	expires, err := c.ExpiresAt()
	if err != nil {
		err = errors.New("timer not running")
	}
	secs = int64(expires.Sub(time.Now()).Seconds())
	return
}

var countDown CountDown = CountDown{
	SetAt:    time.Unix(0, 0),
	Duration: 0,
	Limit:    0,
}

var compiledTemplates = template.Must(template.ParseFiles("404.html"))

var sabNzbFunctions map[string]string = map[string]string{
	"reset_limit":     fmt.Sprintf("%vapi?mode=config&name=speedlimit&value=0&apikey=%v", api_url, api_key),
	"resume_download": fmt.Sprintf("%vapi?mode=resume&apikey=%v", api_url, api_key),
	"pause":           fmt.Sprintf("%vapi?mode=config&name=set_pause&value=%%v&apikey=%v", api_url, api_key),
	"limit":           fmt.Sprintf("%vapi?mode=config&name=speedlimit&value=%%v&apikey=%v", api_url, api_key),
}

func HomeHandler(
	w http.ResponseWriter,
	r *http.Request) {

	indexContent, err := ioutil.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(indexContent))
}

func ResumeHandler(w http.ResponseWriter, r *http.Request) {
	countDown.Duration = 0
	countDown.Limit = 0
	call_sabnzbd(sabNzbFunctions["resume_download"])
	call_sabnzbd(sabNzbFunctions["reset_limit"])
}

func FormHandler(w http.ResponseWriter, r *http.Request) {
	formVars := mux.Vars(r)
	valid_integer_regex := regexp.MustCompile("^[0-9]{1,3}$")
	if !valid_integer_regex.MatchString(strings.TrimSpace(formVars["time"])) ||
		!valid_integer_regex.MatchString(strings.TrimSpace(formVars["limit"])) {
		countDown.Duration = 0
		return
	} else {
		timer_value, _ := strconv.ParseInt(formVars["time"], 10, 32)               //base 10, 32bit integer
		countDown.LimitPercentage, _ = strconv.ParseInt(formVars["limit"], 10, 32) //base 10, 32bit integer
		countDown.Duration = time.Minute * time.Duration(timer_value)
		countDown.Limit = max_speed - ((max_speed / 100) * countDown.LimitPercentage) // percentage give is how much to block, so inverse that to get how much to let through
		time.AfterFunc(countDown.Duration, func() {
			countDown.Duration = 0
			countDown.SetAt = time.Unix(0, 0)
			call_sabnzbd(sabNzbFunctions["resume_download"])
			call_sabnzbd(sabNzbFunctions["reset_limit"])
		})

		if countDown.LimitPercentage == 100 {
			go call_sabnzbd(fmt.Sprintf(sabNzbFunctions["pause"], timer_value))
		} else {
			go call_sabnzbd(fmt.Sprintf(sabNzbFunctions["limit"], countDown.Limit))
		}
		countDown.SetAt = time.Now()
	}
}

func CurrentStateHandler(
	w http.ResponseWriter,
	r *http.Request) {
	var limit, dur int64
	secs, err := countDown.SecondsLeft()
	if err != nil || secs <= 0 {
		limit = 0
		dur = 0
	} else {
		dur = int64(countDown.Duration.Minutes())
		limit = countDown.LimitPercentage
	}
	state := map[string]int64{"secondsLeft": secs, "limit": limit, "time": dur}

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(state)
}

func NotFound(
	w http.ResponseWriter,
	r *http.Request) {

	err := compiledTemplates.ExecuteTemplate(w, "404.html", r.URL)
	if err != nil {
		panic(err)
	}
}

func call_sabnzbd(url string) {
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/action/{time:[0-9]+}/{limit:[0-9]+}", FormHandler).Name("pause")
	r.HandleFunc("/resume", ResumeHandler).Name("resume")
	r.HandleFunc("/state", CurrentStateHandler)
	//r.HandleFunc("/time", GetTimeHandler).Name("time")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "favicon.ico") })
	r.NotFoundHandler = http.HandlerFunc(NotFound)

	http.Handle("/", r)
	http.ListenAndServe(":4000", r)
}
