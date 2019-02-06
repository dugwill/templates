package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/Comcast/gots/scte35"
)

// Event Drive the display of a single event
func Event(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title  string
		Header string
		Event  scte35.Event
		Signal string
		Dir    string
		Thumbs []string
	}{
		Title:  "Event",
		Header: "Event Frames",
	}

	event, ok := r.URL.Query()["event"]
	if !ok || len(event[0]) < 1 {
		Error.LogIt("Url Param 'event' is missing")
		return
	}
	date, ok := r.URL.Query()["date"]
	if !ok || len(event[0]) < 1 {
		Error.LogIt("Url Param 'date' is missing")
		return
	}
	stream, ok := r.URL.Query()["stream"]

	if !ok || len(event[0]) < 1 {
		Error.LogIt("Url Param 'stream' is missing")
		return
	}

	Trace.LogIt(fmt.Sprintf("Url Param 'stream' is: " + stream[0]))
	Trace.LogIt(fmt.Sprintf("Url Param 'date' is: " + date[0]))
	Trace.LogIt(fmt.Sprintf("Url Param 'event' is: " + event[0]))

	data.Event.StreamName = stream[0]
	data.Title = date[0]

	dir := dir + stream[0] + "/" + date[0] + "/" + event[0]

	Trace.LogIt(fmt.Sprintf("%v", dir))

	var eventData scte35.Event
	ts, jpegs, err := readFiles(&eventData, dir)
	if err != nil {
		Error.LogIt(fmt.Sprintf("Error Reading Event File: %v", err))
	}

	data.Event = eventData

	if !jpegs {
		createJPEGs(&ts, &eventData, dir)
		jpegs = false
	}

	data.Signal = (strings.Split(string(data.Event.UPID), ":"))[1]
	data.Dir = dir

	t, _ := template.ParseFiles("template/event.html")
	if err := t.ExecuteTemplate(w, "event.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
