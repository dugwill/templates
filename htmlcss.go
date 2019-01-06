package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

//var tmpls = template.Must(template.ParseFiles("templates/index.html"))
//var tmpls2 = template.Must(template.ParseFiles("templates/graphics.html"))

var tmpls, _ = template.ParseFiles("templates/index.html",
	"templates/graphics.html",
	"templates/eventList.html",
	"templates/event.html",
	"templates/event2.html")

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

	log.Fatalln(server.ListenAndServe())
}

func Event(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title  string
		Header string
		Event  string
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
	}{
		Title:  "Event List",
		Header: "SCTE-35 Events",
	}

	list, _ := ioutil.ReadDir("/html/adAlign/") // 0 to read all files and folders
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
