package dropsonde_test

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"github.com/cloudfoundry-incubator/dropsonde/heartbeat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Dropsonde", func() {
	var (
		jobName           = "awesome-job-name"
		jobInstance int32 = 42
		origin            = &events.Origin{JobName: &jobName, JobInstanceId: &jobInstance}
	)

	Describe("Initialize", func() {
		Context("when there is no DefaultEmitter", func() {
			It("does not panic", func() {
				emitter.DefaultEmitter = nil
				Expect(func() { dropsonde.Initialize(nil) }).ToNot(Panic())
				dropsonde.Cleanup()
			})
		})

		Context("when there is a DefaultEmitter", func() {
			var (
				heartbeatEmitter *emitter.FakeEmitter
			)

			BeforeEach(func() {
				heartbeatEmitter = emitter.NewFake(origin)
				heartbeat.HeartbeatEmitter = heartbeatEmitter
				heartbeat.HeartbeatInterval = 10 * time.Millisecond
			})

			It("Sets the origin information on emitter.DefaultEmitter", func() {
				fakeEmitter := emitter.NewFake(origin)
				emitter.DefaultEmitter = fakeEmitter

				dropsonde.Initialize(origin)
				Expect(fakeEmitter.Origin).To(Equal(origin))
			})

			Context("when the DefaultEmitter is not a HeartbeatEventSource", func() {
				var fakeEmitter = emitter.NewFake(origin)

				BeforeEach(func() {
					emitter.DefaultEmitter = fakeEmitter

				})

				It("does not start the HeartbeatGenerator", func() {
					dropsonde.Initialize(origin)
					Expect(heartbeat.HeartbeatEmitter).To(BeNil())
				})
			})

			Context("when the DefaultEmitter is a HeartbeatEventSource", func() {
				var fakeEmitter *emitter.FakeEmitter

				BeforeEach(func() {
					fakeEmitter = emitter.NewFake(origin)
					emitter.DefaultEmitter, _ = emitter.NewInstrumentedEmitter(fakeEmitter)
				})

				AfterEach(func() {
					dropsonde.Cleanup()
				})

				Context("when called for the first time", func() {
					It("starts the HeartbeatGenerator", func() {
						dropsonde.Initialize(origin)
						Expect(heartbeatEmitter.Origin).To(Equal(origin))

						Eventually(func() int { return len(heartbeatEmitter.GetMessages()) }).ShouldNot(BeZero())
					})
				})

				Context("when subsequently called", func() {
					It("does not create a new HeartbeatGenerator", func() {
						var oldHeartbeatEmitter = heartbeat.HeartbeatEmitter
						dropsonde.Initialize(origin)
						dropsonde.Initialize(origin)
						Expect(heartbeat.HeartbeatEmitter).To(Equal(oldHeartbeatEmitter))
					})
				})
			})
		})
	})

	Describe("Cleanup", func() {
		Context("when no HeartbeatGenerator is running", func() {
			It("does not panic", func() {
				Expect(dropsonde.Cleanup).ToNot(Panic())
			})
		})

		Context("when the HeartbeatGenerator is running", func() {
			It("stops the HeartbeatGenerator", func() {
				fakeEmitter := emitter.NewFake(origin)
				emitter.DefaultEmitter, _ = emitter.NewInstrumentedEmitter(fakeEmitter)
				heartbeatEmitter := emitter.NewFake(origin)
				heartbeat.HeartbeatEmitter = heartbeatEmitter
				dropsonde.Initialize(nil)

				dropsonde.Cleanup()

				Eventually(func() bool { return heartbeatEmitter.IsClosed }).Should(BeTrue())
			})
		})
	})
})
