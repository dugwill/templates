package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Comcast/gots/scte35"
)

// Event Drive the display of a single event
func Event(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title            string
		Header           string
		Event            scte35.Event
		Signal           string
		Dir              string
		Thumbs           []string
		JPEGS            bool
		BJPEG            []string
		AJPEG            []string
		ThisEvent        string
		BFrames, AFrames uint64
	}{
		Title:  "Event",
		Header: "Event Frames",
	}

	var err error
	var iEventPTS uint64

	if r.Method == "GET" {
		data.BFrames = 5
		data.AFrames = 5
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
	data.ThisEvent = event[0]

	dir := dir + stream[0] + "/" + date[0] + "/" + event[0]

	Trace.LogIt(fmt.Sprintf("Event Handler Dir: %v", dir))

	var eventData scte35.Event
	_, data.JPEGS, err = readFiles(&eventData, dir)
	if err != nil {
		Error.LogIt(fmt.Sprintf("%v Error reading event metadata file: %v", data.Event.StreamName, err))
	}

	data.Event = eventData

	data.Signal = (strings.Split(string(data.Event.UPID), ":"))[1]
	data.Dir = dir

	if data.JPEGS {
		//Calc the clock ticks before and after the splice point
		//Each frame ~=1501 ticks
		iEventPTS = uint64(data.Event.PTS)
		timeBefore := data.BFrames * 1501
		timeAfter := data.AFrames * 1501
		//Calc the PTS of the JPEGs to to display
		firstJPG := iEventPTS - timeBefore
		lastJPG := iEventPTS + timeAfter
		Info.LogIt(fmt.Sprintf("Displaying frames from %v to %v", firstJPG, lastJPG))

		// Find jpgs within the PTS range
		list, _ := ioutil.ReadDir(data.Dir)
		//fmt.Println(list)
		for _, file := range list {
			Trace.LogIt(fmt.Sprintf("Name: " + file.Name()))

			if filepath.Ext(file.Name()) == ".jpg" {
				fileStr := (strings.Split(file.Name(), "."))[0]
				filePTS, _ := strconv.ParseUint(fileStr, 10, 64)
				if filePTS >= firstJPG && filePTS <= iEventPTS {
					data.BJPEG = append(data.BJPEG, file.Name())
				}
				if filePTS >= iEventPTS && filePTS <= lastJPG {
					data.AJPEG = append(data.AJPEG, file.Name())
				}
			}
		}
	}

	if r.Method == "GET" {
		t, _ := template.ParseFiles("template/event.html")
		if err := t.ExecuteTemplate(w, "event.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if r.Method == "POST" {

		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}

		Info.LogIt("Frames before: " + r.Form["bframes"][0])
		Info.LogIt("Frames after: " + r.Form["aframes"][0])

		data.BFrames, _ = strconv.ParseUint(r.Form["bframes"][0], 10, 64)
		data.AFrames, _ = strconv.ParseUint(r.Form["aframes"][0], 10, 64)

		timeBefore := data.BFrames * 1501
		timeAfter := data.AFrames * 1501
		//Calc the PTS of the JPEGs to to display
		firstJPG := iEventPTS - timeBefore
		lastJPG := iEventPTS + timeAfter
		Info.LogIt(fmt.Sprintf("Displaying frames from %v to %v", firstJPG, lastJPG))

		// Find jpgs within the PTS range
		list, _ := ioutil.ReadDir(data.Dir)
		//fmt.Println(list)
		for _, file := range list {
			Trace.LogIt(fmt.Sprintf("Name: " + file.Name()))

			if filepath.Ext(file.Name()) == ".jpg" {
				fileStr := (strings.Split(file.Name(), "."))[0]
				filePTS, _ := strconv.ParseUint(fileStr, 10, 64)
				if filePTS >= firstJPG && filePTS <= iEventPTS {
					data.BJPEG = append(data.BJPEG, file.Name())
				}
				if filePTS >= iEventPTS && filePTS <= lastJPG {
					data.AJPEG = append(data.AJPEG, file.Name())
				}
			}
		}

		t, _ := template.ParseFiles("template/event.html")
		if err := t.ExecuteTemplate(w, "event.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}
