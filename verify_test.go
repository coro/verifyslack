package verifyslack_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/coro/verifyslack"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VerifySlackRequests", func() {
	When("the middleware handler receives a request", func() {
		var req *http.Request
		var err error
		var rr *httptest.ResponseRecorder
		var middlewareHandler http.HandlerFunc
		var return200OKHandler http.HandlerFunc

		BeforeEach(func() {
			req, err = http.NewRequest("POST", "/", nil)
			Expect(err).NotTo(HaveOccurred())
			rr = httptest.NewRecorder()
		})

		JustBeforeEach(func() {
			// We want to test the case where the handler is acting as middleware to another handler.
			// If the whole handler stack returns 200 OK with a body of 'OK', we know the middleware
			// handler has verified the request and accepted the connection.
			return200OKHandler = http.HandlerFunc(func(rr http.ResponseWriter, req *http.Request) { fmt.Fprintf(rr, "OK") })
			middlewareHandler = http.HandlerFunc(verifyslack.RequestHandler(return200OKHandler, time.Now()))
		})

		It("accepts the request", func() {
			middlewareHandler.ServeHTTP(rr, req)
			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal("OK"))
		})
	})
})
