package emitter_test

import (
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/dropsonde/factories"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("EventEmitter", func() {
	var origin, deployment, job, index string

	BeforeEach(func() {
		origin = "fake-origin"
		deployment = "fake-deployment"
		job = "fake-job"
		index = "0"
	})

	Describe("Emit", func() {
		DescribeTable("returning errors for empty properties",
			func(origin, deployment, job, index string) {
				innerEmitter := fake.NewFakeByteEmitter()
				eventEmitter := emitter.NewEventEmitter(innerEmitter, origin, deployment, job, index)

				testEvent := factories.NewValueMetric("metric-name", 2.0, "metric-unit")
				err := eventEmitter.Emit(testEvent)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Wrap: "))
			},
			Entry("empty origin", "", "deployment", "job", "index"),
			Entry("empty deployment", "origin", "", "job", "index"),
			Entry("empty job", "origin", "deployment", "", "index"),
			Entry("empty index", "origin", "deployment", "job", ""),
		)

		It("marshals events and delegates to the inner emitter", func() {
			innerEmitter := fake.NewFakeByteEmitter()
			eventEmitter := emitter.NewEventEmitter(innerEmitter, origin, deployment, job, index)

			testEvent := factories.NewValueMetric("metric-name", 2.0, "metric-unit")
			err := eventEmitter.Emit(testEvent)
			Expect(err).ToNot(HaveOccurred())

			Expect(innerEmitter.GetMessages()).To(HaveLen(1))
			msg := innerEmitter.GetMessages()[0]

			var envelope events.Envelope
			err = proto.Unmarshal(msg, &envelope)
			Expect(err).ToNot(HaveOccurred())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_ValueMetric))
		})
	})

	Describe("EmitEnvelope", func() {
		It("marshals events and delegates to the inner emitter with same origin", func() {
			innerEmitter := fake.NewFakeByteEmitter()
			eventEmitter := emitter.NewEventEmitter(innerEmitter, origin, deployment, job, index)

			envOrigin := "original-origin"
			testEnvelope := events.Envelope{
				Origin:     proto.String(envOrigin),
				EventType:  events.Envelope_ValueMetric.Enum(),
				Timestamp:  proto.Int64(time.Now().Unix() * 1000),
				Deployment: proto.String("some-deployment"),
				Job:        proto.String("some-job"),
				Index:      proto.String("some-index"),
				ValueMetric: &events.ValueMetric{
					Name:  proto.String("event-name"),
					Value: proto.Float64(1.23),
					Unit:  proto.String("some-unit"),
				},
			}

			err := eventEmitter.EmitEnvelope(&testEnvelope)
			Expect(err).ToNot(HaveOccurred())

			Expect(innerEmitter.GetMessages()).To(HaveLen(1))
			msg := innerEmitter.GetMessages()[0]

			var envelope events.Envelope
			err = proto.Unmarshal(msg, &envelope)
			Expect(err).ToNot(HaveOccurred())
			Expect(envelope.GetEventType()).To(Equal(events.Envelope_ValueMetric))
			Expect(envelope.Origin).To(Equal(proto.String(envOrigin)))
		})
	})

	Describe("Close", func() {
		It("closes the inner emitter", func() {
			innerEmitter := fake.NewFakeByteEmitter()
			eventEmitter := emitter.NewEventEmitter(innerEmitter, "", deployment, job, index)

			eventEmitter.Close()
			Expect(innerEmitter.IsClosed()).To(BeTrue())
		})
	})
})
