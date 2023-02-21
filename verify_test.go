package verifyslack_test

import (
	"bytes"
	"crypto/hmac"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/coro/verifyslack"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GenerateExpectedSignature", func() {
	When("the function is passed metadata about the Slack request", func() {
		It("returns an expected value for the X-Slack-Signature header", func() {
			timestamp := "1582110731"
			requestBody := []byte("token=abbcbdbebdabddb&team_id=V1C2D3T4GH&team_domain=robbit&channel_id=RIF756S&channel_name=slack")
			signingSecret := "supersneakysecrets"

			// This was generated using:
			// https://github.com/slackapi/python-slack-events-api/blob/01a3d1b55ad3515c854b090599a5260ceb779344/slackeventsapi/server.py#L47
			expectedSignature := "v0=c98133278144dd816a12e9ae48fc056609fa6879eaf20a43ad3ea9f372aebf0d"

			Expect(hmac.Equal([]byte(expectedSignature), []byte(verifyslack.GenerateExpectedSignature(timestamp, requestBody, signingSecret)))).To(BeTrue())
		})
	})
})

var _ = Describe("RequestHandler", func() {
	When("the middleware handler receives a request", func() {
		var req *http.Request
		var requestBody []byte = []byte("token=abbcbdbebdabddb&team_id=V1C2D3T4GH&team_domain=robbit&channel_id=RIF756S&channel_name=slack")
		var err error
		var rr *httptest.ResponseRecorder
		var middlewareHandler http.HandlerFunc
		var return200OKHandler http.HandlerFunc
		var slackRequestTimestamp time.Time
		var validationTime time.Time
		var signingSecret string
		var requestSlackSignature string

		BeforeEach(func() {
			req, err = http.NewRequest("POST", "/", bytes.NewBuffer(requestBody))
			Expect(err).NotTo(HaveOccurred())
			rr = httptest.NewRecorder()
		})

		JustBeforeEach(func() {
			validationTimeGetter := validationTimeGetter{validationTime: validationTime}
			// We want to test the case where the handler is acting as middleware to another handler.
			// If the whole handler stack returns 200 OK with the same body as the request, we know the middleware
			// handler has verified the request and accepted the connection.
			return200OKHandler = http.HandlerFunc(func(rr http.ResponseWriter, req *http.Request) {
				reqBody, _ := ioutil.ReadAll(req.Body)
				fmt.Fprintf(rr, string(reqBody))
			})
			middlewareHandler = http.HandlerFunc(verifyslack.RequestHandler(return200OKHandler, validationTimeGetter, signingSecret))
		})

		When("the request has no timestamp header", func() {
			It("rejects the request with a 400", func() {
				middlewareHandler.ServeHTTP(rr, req)
				Expect(rr.Code).To(Equal(http.StatusBadRequest))
				Expect(rr.Body.String()).To(Equal("request did not contain a request timestamp\n"))
			})
		})

		When("the request has a timestamp header", func() {
			BeforeEach(func() {
				slackRequestTimestamp = time.Date(2010, time.January, 1, 2, 3, 0, 0, time.UTC)
				req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(slackRequestTimestamp.Unix(), 10))
			})

			When("the request has a timestamp header older than the max permitted request age", func() {
				BeforeEach(func() {
					validationTime = slackRequestTimestamp.Add(verifyslack.MaxPermittedRequestAge + time.Second)
				})
				It("rejects the request with a 400", func() {
					middlewareHandler.ServeHTTP(rr, req)
					Expect(rr.Code).To(Equal(http.StatusBadRequest))
					Expect(rr.Body.String()).To(Equal("request is too old to be handled\n"))
				})
			})

			When("the request has a valid timestamp but does not have a Slack signature header", func() {
				BeforeEach(func() {
					validationTime = slackRequestTimestamp.Add(verifyslack.MaxPermittedRequestAge)
				})
				It("rejects the request with a 401", func() {
					middlewareHandler.ServeHTTP(rr, req)
					Expect(rr.Code).To(Equal(http.StatusUnauthorized))
					Expect(rr.Body.String()).To(Equal("request does not provide a Slack-signed signature\n"))
				})
			})

			When("the request has a valid timestamp and has a Slack signature header", func() {
				JustBeforeEach(func() {
					req.Header.Set("X-Slack-Signature", requestSlackSignature)
				})

				When("the signature is any random string", func() {
					BeforeEach(func() {
						requestSlackSignature = "v0=aaabbbcccdddeee1233454567"
					})
					It("rejects the request", func() {
						middlewareHandler.ServeHTTP(rr, req)
						Expect(rr.Code).To(Equal(http.StatusUnauthorized))
						Expect(rr.Body.String()).To(Equal("request is not signed with a valid Slack signature\n"))
					})
				})

				When("the signature is a valid signature", func() {
					BeforeEach(func() {
						requestSlackSignature = verifyslack.GenerateExpectedSignature(strconv.FormatInt(slackRequestTimestamp.Unix(), 10), requestBody, signingSecret)
					})
					It("accepts the request", func() {
						middlewareHandler.ServeHTTP(rr, req)
						Expect(rr.Code).To(Equal(http.StatusOK))
						Expect(rr.Body.String()).To(Equal(string(requestBody)))
					})
				})
			})
		})
	})
})
