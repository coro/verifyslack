# verifyslack
A simple Golang middleware HTTP handler for verifying inbound Slack messages .

# Overview

`verifyslack` provides a simple `RequestHandler` wrapper for `http.HandlerFunc`s. Any inbound HTTP requests which cannot be verified as a legitimate Slack request, signed with a specified Slack App [Signing Secret](https://api.slack.com/authentication/verifying-requests-from-slack#signing_secrets_admin_page), are rejected.

Any requests which are successfully verified are forwarded to the HTTP handler contained within the wrapper.

# Example Usage

```go
package main

import (
  "net/http"
  "os"
  "time"

  "github.com/coro/verifyslack"
)

func wrappedHandler(w http.ResponseWriter, req *http.Request) {
  // ... handle the validated Slack request
}

func main() {
  http.HandleFunc("/slack", verifyslack.RequestHandler(wrappedHandler, time.Now(), os.Getenv("SLACK_SIGNING_SECRET")))
  http.ListenAndServe(":8090", nil)
}

```

Any requests with invalid signatures (or expired timestamps) are then rejected:

```bash
$ curl -i -H "X-Slack-Request-Timestamp: `date +%s`" -H "X-Slack-Signature: v0=abcabcabcabcabc" localhost:8090/slack
HTTP/1.1 401 Unauthorized
Content-Type: text/plain; charset=utf-8
X-Content-Type-Options: nosniff
Date: Wed, 19 Feb 2020 17:27:36 GMT
Content-Length: 51

request is not signed with a valid Slack signature
```

# More info

This repo is based on the [instructions published by Slack](https://api.slack.com/authentication/verifying-requests-from-slack) for verifying requests.
