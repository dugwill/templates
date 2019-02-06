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
		Error.LogIt(fmt.Sprintf("Error initializing goLog: ", err))

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
	http.HandleFunc("/eventList", EventList)
	http.HandleFunc("/event", Event)
	http.HandleFunc("/streamList", StreamList)
	http.HandleFunc("/dateList", DateList)
	http.HandleFunc("/about", about)

	log.Fatalln(server.ListenAndServe())
}

func about(w http.ResponseWriter, r *http.Request) {

	t, _ := template.ParseFiles("template/about.html")
	if err := t.Execute(w, "about.html"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title  string
		Header string
	}{
		Title:  "AdAlignment Monitor",
		Header: "SCTE-35 Signal and Video Alignment Monitor",
	}

	t, _ := template.ParseFiles("template/index.html")
	if err := t.ExecuteTemplate(w, "index.html", data); err != nil {
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
				Error.LogIt(fmt.Sprintf("Error creating ffprofe results: %v", err))
			}

			//fmt.Printf("%s", respBytes)

			err = xml.Unmarshal(respBytes, &mpegFile)
			if err != nil {
				Error.LogIt(fmt.Sprintf("Unmarshal Error: %v", err))
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
				Error.LogIt(fmt.Sprintf("Error: %v", rErr))
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
		Error.LogIt(fmt.Sprintln("Error: %v", err))
	}

}

func readFiles(eventData *scte35.Event, dir string) (ts []string, jpegs bool, err error) {

	// if there is a .dat file Unmarshall XML data

	//dirSplt := strings.Split(dir, "/")
	datFile := dir + ".dat"

	Info.LogIt(fmt.Sprintf("DAT Filename: %v\n", datFile))

	b, err := ioutil.ReadFile(dir + ".dat")

	if err != nil {
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
