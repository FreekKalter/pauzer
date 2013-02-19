package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
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
	SetAt    time.Time
	Duration time.Duration
	Limit    int64
}

func (c CountDown) ExpiresAt() time.Time {
	return c.SetAt.Add(c.Duration)
}

func (c CountDown) ExpiresAtJs() string {
	return (c.ExpiresAt()).Format(time.ANSIC)
}

var countDown CountDown = CountDown{
	SetAt:    time.Now(),
	Duration: -1,
	Limit:    0,
}

type TemplateData struct {
	Time  string
	Error string
}

var compiledTemplates = template.Must(template.ParseFiles("index.tmpl", "404.tmpl"))

var sabNzbFunctions map[string]string = map[string]string{
	"reset_limit":     fmt.Sprintf("%vapi?mode=config&name=speedlimit&value=0&apikey=%v", api_url, api_key),
	"resume_download": fmt.Sprintf("%vapi?mode=resume&apikey=%v", api_url, api_key),
	"pause":           fmt.Sprintf("%vapi?mode=config&name=set_pause&value=%%v&apikey=%v", api_url, api_key),
	"limit":           fmt.Sprintf("%vapi?mode=config&name=speedlimit&value=%%v&apikey=%v", api_url, api_key),
}

func HomeHandler(
	w http.ResponseWriter,
	r *http.Request) {

	tmplData := TemplateData{}
	if (countDown.ExpiresAt()).After(time.Now()) {
		tmplData.Time = countDown.ExpiresAtJs()
	}

	err := compiledTemplates.ExecuteTemplate(w, "index.tmpl", tmplData)
	if err != nil {
		panic(err)
	}
}

func ResumeHandler(w http.ResponseWriter, r *http.Request) {
	countDown.Duration = -1
	call_sabnzbd(sabNzbFunctions["resume_download"])
	call_sabnzbd(sabNzbFunctions["reset_limit"])
	http.Redirect(w, r, "/", 303)
}

func FormHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // get post data for extraction in r.FormValue
	valid_integer_regex := regexp.MustCompile("^[0-9]{1,3}$")
	if !valid_integer_regex.MatchString(strings.TrimSpace(r.FormValue("time"))) ||
		!valid_integer_regex.MatchString(strings.TrimSpace(r.FormValue("limit"))) {
		countDown.Duration = -1

		tmplData := TemplateData{Error: "invalid data"}
		err := compiledTemplates.ExecuteTemplate(w, "index.tmpl", tmplData)
		if err != nil {
			panic(err)
		}
		return
	} else {
		timer_value, _ := strconv.ParseInt(r.FormValue("time"), 10, 32)       //base 10, 32bit integer
		limit_percentage, _ := strconv.ParseInt(r.FormValue("limit"), 10, 32) //base 10, 32bit integer
		countDown.Duration = time.Minute * time.Duration(timer_value)
		countDown.Limit = max_speed - ((max_speed / 100) * limit_percentage) // percentage give is how much to block, so inverse that to get how much to let through
		time.AfterFunc(countDown.Duration, func() {
			countDown.Duration = -1
			call_sabnzbd(sabNzbFunctions["resume_download"])
			call_sabnzbd(sabNzbFunctions["reset_limit"])
		})

		if limit_percentage == 100 {
			go call_sabnzbd(fmt.Sprintf(sabNzbFunctions["pause"], timer_value))
		} else {
			go call_sabnzbd(fmt.Sprintf(sabNzbFunctions["limit"], countDown.Limit))
		}
		countDown.SetAt = time.Now()
	}
	http.Redirect(w, r, "/", 303)
}

func NotFound(
	w http.ResponseWriter,
	r *http.Request) {

	err := compiledTemplates.ExecuteTemplate(w, "404.tmpl", r.URL)
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
	r.HandleFunc("/action", FormHandler).Name("pause")
	r.HandleFunc("/resume", ResumeHandler).Name("resume")
	//r.HandleFunc("/time", GetTimeHandler).Name("time")
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
	r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "favicon.ico") })
	r.NotFoundHandler = http.HandlerFunc(NotFound)

	http.Handle("/", r)
	http.ListenAndServe(":4000", r)
}
