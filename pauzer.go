package main

import (
	"crypto/tls"
	"encoding/json"
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

const api_key = "d2ef95d20181d30d884321fb9cb68cbe"
const api_url = "https://localhost:9100/sabnzbd/"

// ignore invalid certificates (todo: make it accecpt a valid cert)
var tr = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var client = &http.Client{Transport: tr}

var timer_set_at time.Time = time.Now()
var timer_duration time.Duration = 0

type TemplateData struct {
	Time  string
	Error string
}

func HomeHandler(
	w http.ResponseWriter,
	r *http.Request) {

	tmpl, err := template.New("root").ParseFiles("index.tmpl")
	if err != nil {
		panic(err)
	}

	_, err = client.Get("http://localhost:4000/time")
	if err != nil {
		panic(err)
	}
	timer_expire := timer_set_at.Add(timer_duration)
	template_data := TemplateData{}
	if timer_duration == -1 {
		template_data.Error = "invalid time given"
	}

	if time.Now().After(timer_set_at.Add(timer_duration)) {
		err = tmpl.ExecuteTemplate(w, "root", template_data)
	} else {
		template_data.Time = timer_expire.Format(time.ANSIC)
		err = tmpl.ExecuteTemplate(w, "root", template_data)
	}
	if err != nil {
		panic(err)
	}
}

func ResumeHandler(w http.ResponseWriter, r *http.Request) {
	resume_url := fmt.Sprintf("%vapi?mode=resume&apikey=%v", api_url, api_key)
	timer_set_at = time.Now()
	timer_duration = 0
	call_sabnzbd(resume_url)
	http.Redirect(w, r, "/", 303)
}

func PauseHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // get post data for extraction in r.FormValue
	time_regex := regexp.MustCompile("^[0-9]{1,2}$")
	if !time_regex.MatchString(strings.TrimSpace(r.FormValue("time"))) {
		timer_duration = -1
		fmt.Println("invalid data")
	} else {
		timer_value, _ := strconv.ParseInt(r.FormValue("time"), 10, 32)
		timer_duration = time.Minute * time.Duration(timer_value)
		timer_set_at = time.Now()

		pause_url := fmt.Sprintf("%vapi?mode=config&name=set_pause&value=%v&apikey=%v", api_url, timer_value, api_key)
		go call_sabnzbd(pause_url)
	}
	http.Redirect(w, r, "/", 303)
}

func GetTimeHandler(
	w http.ResponseWriter,
	r *http.Request) {

	time_url := fmt.Sprintf("%vapi?mode=qstatus&output=json&apikey=%v", api_url, api_key)
	resp, err := client.Get(time_url)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	type DecodedContent struct {
		Pause_int string
	}
	var m DecodedContent
	err = json.Unmarshal(body, &m)
	if err != nil {
		panic(err)
	}

	time_regex := regexp.MustCompile("[0-9]{1,3}")
	time_array := time_regex.FindAllStringSubmatch(m.Pause_int, 2)
	if len(time_array) > 1 {

		minutes, err := strconv.ParseInt(time_array[0][0], 10, 32)
		if err != nil {
			panic(err)
		}
		seconds, err := strconv.ParseInt(time_array[1][0], 10, 32)
		if err != nil {
			panic(err)
		}
		timer_duration = time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
		timer_set_at = time.Now()
		timer_expire := timer_set_at.Add(timer_duration)

		fmt.Fprint(w, timer_expire.Format(time.ANSIC))
	} else {
		fmt.Fprintln(w, time.Now().Format(time.ANSIC))
		timer_duration = 0
	}
}

func NotFound(
	w http.ResponseWriter,
	r *http.Request) {

	tmpl, err := template.New("main").ParseFiles("404.tmpl")
	if err != nil {
		panic(err)
	}
	err = tmpl.ExecuteTemplate(w, "main", r.URL)
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
	r.HandleFunc("/", HomeHandler).Name("home")
	r.HandleFunc("/pause", PauseHandler).Name("pause")
	r.HandleFunc("/resume", ResumeHandler).Name("resume")
	r.HandleFunc("/time", GetTimeHandler).Name("time")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))
    r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request){ http.ServeFile(w,r,"favicon.ico") } )
	r.NotFoundHandler = http.HandlerFunc(NotFound)

	http.Handle("/", r)
	http.ListenAndServe(":4000", r)
}
