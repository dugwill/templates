package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Comcast/gots/scte35"
)

func EventList(w http.ResponseWriter, r *http.Request) {

	type EventList struct {
		EventFile string
		EventID   uint32
		TypeID    scte35.SegDescType
		UPID      string
		Duration  uint64
	}

	data := struct {
		Title     string
		Header    string
		EventList []EventList
		Stream    string
		Date      string
	}{
		Title:  "Event List",
		Header: "SCTE-35 Events",
	}

	stream, ok := r.URL.Query()["stream"]

	if !ok || len(stream[0]) < 1 {
		Error.LogIt("Url Param 'stream' is missing")
		return
	}

	date, ok := r.URL.Query()["date"]

	if !ok || len(stream[0]) < 1 {
		Error.LogIt("Url Param 'date' is missing")
		return
	}

	Trace.LogIt("Url Param 'stream' is: " + stream[0])
	Trace.LogIt("Url Param 'date' is: " + date[0])
	data.Stream = stream[0]
	data.Date = date[0]

	fileList, _ := ioutil.ReadDir(dir + stream[0] + "/" + date[0]) // 0 to read all files and folders
	for _, file := range fileList {
		Trace.LogIt(fmt.Sprintf("Name: " + file.Name()))
		//fmt.Printf("Dir?: %v\n", file.IsDir())
		if filepath.Ext(file.Name()) == ".dat" {
			Trace.LogIt(fmt.Sprintf("Processing DAT File: %v\n", file.Name()))

			b, err := ioutil.ReadFile(dir + stream[0] + "/" + date[0] + "/" + file.Name())

			if err != nil {
				Error.LogIt(fmt.Sprintf("%v", err))
				continue
			}
			var tempEvent scte35.Event
			var elEntry EventList //For setting up eventlist

			a := strings.Split(file.Name(), ".")
			elEntry.EventFile = a[0] + "." + a[1]
			Info.LogIt(fmt.Sprintf("Processing DAT file: %v", elEntry.EventFile))

			xml.Unmarshal(b, &tempEvent)

			Trace.LogIt("DAT Data")
			Trace.LogIt(fmt.Sprintf("%v", tempEvent.StreamName))
			Trace.LogIt(fmt.Sprintf("%v", tempEvent.EventID))
			elEntry.EventID = tempEvent.EventID
			Trace.LogIt(fmt.Sprintf("%v", tempEvent.EventTime))
			Trace.LogIt(fmt.Sprintf("%v", tempEvent.PTS))
			Trace.LogIt(fmt.Sprintf("%v", tempEvent.Command))
			Trace.LogIt(fmt.Sprintf("%v", tempEvent.TypeID))
			elEntry.TypeID = tempEvent.TypeID
			Trace.LogIt(fmt.Sprintf("%v", (string(tempEvent.UPID))))
			elEntry.UPID = (strings.Split(string(tempEvent.UPID), ":"))[1]
			Trace.LogIt(fmt.Sprintf("%v", tempEvent.BreakDuration))
			elEntry.Duration = uint64(tempEvent.BreakDuration) / 90000

			data.EventList = append(data.EventList, elEntry)
		}
	}
	//fmt.Println(data.EventList)

	//Process .dat files
	t, _ := template.ParseFiles("template/eventList.html")
	if err := t.ExecuteTemplate(w, "eventList.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
