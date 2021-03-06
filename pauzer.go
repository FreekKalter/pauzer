// Copyright 2013 (c) Freek Kalter. All rights reserved.
// Use of this source code is governed by the "Revised BSD License"
// that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"launchpad.net/goyaml"
)

type Config struct {
	Api_key   string
	Api_url   string
	Max_speed int
	Port      int
	Times     []int
}

var slog *log.Logger
var config Config

type countDown struct {
	SetAt                  time.Time
	Duration               time.Duration
	Limit, LimitPercentage int64
	ReturnState            ReturnState
}
type ReturnState struct {
	Speedlimit string
	Paused     bool
}

var cDown countDown = countDown{
	SetAt:    time.Unix(0, 0),
	Duration: 0,
	Limit:    0,
}

var compiledTemplates = template.Must(template.ParseFiles("404.html"))

var sabNzbFunctions map[string]string

// ignore invalid certificates (todo: make it accecpt a valid cert)
var client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

func (c countDown) ExpiresAt() (expire time.Time, err error) {
	if c.SetAt.Equal(time.Unix(0, 0)) {
		err = errors.New("timer not running")
	}
	expire = c.SetAt.Add(c.Duration)
	return
}

func (c countDown) SecondsLeft() (secs int64, err error) {
	expires, err := c.ExpiresAt()
	if err != nil {
		err = errors.New("timer not running")
	}
	secs = int64(expires.Sub(time.Now()).Seconds())
	return
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	indexContent, err := ioutil.ReadFile("index.html")
	if err != nil {
		slog.Panic(err)
	}
	fmt.Fprintf(w, string(indexContent))
}

func resumeHandler(w http.ResponseWriter, r *http.Request) {
	cDown.Duration = 0
	cDown.Limit = 0
	resumeDownload(cDown.ReturnState)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	formVars := mux.Vars(r)
	valid_integer_regex := regexp.MustCompile("^[0-9]{1,3}$")
	if !valid_integer_regex.MatchString(strings.TrimSpace(formVars["time"])) ||
		!valid_integer_regex.MatchString(strings.TrimSpace(formVars["limit"])) {
		cDown.Duration = 0
		return // TODO: proper error message maybe
	}

	timer_value, _ := strconv.ParseInt(formVars["time"], 10, 32)           //base 10, 32bit integer
	cDown.LimitPercentage, _ = strconv.ParseInt(formVars["limit"], 10, 32) //base 10, 32bit integer
	cDown.Duration = time.Minute * time.Duration(timer_value)
	cDown.Limit = int64(config.Max_speed) - ((int64(config.Max_speed) / 100) * cDown.LimitPercentage) // percentage give is how much to block, so inverse that to get how much to let through
	cDown.SetAt = time.Now()
	cDown.ReturnState = getCurrentState()
	slog.Printf("timer started: %+v\n", cDown)
	time.AfterFunc(cDown.Duration, func() {
		slog.Printf("timer done: %+v\n", cDown)
		cDown.Duration = 0
		cDown.SetAt = time.Unix(0, 0)
		resumeDownload(cDown.ReturnState)
	})

	if cDown.LimitPercentage == 100 {
		go callSabnzbd(fmt.Sprintf(sabNzbFunctions["pause_time"], timer_value))
	} else {
		go callSabnzbd(fmt.Sprintf(sabNzbFunctions["limit"], cDown.Limit))
	}

}

func resumeDownload(cs ReturnState) {
	slog.Println("Resuming downloads")
	if cs.Paused {
		go callSabnzbd(sabNzbFunctions["pause"])
	} else {
		go callSabnzbd(sabNzbFunctions["resume_download"])
	}
	go callSabnzbd(fmt.Sprintf(sabNzbFunctions["limit"], cs.Speedlimit))
}

func getCurrentState() ReturnState {
	resp, err := client.Get(sabNzbFunctions["status"])
	if err != nil {
		slog.Panic(err)
	}
	defer resp.Body.Close()

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	type Queue struct {
		Queue ReturnState
	}
	var q Queue
	err = json.Unmarshal(text, &q)
	if err != nil {
		panic(err)
	}
	slog.Printf("%+v", q.Queue)
	return q.Queue
}

func currentStateHandler(w http.ResponseWriter, r *http.Request) {
	var limit, dur int64
	secs, err := cDown.SecondsLeft()
	if err != nil || secs <= 0 {
		limit, dur = 0, 0
	} else {
		dur = int64(cDown.Duration.Minutes())
		limit = cDown.LimitPercentage
	}
	state := map[string]interface{}{"secondsLeft": secs, "limit": limit, "time": dur, "times": config.Times}

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.Encode(state)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	err := compiledTemplates.ExecuteTemplate(w, "404.html", r.URL)
	if err != nil {
		slog.Panic(err)
	}

}

func callSabnzbd(url string) {
	resp, err := client.Get(url)
	if err != nil {
		slog.Panic(err)
	}
	defer resp.Body.Close()
}

func initSabnzbFunctions() {
	sabNzbFunctions = map[string]string{
		"reset_limit":     fmt.Sprintf("%sapi?mode=config&name=speedlimit&value=0&apikey=%v", config.Api_url, config.Api_key),
		"resume_download": fmt.Sprintf("%vapi?mode=resume&apikey=%v", config.Api_url, config.Api_key),
		"pause_time":      fmt.Sprintf("%vapi?mode=config&name=set_pause&value=%%v&apikey=%v", config.Api_url, config.Api_key),
		"pause":           fmt.Sprintf("%vapi?mode=pause&apikey=%v", config.Api_url, config.Api_key),
		"limit":           fmt.Sprintf("%vapi?mode=config&name=speedlimit&value=%%v&apikey=%v", config.Api_url, config.Api_key),
		"status":          fmt.Sprintf("%vapi?mode=queue&start=START&limit=LIMIT&apikey=%v&output=json", config.Api_url, config.Api_key),
	}
}

func cacheHandler(dur time.Duration, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", int64(dur.Seconds())))
		h.ServeHTTP(w, r)
	})
}

func main() {
	// Set up logging
	var err error
	slog = log.New(os.Stdout, "pauzer: ", log.LstdFlags)
	if err != nil {
		slog.Panic(err)
	}

	// Set up gracefull termination
	killChannel := make(chan os.Signal, 1)
	signal.Notify(killChannel, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal, l *log.Logger) {
		<-c
		l.Println("shutting down")
		os.Exit(0)
	}(killChannel, slog)

	// Load config
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		slog.Panic(err)
	}
	err = goyaml.Unmarshal(configFile, &config)
	if err != nil {
		slog.Panic(err)
	}
	initSabnzbFunctions()

	// set up gorilla/mux handlers
	r := mux.NewRouter()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/action/{time:[0-9]+}/{limit:[0-9]+}", formHandler)
	r.HandleFunc("/resume", resumeHandler)
	r.HandleFunc("/state", currentStateHandler)

	// static files get served directly
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", cacheHandler(time.Second*2678400, http.FileServer(http.Dir("js/")))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", cacheHandler(time.Second*2678400, http.FileServer(http.Dir("img/")))))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", cacheHandler(time.Second*2678400, http.FileServer(http.Dir("css/")))))
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", int64(time.Second*2678400)))
		http.ServeFile(w, r, "favicon.ico")
	})
	r.NotFoundHandler = http.HandlerFunc(notFound)

	http.Handle("/", r)
	slog.Println("started on port", config.Port)
	http.ListenAndServe(":4000", r)
}
