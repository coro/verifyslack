package verifyslack

import (
	"net/http"
	"time"
)

func RequestHandler(handler http.HandlerFunc, timeNow time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}
}
