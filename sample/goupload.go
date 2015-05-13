/*
	https://www.socketloop.com/tutorials/golang-upload-file
	http://stackoverflow.com/questions/28940005/golang-get-multipart-form-data
*/
package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	port    = flag.String("port", "8080", "server port")
	dir     = flag.String("dir", "./static/", "directory to handle files")
	version = flag.String("version", "0.0.9", "program version")
)

func init() {
	log.SetOutput(os.Stdout)
	//log.SetPrefix("TRACE: ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flag.Parse()
}

var index_tmpl = `
<html lang="en">
<title>{{ .Title }}</title>
<body>
<form action="http://localhost:{{ .Port }}/receive" method="post" enctype="multipart/form-data">
<label for="file">Filename:</label>
<input type="file" name="file" id="file">
delete:<input type="checkbox" name="delete" value="remove">
<input type="submit" name="submit" value="Submit">
</form>
</body>
</html>
`
var static_tmpl = `
<html>
<title>Change to static</title>
<head>
<script>
window.location = "http://localhost:{{ .Port }}/{{ .Dir }}";
</script>
</head>
<body>
</body>
</html>
`

type Context struct {
	Title string
	Port  string
	Dir   string
}

func sendPage(w http.ResponseWriter, page string) {
	ctx := Context{
		Title: "Go file upload and delete",
		Port:  *port,
		Dir:   *dir,
	}

	t := template.New("upload")
	t, err := t.Parse(page)
	if err != nil {
		log.Print("template parsing error: ", err)
	}

	t.Execute(w, ctx)
	return
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v\n", r.URL)

	if strings.Contains(r.URL.Path, "favicon.ico") {
		w.Header().Add("Content-Type:", "image/icon")
		body, _ := ioutil.ReadFile("static/favicon.ico")
		fmt.Fprint(w, string(body))
		return
	}

	sendPage(w, index_tmpl)
	return
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

	/*
		//fmt.Printf("check at http://localhost:%s/static\n", *port)
		w.Header().Add("Content-Type:", "text/html")
		body, _ := ioutil.ReadFile("static/gostatic.html")
		fmt.Fprint(w, string(body))
	*/
	sendPage(w, static_tmpl)
}

func main() {
	// to serve goupload.html file in the browser
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/receive", uploadHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Printf("visit with http://localhost:%s\n", *port)
	http.ListenAndServe(":"+*port, nil)
}
