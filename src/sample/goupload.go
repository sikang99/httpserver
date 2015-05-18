/*
	https://www.socketloop.com/tutorials/golang-upload-file
	http://stackoverflow.com/questions/28940005/golang-get-multipart-form-data
	http://blog.zmxv.com/2011/09/go-template-examples.html
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
	host    = flag.String("host", "localhost", "server address")
	port    = flag.String("port", "8080", "server port")
	dir     = flag.String("dir", "./static/", "directory to handle files")
	version = flag.String("version", "0.1.0", "program version")
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
<form action="http://{{ .Host }}:{{ .Port }}/receive" method="post" enctype="multipart/form-data">
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
window.location = "http://{{ .Host }}:{{ .Port }}/{{ .Dir }}";
</script>
</head>
<body>
</body>
</html>
`

type Context struct {
	Title string
	Host  string
	Port  string
	Dir   string
}

func sendPage(w http.ResponseWriter, page string) error {
	ctx := Context{
		Title: "Go file upload and delete",
		Host:  *host,
		Port:  *port,
		Dir:   *dir,
	}

	t, err := template.New("upload").Parse(page)
	if err != nil {
		log.Print("template parsing error: ", err)
		return err
	}

	return t.Execute(w, ctx)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v\n", r.URL)

	if strings.Contains(r.URL.Path, "favicon.ico") {
		w.Header().Set("Content-Type", "image/icon")
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

	// remove the file if checked
	if r.FormValue("delete") != "" {
		err = os.Remove(*dir + header.Filename)
		if err != nil {
			log.Printf("Unable to delete the file. Check your remove privilege")
			return
		}
		log.Printf("File '%s' deleted successfully.\n", header.Filename)
	}

	sendPage(w, static_tmpl)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/receive", uploadHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Printf("visit with http://%s:%s\n", *host, *port)
	http.ListenAndServe(":"+*port, nil)
}
