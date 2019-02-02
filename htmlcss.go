package main

import (
	"encoding/xml"
	"errors"
	"ffMpegOutput"
	"flag"
	"fmt"
	"golog"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Comcast/gots/scte35"
)

var tmpls, _ = template.ParseFiles(
	"template/index.html",
	"template/graphics.html",
	"template/eventList.html",
	"template/event.html",
	"template/event1.html",
	"template/dateList.html",
	"template/streamList.html",
	"template/newEventHandle.html",
	"template/frames.html")

var dir = "/app/html/AdAlign/"

var Info goLog.Info
var Warning goLog.Warning
var Error goLog.Error
var Trace goLog.Trace
var logPointer goLog.Pointers

func main() {
	var err error

	//********** Parse Flags **********
	l := flag.String("t", "", "Log Level: t for trace")
	d := flag.String("d", "/app/html/AdAlign/", "Base directory for the file system")
	p := flag.Int("p", 9000, "Listenr port")

	flag.Parse()

	dir = *d

	//Opent log file

	logFile := "/var/log/scte35web.log"

	_, err = goLog.Initialize(logFile, *l)
	if err != nil {
		fmt.Println("Error initializing goLog: ", err)

		logPointer.ILog = Info
		logPointer.TLog = Trace
		logPointer.ELog = Error
		logPointer.WLog = Warning
	}

	server := http.Server{
		Addr: ":" + fmt.Sprintf("%d", *p),
	}

	http.Handle(dir, http.StripPrefix(dir, http.FileServer(http.Dir(dir))))
	http.Handle("/AdAlign/", http.StripPrefix("/AdAlign/", http.FileServer(http.Dir("/html/AdAlign"))))
	http.HandleFunc("/", Index)
	http.HandleFunc("/graphics", Graphics)
	http.HandleFunc("/eventList", EventList)
	http.HandleFunc("/adAlign/event", Event)
	http.HandleFunc("/streamList", StreamList)
	http.HandleFunc("/dateList", DateList)
	http.HandleFunc("/adAlign/neweventhandle", NewEventHandle)
	http.HandleFunc("/adAlign/frames", Frames)

	log.Fatalln(server.ListenAndServe())
}

func NewEventHandle(w http.ResponseWriter, r *http.Request) {

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
		fmt.Println(err)
	}

	data.Event = eventData

	if !jpegs {
		createJPEGs(&ts, &eventData, dir)
		jpegs = false
	}

	data.Signal = (strings.Split(string(data.Event.UPID), ":"))[1]
	data.Dir = dir

	if err := tmpls.ExecuteTemplate(w, "newEventHandle.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func Frames(w http.ResponseWriter, r *http.Request) {
	data := struct {
		StreamName    string
		EventID       string
		EventPTS      string
		EventTypeID   string
		EventSignal   string
		EventDuration string
		Dir           string
		Thumbs        []string
	}{}

	streamName, ok := r.URL.Query()["StreamName"]
	if !ok || len(streamName) < 1 {
		Error.LogIt("Url Param 'StreamName' is missing")
		return
	}
	data.StreamName = streamName[0]
	fmt.Println(data.StreamName)

	eventID, ok := r.URL.Query()["EventID"]
	if !ok || len(eventID) < 1 {
		Error.LogIt("Url Param 'EventID' is missing")
		return
	}
	data.EventID = eventID[0]
	fmt.Println(data.EventID)

	eventPTS, ok := r.URL.Query()["EventPTS"]
	if !ok || len(eventPTS) < 1 {
		Error.LogIt("Url Param 'EventPTS' is missing")
		return
	}
	data.EventPTS = eventPTS[0]
	fmt.Println(data.EventPTS)

	eventTypeId, ok := r.URL.Query()["EventTypeID"]
	if !ok || len(eventTypeId[0]) < 1 {
		Error.LogIt("Url Param 'EventTypeID' is missing")
		return
	}
	data.EventTypeID = eventTypeId[0]
	fmt.Println(data.EventTypeID)

	eventSignal, ok := r.URL.Query()["EventSignal"]
	if !ok || len(eventSignal) < 1 {
		Error.LogIt("Url Param 'EventSignal' is missing")
		return
	}
	data.EventSignal = eventSignal[0]
	fmt.Println(data.EventSignal)

	eventDuration, ok := r.URL.Query()["EventDuration"]
	if !ok || len(eventDuration) < 1 {
		Error.LogIt("Url Param 'EventDuration' is missing")
		return
	}
	data.EventDuration = eventDuration[0]
	fmt.Println(data.EventDuration)

	dir, ok := r.URL.Query()["dir"]
	if !ok || len(dir) < 1 {
		Error.LogIt("Url Param 'dir' is missing")
		return
	}
	data.Dir = dir[0]
	fmt.Println(data.Dir)

	/*
		if err := tmpls.ExecuteTemplate(w, "frames.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	*/

}

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
		fmt.Println(err)
	}

	data.Event = eventData

	if !jpegs {
		createJPEGs(&ts, &eventData, dir)
		jpegs = false
	}

	//time.Sleep(5 * time.Second)

	list, _ := ioutil.ReadDir(dir) // 0 to read all files and folders
	for _, file := range list {

		if filepath.Ext(file.Name()) == ".jpg" {
			data.Thumbs = append(data.Thumbs, file.Name())
		}
	}
	fmt.Println(data.Thumbs)
	data.Signal = (strings.Split(string(data.Event.UPID), ":"))[1]
	data.Dir = dir

	if err := tmpls.ExecuteTemplate(w, "event.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

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
				fmt.Print(err)
				continue
			}
			var tempEvent scte35.Event
			var elEntry EventList

			a := strings.Split(file.Name(), ".")
			elEntry.EventFile = a[0] + "." + a[1]
			fmt.Println(elEntry.EventFile)

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

	if err := tmpls.ExecuteTemplate(w, "eventList.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func StreamList(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title      string
		Header     string
		StreamList []string
	}{
		Title:  "Stream List",
		Header: "Streams",
	}

	list, _ := ioutil.ReadDir(dir) // 0 to read all files and folders
	for _, file := range list {
		Trace.LogIt(fmt.Sprintln("Name: " + file.Name()))
		Trace.LogIt(fmt.Sprintf("Dir?: %v\n", file.IsDir()))

		if file.IsDir() {
			data.StreamList = append(data.StreamList, file.Name())
		}
	}
	Trace.LogIt(fmt.Sprintf("%v", data.StreamList))

	if err := tmpls.ExecuteTemplate(w, "streamList.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func DateList(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title    string
		Header   string
		DateList []string
		Stream   string
	}{
		Title:  "Days",
		Header: "Choose a Day",
	}

	stream, ok := r.URL.Query()["stream"]

	if !ok || len(stream[0]) < 1 {
		Error.LogIt("Url Param event is missing")
		return
	}

	log.Println("Url Param 'stream' is: " + stream[0])
	data.Title = "Stream: " + stream[0]
	data.Stream = stream[0]

	list, _ := ioutil.ReadDir(dir + stream[0]) // 0 to read all files and folders
	for _, file := range list {
		Info.LogIt(fmt.Sprintf("Name: " + file.Name()))
		Info.LogIt(fmt.Sprintf("Dir?: %v\n", file.IsDir()))

		if file.IsDir() {
			data.DateList = append(data.DateList, file.Name())
		}
	}
	Trace.LogIt(fmt.Sprintf("%v", data.DateList))

	if err := tmpls.ExecuteTemplate(w, "dateList.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title  string
		Header string
	}{
		Title:  "Index Page",
		Header: "Hello, World!",
	}

	if err := tmpls.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Graphics(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title  string
		Header string
		Slice  []string
	}{
		Title:  "Graphics Page",
		Header: "Here are some Graphics!",
	}

	fmt.Println(len(data.Slice))
	fmt.Println(data.Slice)
	for f := range data.Slice {
		data.Slice[f] = filepath.Base(data.Slice[f])
	}

	fmt.Println(data.Slice)

	fmt.Println("Serving Graphics")

	if err := tmpls.ExecuteTemplate(w, "graphics.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createJPEGs(ts *[]string, eventData *scte35.Event, dir string) {

	numJPEGS := 10

	Info.LogIt(fmt.Sprintf("TS Files to process: %v\n", ts))

	// for the .ts files find the one with the time < .pts time from .dat

	for _, tsFile := range *ts {
		filePTS := extractPTS(tsFile)
		targetPTS := uint64(eventData.PTS)

		Info.LogIt(fmt.Sprintf("Target PTS:= %v\n", targetPTS))
		Info.LogIt(fmt.Sprintf("File PTS:=%v\n", filePTS))

		var diff uint64
		if filePTS > targetPTS {
			diff = filePTS - targetPTS
		} else {
			diff = uint64(targetPTS) - filePTS
		}

		if diff >= 750 {

			var mpegFile ffmpegOutput.FFprobe
			inputFile := dir + "/" + tsFile

			// Read number of frames in .ts file
			c := exec.Command(`ffprobe`, `-v`, `error`,
				`-show_entries`, `stream=duration,nb_read_frames`,
				`-count_frames`, `-of`, `xml`, inputFile)

			respBytes, err := c.CombinedOutput()
			if err != nil {
				fmt.Println("Error: ", err)
			}

			//fmt.Printf("%s", respBytes)

			err = xml.Unmarshal(respBytes, &mpegFile)
			if err != nil {
				fmt.Println("Unmarshal Error: ", err)
			}

			//fmt.Println(mpegFile.Streams[0].Stream[0].Duration)
			//fmt.Println(mpegFile.Streams[0].Stream[0].NbReadFrames)

			// extract the last 10 frames as jpg
			startFrame, _ := strconv.ParseInt(mpegFile.Streams[0].Stream[0].NbReadFrames, 10, 32)
			startFrame = startFrame - int64(numJPEGS)

			before := true

			extractJPGS(numJPEGS, int(startFrame), dir, tsFile, before)

			// Get frame data
			c = exec.Command(`ffprobe`, `-v`, `error`,
				`-show_frames`, `-of`, `xml`, inputFile)

			Info.LogIt("Getting Frame Data from before file")
			//fmt.Println(c)
			rBytes, rErr := c.CombinedOutput()
			if rErr != nil {
				fmt.Println("Error: ", rErr)
			}

			err = xml.Unmarshal(rBytes, &mpegFile)
			if err != nil {
				Error.LogIt(fmt.Sprintf("Unmarshal Error: %v", err))
			}

			for i := 0; i < len(mpegFile.Frames[0].Frame); i++ {
				Trace.LogIt(fmt.Sprintf("CPB %v, DPNum %v PTS %v DTS %v\n",
					mpegFile.Frames[0].Frame[i].CodedPictureNumber,
					mpegFile.Frames[0].Frame[i].DisplayPictureNumber,
					mpegFile.Frames[0].Frame[i].PktPts,
					mpegFile.Frames[0].Frame[i].PktDts))
			}

			for i := 1; i <= 10; i++ {

				oldName := dir + "/" + "before_" + fmt.Sprintf("%d", i) + ".jpg"
				newName := dir + "/" + mpegFile.Frames[0].Frame[startFrame+int64(i-1)].PktPts + ".jpg"

				Trace.LogIt(fmt.Sprintf("OldName: %v  NewName: %v\n", oldName, newName))

				os.Rename(oldName, newName)

			}
		}

		if diff < 750 {
			Trace.LogIt("Here I will extract the first 10 frames.")

			startFrame := 0
			before := false

			var mpegFile ffmpegOutput.FFprobe
			inputFile := dir + "/" + tsFile

			extractJPGS(numJPEGS, startFrame, dir, tsFile, before)

			// Get frame data
			c := exec.Command(`ffprobe`, `-v`, `error`,
				`-show_frames`, `-of`, `xml`, inputFile)

			Trace.LogIt("Getting Frame Data from before file")
			//fmt.Println(c)
			rBytes, rErr := c.CombinedOutput()
			if rErr != nil {
				Trace.LogIt(fmt.Sprintln("Error: ", rErr))
			}

			err := xml.Unmarshal(rBytes, &mpegFile)
			if err != nil {
				fmt.Println("Unmarshal Error: ", err)
			}

			for i := 0; i < len(mpegFile.Frames[0].Frame); i++ {
				Trace.LogIt(fmt.Sprintf("CPB %v, DPNum %v PTS %v DTS %v\n",
					mpegFile.Frames[0].Frame[i].CodedPictureNumber,
					mpegFile.Frames[0].Frame[i].DisplayPictureNumber,
					mpegFile.Frames[0].Frame[i].PktPts,
					mpegFile.Frames[0].Frame[i].PktDts))
			}

			for i := 1; i <= 10; i++ {

				oldName := dir + "/" + "after_" + fmt.Sprintf("%d", i) + ".jpg"
				newName := dir + "/" + mpegFile.Frames[0].Frame[startFrame+(i-1)].PktPts + ".jpg"

				Trace.LogIt(fmt.Sprintf("OldName: %v  NewName: %v\n", oldName, newName))

				os.Rename(oldName, newName)
			}

		}
	}

	// find the file with time > pts from .day

	// extract the first 10 frames a jpg

	return

}

func extractJPGS(numJPEGS, startFrame int, dir, fileName string, before bool) {

	Info.LogIt(fmt.Sprintf("Extracting frames from %s: %d - %d\n", fileName, startFrame, startFrame+numJPEGS-1))

	frames := "select='"
	for i := startFrame; i < startFrame+numJPEGS; i++ {
		if i == startFrame {
			frames = frames + "eq(n\\,"
		} else {
			frames = frames + "+eq(n\\,"
		}
		frames = frames + strconv.Itoa(i) + ")"
	}
	frames = frames + "'"

	Info.LogIt(fmt.Sprintf("%v", frames))
	outputDir := ""
	fullPath := dir + "/" + fileName
	if before {
		outputDir = dir + "/" + "before_%01d.jpg"
	} else {
		outputDir = dir + "/" + "after_%01d.jpg"
	}

	c := exec.Command(`ffmpeg`, `-v`, `error`, `-i`, fullPath, `-vf`, frames, `-vsync`, `0`, outputDir)

	//fmt.Println(c)

	err := c.Run()
	if err != nil {
		fmt.Println("Error: ", err)
	}

}

func readFiles(eventData *scte35.Event, dir string) (ts []string, jpegs bool, err error) {

	// if there is a .dat file Unmarshall XML data

	//dirSplt := strings.Split(dir, "/")
	datFile := dir + ".dat"

	Info.LogIt(fmt.Sprintf("DAT Filename: %v\n", datFile))

	b, err := ioutil.ReadFile(dir + ".dat")

	if err != nil {
		fmt.Print(err)
		err := errors.New("Error reading metadata file")
		return ts, jpegs, err
	}

	xml.Unmarshal(b, &eventData)

	Trace.LogIt("DAT Data")
	Trace.LogIt(fmt.Sprintf("%v", eventData.StreamName))
	Trace.LogIt(fmt.Sprintf("%v", eventData.EventID))
	Trace.LogIt(fmt.Sprintf("%v", eventData.EventTime))
	Trace.LogIt(fmt.Sprintf("%v", eventData.PTS))
	Trace.LogIt(fmt.Sprintf("%v", eventData.Command))
	Trace.LogIt(fmt.Sprintf("%v", eventData.TypeID))
	Trace.LogIt(fmt.Sprintf(string(eventData.UPID)))
	Trace.LogIt(fmt.Sprintf("%v", eventData.BreakDuration))

	// Read the file names from dir save them

	list, _ := ioutil.ReadDir(dir + "/")
	for _, file := range list {
		Trace.LogIt(fmt.Sprintln("Name: " + file.Name()))
		//nameSplt := strings.Split(file.Name(), ".")

		// if .jpg files exist, return
		if filepath.Ext(file.Name()) == ".jpg" {
			Trace.LogIt(fmt.Sprintln(".jpg file found: " + file.Name() + "Returning"))
			jpegs = true
			return ts, jpegs, nil
		}

		if filepath.Ext(file.Name()) == ".ts" {
			Trace.LogIt(fmt.Sprintln(".ts file found: " + file.Name()))
			ts = append(ts, file.Name())
		}

	}

	return ts, jpegs, nil

}

func extractPTS(file string) (filePTS uint64) {

	a := strings.Split(file, "_")
	b := strings.Split((a[len(a)-1]), ".")
	c := b[0]

	filePTS, _ = strconv.ParseUint(c, 10, 64)

	Trace.LogIt(fmt.Sprintf("Extract PTS from Filename:  %v  PTS: %v\n", file, filePTS))

	return filePTS
}
