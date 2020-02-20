package verifyslack

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const MaxPermittedRequestAge time.Duration = 100 * time.Second

type timeGetter interface {
	Now() time.Time
}

func RequestHandler(handler http.HandlerFunc, timeGetter timeGetter, signingSecret string) http.HandlerFunc {
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

		timeNow := timeGetter.Now()
		if timeNow.After(time.Unix(intTimestamp, 0).Add(MaxPermittedRequestAge)) {
			http.Error(w, "request is too old to be handled", http.StatusBadRequest)
			return
		}

		var slackSignature string
		if slackSignature = r.Header.Get("X-Slack-Signature"); slackSignature == "" {
			http.Error(w, "request does not provide a Slack-signed signature", http.StatusUnauthorized)
			return
		}

		requestBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		expectedSignature := GenerateExpectedSignature(timestamp, requestBody, signingSecret)

		if !hmac.Equal([]byte(expectedSignature), []byte(slackSignature)) {
			http.Error(w, "request is not signed with a valid Slack signature", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}

func GenerateExpectedSignature(timestamp string, requestBody []byte, signingSecret string) string {
	baseSignature := append([]byte(fmt.Sprintf("v0:%s:", timestamp)), requestBody...)
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write(baseSignature)

	expectedSignature := fmt.Sprintf("v0=%s", hex.EncodeToString(mac.Sum(nil)))
	return expectedSignature
}
