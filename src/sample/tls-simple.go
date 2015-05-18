/*
	http://stackoverflow.com/questions/12122159/golang-how-to-do-a-https-request-with-bad-certificate
	http://golang.org/pkg/crypto/tls/
*/
package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

func main() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	_, err := client.Get("https://golang.org/")
	if err != nil {
		fmt.Println(err)
	}
}
