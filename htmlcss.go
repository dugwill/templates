package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
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

func main() {

	server := http.Server{
		Addr: ":9000",
	}

	http.Handle("/jpegs/", http.StripPrefix("/jpegs/", http.FileServer(http.Dir("C:\\Users\\douglaswill\\goProjects\\src\\templates\\templates\\jpegs"))))
	http.HandleFunc("/", Index)
	http.HandleFunc("/graphics", Graphics)

	log.Fatalln(server.ListenAndServe())
}
