package main

import (
	"github.com/gorilla/mux"
	"html/template"
	"net/http"
    "fmt"
    "io/ioutil"
    "crypto/tls"
    "time"
    "strconv"
)

var tr = &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}
var client = &http.Client{Transport: tr}

var timer_set_at time.Time = time.Now()
var timer_duration time.Duration = 0

func HomeHandler(
	w http.ResponseWriter,
	r *http.Request) {

	tmpl, err := template.New("root").ParseFiles("index.html")
	if err != nil { panic(err) }

    timer_expire := timer_set_at.Add( timer_duration )
    if time.Now().After( timer_set_at.Add(timer_duration) ) {
        err = tmpl.ExecuteTemplate(w, "root", nil)
        if err != nil { panic(err) }
    }else{
        err = tmpl.ExecuteTemplate(w, "root", timer_expire.Format(time.ANSIC))
        if err != nil { panic(err) }
    }
}

func ResumeHandler( w http.ResponseWriter, r *http.Request){
    resume_url :=  "https://localhost:9100/sabnzbd/api?mode=resume&apikey=d2ef95d20181d30d884321fb9cb68cbe"
    timer_set_at=time.Now()
    timer_duration =0
    resp, err := client.Get(resume_url)
    if err != nil{ panic(err) }
    defer resp.Body.Close()
    _, err = ioutil.ReadAll(resp.Body)
    if err != nil { panic(err) }
    http.Redirect(w,r, "/", 302)
}

func PauseHandler(w http.ResponseWriter, r *http.Request) {
    r.ParseForm() // get post data for extraction in r.FormValue
    timer_value, _ := strconv.ParseInt(r.FormValue("time"), 0, 32)
    timer_duration = time.Minute * time.Duration(timer_value)
    timer_set_at = time.Now()

    pause_url := fmt.Sprintf("https://localhost:9100/sabnzbd/api?mode=config&name=set_pause&value=%v&apikey=d2ef95d20181d30d884321fb9cb68cbe", timer_value)

    resp, err := client.Get(pause_url)
    if err != nil{ panic(err) }
    defer resp.Body.Close()
    _, err = ioutil.ReadAll(resp.Body)

    http.Redirect(w,r, "/", 302)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Name("home")
	r.HandleFunc("/pause", PauseHandler).Name("pause")
    r.HandleFunc("/resume", ResumeHandler).Name("resume")
    r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("js/"))))
    r.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir("img/"))))
    r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("css/"))))

	http.Handle("/", r)
	http.ListenAndServe(":4000", r)
}
