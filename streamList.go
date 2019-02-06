package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

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

	t, _ := template.ParseFiles("template/streamList.html")
	if err := t.ExecuteTemplate(w, "streamList.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
