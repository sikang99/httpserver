/*
	https://www.socketloop.com/tutorials/golang-upload-file
*/
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v\n", r.URL)

	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	out, err := os.Create("/tmp/uploadedfile")
	if err != nil {
		fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
		return
	}

	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintf(w, "File '%s' uploaded successfully.\n", header.Filename)
}

func main() {
	// to serve goupload.html file in the browser
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/receive", uploadHandler)

	log.Printf("First, access to server at http://localhost:8080/static\n")
	http.ListenAndServe(":8080", nil)
}
