package main

import (
	"net/http"
	"time"
)

func VerifySlackRequests(handler http.HandlerFunc, timeNow time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}
}
