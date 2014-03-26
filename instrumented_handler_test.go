package dropsonde_test

import (
	"github.com/cloudfoundry/dropsonde"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
}

var _ = Describe("InstrumentedHandler", func() {
	Describe("request ID", func() {
		var fh, h http.Handler
		var req *http.Request

		BeforeEach(func() {
			var err error
			fh = FakeHandler{}
			h = dropsonde.InstrumentedHandler(fh)
			req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
			Expect(err).To(BeNil())
		})

		It("should add it to the request", func() {
			h.ServeHTTP(httptest.NewRecorder(), req)
			Expect(req.Header.Get("X-CF-RequestID")).ToNot(BeEmpty())
		})

		It("should not add it to the request if it's already there", func() {
			req.Header.Set("X-CF-RequestID", "already-there")
			h.ServeHTTP(httptest.NewRecorder(), req)
			Expect(req.Header.Get("X-CF-RequestID")).To(Equal("already-there"))
		})

		It("should add it to the response", func() {
			req.Header.Set("X-CF-RequestID", "already-there")
			response := httptest.NewRecorder()
			h.ServeHTTP(response, req)
			Expect(response.Header().Get("X-CF-RequestID")).To(Equal("already-there"))
		})
	})

})
