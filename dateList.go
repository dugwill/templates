package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

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

	t, _ := template.ParseFiles("template/dateList.html")
	if err := t.ExecuteTemplate(w, "dateList.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
