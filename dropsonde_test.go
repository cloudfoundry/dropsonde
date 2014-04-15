package dropsonde_test

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"runtime"
	"time"
)

var _ = Describe("Dropsonde", func() {
	var origin = "awesome-job-name/42"

	Describe("Initialize", func() {
		It("errors if passed an origin with empty job name", func() {
			malformedOriginString := ""

			err := dropsonde.Initialize(malformedOriginString)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Cannot initialize dropsonde without an origin"))
		})

		PIt("creates a DefaultEmitter and starts generating heartbeats", func() {
			err := dropsonde.Initialize(origin)
			Expect(err).ToNot(HaveOccurred())
			emitter.DefaultEmitter = nil
			runtime.GC()

			//			Eventually(heartbeatEmitter.IsClosed).Should(BeTrue())
		})

		Context("when there is a preexisting DefaultEmitter", func() {
			// the existing emitter is overwritten
		})

		Context("something something", func() {
			var (
				heartbeatEmitter *emitter.FakeEmitter
			)

			BeforeEach(func() {
				heartbeatEmitter = emitter.NewFake(origin)
				emitter.HeartbeatInterval = 10 * time.Millisecond
			})

			It("Sets the origin information on emitter.DefaultEmitter", func() {
				fakeEmitter := emitter.NewFake(origin)
				emitter.DefaultEmitter = fakeEmitter

				dropsonde.Initialize(origin)
				Expect(fakeEmitter.Origin).To(Equal(origin))
			})

			Context("when the DefaultEmitter is a HeartbeatEventSource", func() {
				var fakeEmitter *emitter.FakeEmitter

				BeforeEach(func() {
					fakeEmitter = emitter.NewFake(origin)
					emitter.DefaultEmitter, _ = emitter.NewInstrumentedEmitter(fakeEmitter)
				})

				AfterEach(func() {
					Eventually(heartbeatEmitter.IsClosed).Should(BeTrue())
				})

				Context("when called for the first time", func() {
					// Figure out how this has changed
					PIt("starts the HeartbeatGenerator", func() {
						dropsonde.Initialize(origin)
						Expect(heartbeatEmitter.Origin).To(Equal(origin))

						Eventually(func() int { return len(heartbeatEmitter.GetMessages()) }).ShouldNot(BeZero())
					})
				})
			})
		})
	})
})
