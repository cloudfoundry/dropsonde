package dropsonde_test

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

type FakeRoundTripper struct {
	FakeError error
}

func (frt *FakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 123, ContentLength: 1234}, frt.FakeError
}

var _ = Describe("InstrumentedRoundTripper", func() {
	var fakeRoundTripper *FakeRoundTripper
	var rt http.RoundTripper
	var req *http.Request
	var fake *emitter.FakeEmitter

	var origin = "testRoundtripper/42"

	Context("when dropsonde.Initialize succeeds", func() {
		BeforeEach(func() {
			var err error
			fake = emitter.NewFake(origin)
			emitter.DefaultEmitter = fake

			fakeRoundTripper = new(FakeRoundTripper)
			rt, err = dropsonde.InstrumentedRoundTripper(fakeRoundTripper)
			Expect(err).ToNot(HaveOccurred())

			req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
			Expect(err).ToNot(HaveOccurred())
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

			It("should emit a start event", func() {
				rt.RoundTrip(req)
				Expect(fake.Messages[0].Event).To(BeAssignableToTypeOf(new(events.HttpStart)))
				Expect(fake.Messages[0].Origin).To(Equal("testRoundtripper/42"))
			})

			Context("if request ID already exists", func() {
				var existingRequestId *uuid.UUID

				BeforeEach(func() {
					existingRequestId, _ = uuid.NewV4()
					req.Header.Set("X-CF-RequestID", existingRequestId.String())
				})

				It("should emit the existing request ID as the parent request ID", func() {
					rt.RoundTrip(req)
					startEvent := fake.Messages[0].Event.(*events.HttpStart)
					Expect(startEvent.GetParentRequestId()).To(Equal(events.NewUUID(existingRequestId)))
				})
			})

			Context("if round tripper returns an error", func() {
				It("should emit a stop event with blank response fields", func() {
					fakeRoundTripper.FakeError = errors.New("fake error")
					rt.RoundTrip(req)

					Expect(fake.Messages[1].Event).To(BeAssignableToTypeOf(new(events.HttpStop)))

					stopEvent := fake.Messages[1].Event.(*events.HttpStop)
					Expect(stopEvent.GetStatusCode()).To(BeNumerically("==", 0))
					Expect(stopEvent.GetContentLength()).To(BeNumerically("==", 0))
				})
			})

			Context("if round tripper does not return an error", func() {
				It("should emit a stop event with the round tripper's response", func() {
					rt.RoundTrip(req)

					Expect(fake.Messages[1].Event).To(BeAssignableToTypeOf(new(events.HttpStop)))

					stopEvent := fake.Messages[1].Event.(*events.HttpStop)
					Expect(stopEvent.GetStatusCode()).To(BeNumerically("==", 123))
					Expect(stopEvent.GetContentLength()).To(BeNumerically("==", 1234))
				})
			})
		})
	})

})
