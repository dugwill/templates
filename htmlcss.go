package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var tmpls = template.Must(template.ParseFiles("templates/index.html"))

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

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", Index)

	r.PathPrefix("/styles/").Handler(http.StripPrefix("/styles/",
		http.FileServer(http.Dir("templates/styles/"))))

	http.Handle("/", r)
	log.Fatalln(http.ListenAndServe(":9000", nil))
}
