package dropsonde_test

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Hello World!"))
	rw.WriteHeader(123)
}

var _ = Describe("InstrumentedHandler", func() {
	var fake *emitter.FakeEmitter
	var h http.Handler
	var req *http.Request

	var origin = events.NewOrigin("testHandler", 41)

	Context("when dropsonde.Initialize fails", func() {
		var originalAddr string
		BeforeEach(func() {
			emitter.DefaultEmitter = nil
			originalAddr = dropsonde.DefaultEmitterRemoteAddr
			dropsonde.DefaultEmitterRemoteAddr = "invalid-address:"
		})

		AfterEach(func() {
			dropsonde.DefaultEmitterRemoteAddr = originalAddr
		})

		It("returns an error", func() {
			fh := FakeHandler{}
			h, err := dropsonde.InstrumentedHandler(fh, "job-name", 41)
			Expect(err).To(HaveOccurred())
			Expect(h).To(BeNil())
		})
	})

	Context("when dropsonde.Initialize doesn't fail", func() {
		BeforeEach(func() {
			fake = emitter.NewFake(origin)
			emitter.DefaultEmitter = fake

			var err error
			fh := FakeHandler{}
			h, err = dropsonde.InstrumentedHandler(fh, "testHandler", 41)
			Expect(err).ToNot(HaveOccurred())
			req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
			Expect(err).ToNot(HaveOccurred())
			req.RemoteAddr = "127.0.0.1"
			req.Header.Set("User-Agent", "our-testing-client")
		})

		Describe("request ID", func() {
			It("should add it to the request", func() {
				h.ServeHTTP(httptest.NewRecorder(), req)
				Expect(req.Header.Get("X-CF-RequestID")).ToNot(BeEmpty())
			})

			It("should not add it to the request if it's already there", func() {
				id, _ := uuid.NewV4()
				req.Header.Set("X-CF-RequestID", id.String())
				h.ServeHTTP(httptest.NewRecorder(), req)
				Expect(req.Header.Get("X-CF-RequestID")).To(Equal(id.String()))
			})

			It("should create a valid one if it's given an invalid one", func() {
				req.Header.Set("X-CF-RequestID", "invalid")
				h.ServeHTTP(httptest.NewRecorder(), req)
				Expect(req.Header.Get("X-CF-RequestID")).ToNot(Equal("invalid"))
				Expect(req.Header.Get("X-CF-RequestID")).ToNot(BeEmpty())
			})

			It("should add it to the response", func() {
				id, _ := uuid.NewV4()
				req.Header.Set("X-CF-RequestID", id.String())
				response := httptest.NewRecorder()
				h.ServeHTTP(response, req)
				Expect(response.Header().Get("X-CF-RequestID")).To(Equal(id.String()))
			})
		})

		Describe("event emission", func() {
			var requestId *uuid.UUID

			BeforeEach(func() {
				requestId, _ = uuid.NewV4()
				req.Header.Set("X-CF-RequestID", requestId.String())
			})

			Context("without an application ID or instanceIndex", func() {
				BeforeEach(func() {
					h.ServeHTTP(httptest.NewRecorder(), req)
				})

				It("should emit a start event with the right origin", func() {
					Expect(fake.Messages[0].Event).To(BeAssignableToTypeOf(new(events.HttpStart)))
					Expect(fake.Messages[0].Origin.GetJobName()).To(Equal("testHandler"))
					Expect(fake.Messages[0].Origin.GetJobInstanceId()).To(BeNumerically("==", 41))
				})

				It("should emit a stop event", func() {
					Expect(fake.Messages[1].Event).To(BeAssignableToTypeOf(new(events.HttpStop)))
					stopEvent := fake.Messages[1].Event.(*events.HttpStop)
					Expect(stopEvent.GetStatusCode()).To(BeNumerically("==", 123))
					Expect(stopEvent.GetContentLength()).To(BeNumerically("==", 12))
				})
			})
		})
	})

})
