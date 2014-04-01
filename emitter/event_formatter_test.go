package emitter_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	uuid "github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type unknownEvent struct{}

func (*unknownEvent) ProtoMessage() {}

var _ = Describe("EventFormatter", func() {
	Describe("wrap", func() {
		var origin events.Origin
		var jobIndex int32 = 42

		BeforeEach(func() {
			jobName := "testEventFormatter"
			origin = events.Origin{JobName: &jobName, JobInstanceId: &jobIndex}
		})

		It("should work with HttpStart events", func() {
			id, _ := uuid.NewV4()
			testEvent := &events.HttpStart{RequestId: events.NewUUID(id)}

			envelope, err := emitter.Wrap(testEvent, &origin)
			Expect(err).To(BeNil())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_HttpStart))
			Expect(envelope.GetHttpStart()).To(Equal(testEvent))
		})

		It("should work with HttpStop events", func() {
			id, _ := uuid.NewV4()
			testEvent := &events.HttpStop{RequestId: events.NewUUID(id)}

			envelope, err := emitter.Wrap(testEvent, &origin)
			Expect(err).To(BeNil())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_HttpStop))
			Expect(envelope.GetHttpStop()).To(Equal(testEvent))
		})

		It("should error with unknown events", func() {
			envelope, err := emitter.Wrap(new(unknownEvent), &origin)
			Expect(envelope).To(BeNil())
			Expect(err).ToNot(BeNil())
		})

		It("should work with dropsonde status events", func() {
			statusEvent := &events.DropsondeStatus{SentCount: proto.Uint64(1), ErrorCount: proto.Uint64(0)}
			envelope, err := emitter.Wrap(statusEvent, &origin)
			Expect(err).To(BeNil())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_DropsondeStatus))
			Expect(envelope.GetDropsondeStatus()).To(Equal(statusEvent))
		})

		Context("with a known event type", func() {

			var testEvent emitter.Event

			BeforeEach(func() {
				id, _ := uuid.NewV4()
				testEvent = &events.HttpStop{RequestId: events.NewUUID(id)}
			})

			It("should contain the jobName in the origin", func() {
				envelope, _ := emitter.Wrap(testEvent, &origin)
				Expect(envelope.GetOrigin().GetJobName()).To(Equal("testEventFormatter"))
			})

			It("should contain the jobIndex in the origin", func() {
				envelope, _ := emitter.Wrap(testEvent, &origin)
				Expect(envelope.GetOrigin().GetJobInstanceId()).To(BeNumerically("==", 42))
			})

			Context("when the origin is nil", func() {
				It("should error with a helpful message", func() {
					envelope, err := emitter.Wrap(testEvent, nil)
					Expect(envelope).To(BeNil())
					Expect(err.Error()).To(Equal("Event not emitted due to missing origin information"))
				})
			})
		})
	})
})
