package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

//var tmpls = template.Must(template.ParseFiles("templates/index.html"))
//var tmpls2 = template.Must(template.ParseFiles("templates/graphics.html"))

var tmpls, _ = template.ParseFiles("templates/index.html", "templates/graphics.html")

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

	data.Slice = []string{"bob", "joe", "frank", "pete", "doug"}

	fmt.Println("Serving Graphics")

	if err := tmpls.ExecuteTemplate(w, "graphics.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	server := http.Server{
		Addr: "127.0.0.1:9000",
	}

	http.HandleFunc("/", Index)
	http.HandleFunc("/graphics", Graphics)
	log.Fatalln(server.ListenAndServe())
}
