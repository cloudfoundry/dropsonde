package emitter_test

import (
	"github.com/cloudfoundry/dropsonde/emitter"

	"time"

	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/nu7hatch/gouuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/localip"
)

type unknownEvent struct{}

func (*unknownEvent) ProtoMessage() {}

var _ = Describe("EventFormatter", func() {
	Describe("wrap", func() {
		var (
			origin, deployment, job, index string
		)

		BeforeEach(func() {
			origin = "testEventFormatter/42"
			deployment = "some-deployment"
			job = "some-job"
			index = "0"
		})

		It("works with HttpStart events", func() {
			id, _ := uuid.NewV4()
			testEvent := &events.HttpStart{RequestId: factories.NewUUID(id)}

			envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_HttpStart))
			Expect(envelope.GetHttpStart()).To(Equal(testEvent))
		})

		It("works with HttpStop events", func() {
			id, _ := uuid.NewV4()
			testEvent := &events.HttpStop{RequestId: factories.NewUUID(id)}

			envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_HttpStop))
			Expect(envelope.GetHttpStop()).To(Equal(testEvent))
		})

		It("works with ValueMetric events", func() {
			testEvent := &events.ValueMetric{Name: proto.String("test-name")}

			envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_ValueMetric))
			Expect(envelope.GetValueMetric()).To(Equal(testEvent))
		})

		It("works with CounterEvent events", func() {
			testEvent := &events.CounterEvent{Name: proto.String("test-counter")}

			envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_CounterEvent))
			Expect(envelope.GetCounterEvent()).To(Equal(testEvent))
		})

		It("works with HttpStartStop events", func() {
			testEvent := &events.HttpStartStop{
				StartTimestamp: proto.Int64(200),
				StopTimestamp:  proto.Int64(500),
				RequestId: &events.UUID{
					Low:  proto.Uint64(200),
					High: proto.Uint64(300),
				},
				PeerType:      events.PeerType_Client.Enum(),
				Method:        events.Method_GET.Enum(),
				Uri:           proto.String("http://some.example.com"),
				RemoteAddress: proto.String("http://remote.address"),
				UserAgent:     proto.String("some user agent"),
				ContentLength: proto.Int64(200),
				StatusCode:    proto.Int32(200),
			}

			envelope, err := emitter.Wrap(testEvent, origin, deployment, job, index)
			Expect(err).ToNot(HaveOccurred())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_HttpStartStop))
			Expect(envelope.GetHttpStartStop()).To(Equal(testEvent))
		})

		It("errors with unknown events", func() {
			envelope, err := emitter.Wrap(new(unknownEvent), origin, deployment, job, index)
			Expect(envelope).To(BeNil())
			Expect(err).To(HaveOccurred())
		})

		It("checks that origin is non-empty", func() {
			id, _ := uuid.NewV4()
			malformedOrigin := ""
			testEvent := &events.HttpStart{RequestId: factories.NewUUID(id)}
			envelope, err := emitter.Wrap(testEvent, malformedOrigin, deployment, job, index)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Event not emitted due to missing origin information"))
			Expect(envelope).To(BeNil())
		})

		Context("with a known event type", func() {
			var testEvent events.Event

			BeforeEach(func() {
				id, _ := uuid.NewV4()
				testEvent = &events.HttpStop{RequestId: factories.NewUUID(id)}
			})

			It("contains the origin", func() {
				envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
				Expect(envelope.GetOrigin()).To(Equal("testEventFormatter/42"))
			})

			It("contains the deployment", func() {
				envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
				Expect(envelope.GetDeployment()).To(Equal(deployment))
			})

			It("contains the job", func() {
				envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
				Expect(envelope.GetJob()).To(Equal(job))
			})

			It("contains the index", func() {
				envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
				Expect(envelope.GetIndex()).To(Equal(index))
			})

			It("contains the IP", func() {
				envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
				ip, _ := localip.LocalIP()
				Expect(envelope.GetIp()).To(Equal(ip))
			})

			Context("when the origin is empty", func() {
				It("errors with a helpful message", func() {
					envelope, err := emitter.Wrap(testEvent, "", deployment, job, index)
					Expect(envelope).To(BeNil())
					Expect(err.Error()).To(Equal("Event not emitted due to missing origin information"))
				})
			})

			It("sets the timestamp to now", func() {
				envelope, _ := emitter.Wrap(testEvent, origin, deployment, job, index)
				Expect(time.Unix(0, envelope.GetTimestamp())).To(BeTemporally("~", time.Now(), 100*time.Millisecond))
			})
		})
	})
})
