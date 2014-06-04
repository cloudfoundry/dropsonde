package factories_test

import (
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"github.com/cloudfoundry-incubator/dropsonde/factories"
	"net/http"
)

var _ = Describe("HTTP event creation", func() {
	var requestId *uuid.UUID
	BeforeEach(func() {
		requestId, _ = uuid.NewV4()
	})

	Describe("NewHttpStart", func() {
		var req *http.Request

		BeforeEach(func() {
			var err error

			req, err = http.NewRequest("GET", "http://foo.example.com/", nil)
			Expect(err).To(BeNil())

			req.RemoteAddr = "127.0.0.1"
			req.Header.Set("User-Agent", "our-testing-client")
		})

		Context("without an application ID or instanceIndex", func() {

			It("should set appropriate fields", func() {
				expectedStartEvent := &events.HttpStart{
					RequestId:     factories.NewUUID(requestId),
					PeerType:      events.PeerType_Server.Enum(),
					Method:        events.HttpStart_GET.Enum(),
					Uri:           proto.String("foo.example.com/"),
					RemoteAddress: proto.String("127.0.0.1"),
					UserAgent:     proto.String("our-testing-client"),
				}

				startEvent := factories.NewHttpStart(req, events.PeerType_Server, requestId)

				Expect(startEvent.GetTimestamp()).ToNot(BeZero())
				startEvent.Timestamp = nil

				Expect(startEvent).To(Equal(expectedStartEvent))
			})
		})

		Context("with an application ID", func() {
			It("should include it in the start event", func() {
				applicationId, _ := uuid.NewV4()
				req.Header.Set("X-CF-ApplicationID", applicationId.String())

				startEvent := factories.NewHttpStart(req, events.PeerType_Server, requestId)

				Expect(startEvent.GetApplicationId()).To(Equal(factories.NewUUID(applicationId)))
			})
		})

		Context("with an application instance index", func() {
			It("should include it in the start event", func() {
				req.Header.Set("X-CF-InstanceIndex", "1")

				startEvent := factories.NewHttpStart(req, events.PeerType_Server, requestId)

				Expect(startEvent.GetInstanceIndex()).To(BeNumerically("==", 1))
			})
		})
	})

	Describe("NewHttpStop", func() {
		It("should set appropriate fields", func() {
			expectedStopEvent := &events.HttpStop{
				RequestId:     factories.NewUUID(requestId),
				PeerType:      events.PeerType_Server.Enum(),
				StatusCode:    proto.Int32(200),
				ContentLength: proto.Int64(12),
			}

			stopEvent := factories.NewHttpStop(200, 12, events.PeerType_Server, requestId)

			Expect(stopEvent.GetTimestamp()).ToNot(BeZero())
			stopEvent.Timestamp = nil

			Expect(stopEvent).To(Equal(expectedStopEvent))
		})
	})

	Describe("StringFromUUID", func() {
		It("returns a string for a UUID", func() {
			id := factories.NewUUID(requestId)
			Expect(factories.StringFromUUID(id)).To(Equal(requestId.String()))
		})
	})
})
