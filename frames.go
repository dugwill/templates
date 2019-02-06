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
)

func Frames(w http.ResponseWriter, r *http.Request) {

	data := struct {
		StreamName    string
		EventID       string
		EventPTS      string
		EventTypeID   string
		EventSignal   string
		EventDuration string
		Dir           string
		BJPEG         []string
		AJPEG         []string
	}{}

	//fmt.Println("method:", r.Method) //get request method

	//if r.Method == "GET" {
	//	fmt.Println("Frames Get")
	//} else {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	Info.LogIt("Form Data")
	Info.LogIt("StreamName: " + r.Form["StreamName"][0])
	Info.LogIt("EventID: " + r.Form["EventID"][0])
	Info.LogIt("EventPTS: " + r.Form["EventPTS"][0])
	Info.LogIt("EventTypeID: " + r.Form["EventTypeID"][0])
	Info.LogIt("EventSignal: " + r.Form["EventSignal"][0])
	Info.LogIt("EventDuration: " + r.Form["EventDuration"][0])
	Info.LogIt("Frames before: " + r.Form["bframes"][0])
	Info.LogIt("Frames after: " + r.Form["aframes"][0])
	Info.LogIt("JPEG directory: " + r.Form["Dir"][0])

	data.StreamName = r.Form["StreamName"][0]
	data.EventID = r.Form["EventID"][0]
	data.EventPTS = r.Form["EventPTS"][0]
	data.EventTypeID = r.Form["EventTypeID"][0]
	data.EventSignal = r.Form["EventSignal"][0]
	data.EventDuration = r.Form["EventDuration"][0]
	data.Dir = r.Form["Dir"][0]

	//Calc the clock ticks before and after the splice point
	//Each frame ~=1501 ticks
	iBframes, _ := strconv.ParseUint(r.Form["bframes"][0], 10, 64)
	iAframes, _ := strconv.ParseUint(r.Form["aframes"][0], 10, 64)
	iEventPTS, _ := strconv.ParseUint(r.Form["EventPTS"][0], 10, 64)
	timeBefore := iBframes * 1501
	timeAfter := iAframes * 1501
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
	//}

	t, _ := template.ParseFiles("template/frames.html")
	if err := t.ExecuteTemplate(w, "frames.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
