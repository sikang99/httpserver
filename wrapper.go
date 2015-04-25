package main

import (
	"log"
	"net/http"
)

// wrapper function for logging of http.HandlerFunc
func logging(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Before")
		fn(w, r)
		log.Println("After")
	}
}

// check API required parameters
func checkAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Query().Get("key")) == 0 {
			http.Error(w, "missing key", http.StatusUnauthorized)
			return // don't call original handler
		}
		fn(w, r)
	}
}

// Passing arguments to the wrappers
func MustParams(fn http.HandlerFunc, params ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for _, param := range params {
			if len(r.URL.Query().Get(param)) == 0 {
				http.Error(w, "missing "+param, http.StatusBadRequest)
				return
			}
		}
		fn(w, r) // success - call handler
	}
}
