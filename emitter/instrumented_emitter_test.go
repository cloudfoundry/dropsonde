package emitter_test

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"github.com/cloudfoundry-incubator/dropsonde/heartbeat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InstrumentedEmitter", func() {
	It("implements HeartbeatDataSource", func() {
		// will fail during compile time
		var instrumentedEmitter heartbeat.HeartbeatDataSource = new(emitter.InstrumentedEmitter)
		Expect(instrumentedEmitter).ToNot(BeNil())
	})

	Describe("Delegators", func() {
		var fakeEmitter *emitter.FakeEmitter
		var instrumentedEmitter *emitter.InstrumentedEmitter

		BeforeEach(func() {
			fakeEmitter = emitter.NewFake()
			instrumentedEmitter, _ = emitter.NewInstrumentedEmitter(fakeEmitter)
		})

		It("delegates Close() to the concreteEmitter", func() {
			instrumentedEmitter.Close()
			Expect(fakeEmitter.IsClosed).To(BeTrue())
		})

		It("delegates SetOrigin() to the concreteEmitter", func() {
			origin := new(events.Origin)
			instrumentedEmitter.SetOrigin(origin)
			Expect(fakeEmitter.Origin).To(Equal(origin))
		})
	})

	Describe("Emit()", func() {
		var instrumentedEmitter *emitter.InstrumentedEmitter
		var testEvent *events.DropsondeStatus
		var fakeEmitter *emitter.FakeEmitter
		var origin events.Origin
		var jobIndex int32

		BeforeEach(func() {
			testEvent = &events.DropsondeStatus{SentCount: proto.Uint64(1), ErrorCount: proto.Uint64(0)}
			fakeEmitter = emitter.NewFake()
			instrumentedEmitter, _ = emitter.NewInstrumentedEmitter(fakeEmitter)
			jobName := "testInstrumentedEmitter"
			origin = events.Origin{JobName: &jobName, JobInstanceId: &jobIndex}
			instrumentedEmitter.SetOrigin(&origin)
		})
		It("calls the concrete emitter", func() {
			Expect(fakeEmitter.Messages).To(HaveLen(0))

			err := instrumentedEmitter.Emit(testEvent)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeEmitter.Messages).To(HaveLen(1))
			Expect(fakeEmitter.Messages[0].Event).To(Equal(testEvent))
			Expect(fakeEmitter.Messages[0].Origin).To(Equal(&origin))
		})
		It("increments the ReceivedMetricsCounter", func() {
			Expect(instrumentedEmitter.ReceivedMetricsCounter).To(BeNumerically("==", 0))

			err := instrumentedEmitter.Emit(testEvent)
			Expect(err).ToNot(HaveOccurred())

			Expect(instrumentedEmitter.ReceivedMetricsCounter).To(BeNumerically("==", 1))
		})
		Context("when the concrete Emitter returns no error on Emit()", func() {
			It("increments the SentMetricsCounter", func() {
				Expect(instrumentedEmitter.SentMetricsCounter).To(BeNumerically("==", 0))

				err := instrumentedEmitter.Emit(testEvent)
				Expect(err).ToNot(HaveOccurred())

				Expect(instrumentedEmitter.SentMetricsCounter).To(BeNumerically("==", 1))
			})
		})
		Context("when the concrete Emitter returns an error on Emit()", func() {
			BeforeEach(func() {
				fakeEmitter.ReturnError = true
			})
			It("increments the ErrorCounter", func() {
				Expect(instrumentedEmitter.ErrorCounter).To(BeNumerically("==", 0))
				Expect(instrumentedEmitter.ReceivedMetricsCounter).To(BeNumerically("==", 0))
				Expect(instrumentedEmitter.SentMetricsCounter).To(BeNumerically("==", 0))

				err := instrumentedEmitter.Emit(testEvent)
				Expect(err).To(HaveOccurred())

				Expect(instrumentedEmitter.ErrorCounter).To(BeNumerically("==", 1))
				Expect(instrumentedEmitter.ReceivedMetricsCounter).To(BeNumerically("==", 1))
				Expect(instrumentedEmitter.SentMetricsCounter).To(BeNumerically("==", 0))
			})
		})
	})

	Describe("NewInstrumentedEmitter", func() {
		Context("when the concrete Emitter is nil", func() {
			It("returns a nil instrumented emitter", func() {
				emitter, _ := emitter.NewInstrumentedEmitter(nil)
				Expect(emitter).To(BeNil())
			})
			It("returns a helpful error", func() {
				_, err := emitter.NewInstrumentedEmitter(nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
