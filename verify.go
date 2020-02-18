package verifyslack

import (
	"net/http"
	"strconv"
	"time"
)

const MaxPermittedRequestAge time.Duration = 100 * time.Second

func RequestHandler(handler http.HandlerFunc, timeNow time.Time) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var timestamp string
		if timestamp = r.Header.Get("X-Slack-Request-Timestamp"); timestamp == "" {
			http.Error(w, "request did not contain a request timestamp", http.StatusBadRequest)
			return
		}

		intTimestamp, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			http.Error(w, "failed to parse request timestamp", http.StatusInternalServerError)
			return
		}

		if timeNow.After(time.Unix(intTimestamp, 0).Add(MaxPermittedRequestAge)) {
			http.Error(w, "request did not contain a request timestamp", http.StatusBadRequest)
			return
		}

		handler(w, r)
	}
}
