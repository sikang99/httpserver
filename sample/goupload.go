/*
	https://www.socketloop.com/tutorials/golang-upload-file
	http://stackoverflow.com/questions/28940005/golang-get-multipart-form-data
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var port = flag.String("port", "8080", "server port")
var dir = flag.String("dir", "./static/", "directory to handle files")

func init() {
	log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v multipart\n", r.URL)

	err := r.ParseMultipartForm(0)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	// the FormFile function takes in the POST input id file
	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	out, err := os.Create(*dir + header.Filename)
	if err != nil {
		log.Printf("Unable to create the file for writing. Check your write privilege")
		return
	}
	defer out.Close()

	// write the content from POST to the file
	_, err = io.Copy(out, file)
	if err != nil {
		log.Println(w, err)
	}

	log.Printf("File '%s' uploaded successfully.\n", header.Filename)

	// remove the file
	if r.FormValue("delete") != "" {
		err = os.Remove(*dir + header.Filename)
		if err != nil {
			log.Printf("Unable to delete the file. Check your remove privilege")
			return
		}
		log.Printf("File '%s' deleted successfully.\n", header.Filename)
	}

	//fmt.Printf("check at http://localhost:%s/static\n", *port)
	w.Header().Add("Content-Type:", "text/html")
	body, _ := ioutil.ReadFile("static/gostatic.html")
	fmt.Fprint(w, string(body))
}

func main() {
	// to serve goupload.html file in the browser
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/receive", uploadHandler)

	fmt.Printf("visit with http://localhost:%s/static\n", *port)
	http.ListenAndServe(":"+*port, nil)
}
