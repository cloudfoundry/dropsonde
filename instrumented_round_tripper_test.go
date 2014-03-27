package dropsonde_test

import (
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

type FakeRoundTripper struct{}

func (frt FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 123, ContentLength: 1234}, nil
}

var _ = Describe("InstrumentedRoundTripper", func() {
	var rt http.RoundTripper
	var req *http.Request

	BeforeEach(func() {
		var err error
		rt = dropsonde.InstrumentedRoundTripper(FakeRoundTripper{})

		req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
		Expect(err).To(BeNil())
		req.RemoteAddr = "127.0.0.1"
		req.Header.Set("User-Agent", "our-testing-client")

	})

	Describe("request ID", func() {
		It("should generate a new request ID", func() {
			rt.RoundTrip(req)
			Expect(req.Header.Get("X-CF-RequestID")).ToNot(BeEmpty())
		})

	})

	Context("event emission", func() {
		var fake *emitter.Fake

		BeforeEach(func() {
			fake = emitter.NewFake()
			emitter.DefaultEmitter = fake
		})

		It("should emit a start event", func() {
			rt.RoundTrip(req)
			Expect(fake.Messages[0]).To(BeAssignableToTypeOf(new(events.HttpStart)))
		})

		Context("if request ID already exists", func() {
			var existingRequestId *uuid.UUID

			BeforeEach(func() {
				existingRequestId, _ = uuid.NewV4()
				req.Header.Set("X-CF-RequestID", existingRequestId.String())
			})

			It("should emit the existing request ID as the parent request ID", func() {
				rt.RoundTrip(req)
				startEvent := fake.Messages[0].(*events.HttpStart)
				Expect(startEvent.GetParentRequestId()).To(Equal(events.NewUUID(existingRequestId)))
			})
		})

		It("should emit a stop event", func() {
			rt.RoundTrip(req)

			Expect(fake.Messages[1]).To(BeAssignableToTypeOf(new(events.HttpStop)))

			stopEvent := fake.Messages[1].(*events.HttpStop)
			Expect(stopEvent.GetStatusCode()).To(BeNumerically("==", 123))
			Expect(stopEvent.GetContentLength()).To(BeNumerically("==", 1234))
		})
	})
})
