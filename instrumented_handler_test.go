package dropsonde_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

type FakeHandler struct{}

func (fh FakeHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("Hello World!"))
}

var _ = Describe("InstrumentedHandler", func() {

	var h http.Handler
	var req *http.Request

	BeforeEach(func() {
		var err error
		fh := FakeHandler{}
		h = dropsonde.InstrumentedHandler(fh)
		req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
		req.RemoteAddr = "127.0.0.1"
		req.Header.Set("User-Agent", "our-testing-client")
		Expect(err).To(BeNil())
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

		var fake *emitter.Fake
		var requestId *uuid.UUID

		BeforeEach(func() {
			fake = emitter.NewFake()
			emitter.DefaultEmitter = fake

			requestId, _ = uuid.NewV4()
			req.Header.Set("X-CF-RequestID", requestId.String())
			h.ServeHTTP(httptest.NewRecorder(), req)
		})

		It("should emit a start event", func() {
			expectedStartEvent := &events.HttpStart{
				RequestId:     events.NewUUID(requestId),
				PeerType:      events.PeerType_Server.Enum(),
				Method:        events.HttpStart_GET.Enum(),
				Uri:           proto.String("foo.example.com/"),
				RemoteAddress: proto.String("127.0.0.1"),
				UserAgent:     proto.String("our-testing-client"),
			}

			startEvent := fake.Messages[0].(*events.HttpStart)
			Expect(startEvent).ToNot(BeNil())
			Expect(startEvent.GetTimestamp()).ToNot(BeZero())
			startEvent.Timestamp = nil

			Expect(startEvent).To(Equal(expectedStartEvent))
		})

		It("should emit a stop event", func() {
			expectedStopEvent := &events.HttpStop{
				RequestId:     events.NewUUID(requestId),
				PeerType:      events.PeerType_Server.Enum(),
				StatusCode:    proto.Int32(200),
				ContentLength: proto.Int32(12),
			}

			stopEvent := fake.Messages[1].(*events.HttpStop)
			Expect(stopEvent).ToNot(BeNil())
			Expect(stopEvent.GetTimestamp()).ToNot(BeZero())
			stopEvent.Timestamp = nil
			Expect(stopEvent).To(Equal(expectedStopEvent))
		})
	})

})
