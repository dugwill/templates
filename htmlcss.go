package main

import (
	"encoding/xml"
	"errors"
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

//var tmpls = template.Must(template.ParseFiles("templates/index.html"))
//var tmpls2 = template.Must(template.ParseFiles("templates/graphics.html"))

var tmpls, _ = template.ParseFiles("templates/index.html",
	"templates/graphics.html",
	"templates/eventList.html",
	"templates/event.html",
	"templates/event1.html",
	"templates/streamList.html")

func main() {

	server := http.Server{
		Addr: ":9000",
	}

	http.Handle("/jpegs/", http.StripPrefix("/jpegs/", http.FileServer(http.Dir("C:\\Users\\douglaswill\\goProjects\\src\\templates\\templates\\jpegs"))))
	http.Handle("/html/adAlign/", http.StripPrefix("/html/adAlign/", http.FileServer(http.Dir("C:\\html\\adAlign"))))
	http.HandleFunc("/", Index)
	http.HandleFunc("/graphics", Graphics)
	http.HandleFunc("/eventList", EventList)
	http.HandleFunc("/adAlign/event", Event)
	http.HandleFunc("/streamList", StreamList)

	log.Fatalln(server.ListenAndServe())
}

func Event(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title  string
		Header string
		Event  scte35.Event
		Thumbs []string
	}{
		Title:  "Event",
		Header: "Event Frames",
	}

	event, ok := r.URL.Query()["event"]

	if !ok || len(event[0]) < 1 {
		log.Println("Url Param event is missing")
		return
	}

	log.Println("Url Param 'event' is: " + event[0])
	data.Event = event[0]
	data.Title = event[0]

	dir := "/html/adAlign/" + event[0]

	var eventData scte35.Event
	ts, err := readFiles(&eventData, dir)
	if err != nil {
		fmt.Println(err)
	}

	// Send first event page

	data.Event = strconv.Itoa(int(eventData.EventID))
	data.Title = eventData.StreamName

	if err := tmpls.ExecuteTemplate(w, "event1.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createJPEGs(&ts, &eventData, dir)

	list, _ := ioutil.ReadDir("/html/adAlign/" + event[0]) // 0 to read all files and folders
	for _, file := range list {
		//fmt.Println("Name: " + file.Name())

		if filepath.Ext(file.Name()) == ".jpg" {
			data.Thumbs = append(data.Thumbs, file.Name())
		}
	}
	fmt.Println(data.Thumbs)

	if err := tmpls.ExecuteTemplate(w, "event.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func EventList(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title     string
		Header    string
		EventList []string
		Stream    string
	}{
		Title:  "Event List",
		Header: "SCTE-35 Events",
	}

	stream, ok := r.URL.Query()["stream"]

	if !ok || len(stream[0]) < 1 {
		log.Println("Url Param event is missing")
		return
	}

	log.Println("Url Param 'stream' is: " + stream[0])
	data.Title = "Stream: " + stream[0]
	data.Stream = stream[0]

	list, _ := ioutil.ReadDir("/html/adAlign/" + stream[0]) // 0 to read all files and folders
	for _, file := range list {
		fmt.Println("Name: " + file.Name())
		fmt.Printf("Dir?: %v\n", file.IsDir())

		if file.IsDir() {
			data.EventList = append(data.EventList, file.Name())
		}
	}
	fmt.Println(data.EventList)

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

	list, _ := ioutil.ReadDir("/html/adAlign/") // 0 to read all files and folders
	for _, file := range list {
		fmt.Println("Name: " + file.Name())
		fmt.Printf("Dir?: %v\n", file.IsDir())

		if file.IsDir() {
			data.StreamList = append(data.StreamList, file.Name())
		}
	}
	fmt.Println(data.StreamList)

	if err := tmpls.ExecuteTemplate(w, "streamList.html", data); err != nil {
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

	//data.Slice = []string{"thumb0001.jpg", "thumb0002.jpg", "thumb0003.jpg", "thumb0004.jpg", "thumb0005.jpg", "thumb0006.jpg", "thumb0007.jpg", "thumb0008.jpg", "thumb0009.jpg", "thumb0010.jpg", "thumb0011.jpg", "thumb0012.jpg", "thumb0013.jpg", "thumb0014.jpg", "thumb0015.jpg"}

	data.Slice, _ = filepath.Glob(".\\templates\\jpegs\\*.jpg")
	fmt.Println(len(data.Slice))
	fmt.Println(data.Slice)
	for f := range data.Slice {
		data.Slice[f] = filepath.Base(data.Slice[f])
	}

	fmt.Println(data.Slice)

	/*
		files, err := ioutil.ReadDir("C:\\Users\\douglaswill\\goProjects\\src\\templates\\templates\\jpegs")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			fmt.Println(f.Name())
		}
	*/

	fmt.Println("Serving Graphics")

	if err := tmpls.ExecuteTemplate(w, "graphics.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createJPEGs(ts *[]string, eventData *scte35.Event, dir string) {

	fmt.Printf("Here I will create some JPGs for your viewing pleasure.\n")

	//targetPTS, _ := strconv.ParseUint(eventData.PTS, 10, 64)

	// for the .ts files find the one with the time < .pts time from .dat

	for _, tsFile := range *ts {
		filePTS := extractPTS(tsFile)

		fmt.Println(filePTS)

		var diff uint64
		if filePTS > uint64(eventData.PTS) {
			diff = filePTS - uint64(eventData.PTS)
		} else {
			diff = uint64(eventData.PTS) - filePTS
		}

		if diff > 1501 {
			fmt.Printf("Here I will extract the last 10 frames.\n")

		}

		if diff < 1501 {
			fmt.Printf("Here I will extract the first 10 frames.\n")

			//ffmpeg -v error -i DSCHD_HD_NAT_16122_0_5965163366898931163_7145577515.ts -vf select='eq(n\,0)+eq(n\,1)+eq(n\,2)+eq(n\,3)+eq(n\,5)+eq(n\,6)+eq(n\,7)+eq(n\,8)+eq(n\,9)+eq(n\,10)' -vsync 0 7145577515_%02d.jpg

		}
	}

	// extract the last 10 frames as jpg

	// find the file with time > pts from .day

	// extract the first 10 frames a jpg

	return

}

func readFiles(eventData *scte35.Event, dir string) (ts []string, err error) {

	var dat []string
	// Read the file names from dir save them

	list, _ := ioutil.ReadDir(dir)
	for _, file := range list {
		fmt.Println("Name: " + file.Name())
		nameSplt := strings.Split(file.Name(), ".")

		// if .jpg files exist, return
		if nameSplt[len(nameSplt)-1] == "jpg" {
			fmt.Println(".jpg file found: " + file.Name() + "Returning")
			return ts, nil
		}

		if nameSplt[len(nameSplt)-1] == "dat" {
			fmt.Println(".dat file found: " + file.Name())
			dat = append(dat, file.Name())

		}

		if nameSplt[len(nameSplt)-1] == "ts" {
			fmt.Println(".ts file found: " + file.Name())
			ts = append(ts, file.Name())
		}

	}

	// if there is a .dat file Unmarshall XML data

	var b []byte

	if len(dat[0]) > 0 {
		b, err = ioutil.ReadFile(dir + "/" + dat[0])
		if err != nil {
			fmt.Print(err)
			err := errors.New("Error reading metadata file")
			return ts, err
		}

		xml.Unmarshal(b, &eventData)

		fmt.Println(eventData.StreamName)
		fmt.Println(eventData.EventID)
		fmt.Println(eventData.EventTime)
		fmt.Println(eventData.PTS)
		fmt.Println(eventData.Command)
		fmt.Println(eventData.TypeID)
		fmt.Println(string(eventData.UPID))
		fmt.Println(eventData.BreakDuration)
	} else {
		err := errors.New("no metatdata file found")
		return ts, err
	}

	return ts, nil

}

func extractPTS(file string) (filePTS uint64) {

	a := strings.Split(file, "_")
	b := strings.Split((a[len(a)-1]), ".")
	c := b[0]
	filePTS, _ = strconv.ParseUint(c, 10, 64)
	return
}
