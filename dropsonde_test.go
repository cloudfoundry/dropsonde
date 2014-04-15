package dropsonde_test

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Dropsonde", func() {
	var origin = "awesome-job-name/42"
	var heartbeatEmitter *emitter.FakeEmitter

	Describe("Initialize", func() {
		BeforeEach(func() {
			heartbeatEmitter = emitter.NewFake(origin)
			emitter.HeartbeatInterval = 10 * time.Millisecond
		})

		It("errors if passed an origin with empty job name", func() {
			malformedOriginString := ""

			err := dropsonde.Initialize(malformedOriginString)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Cannot initialize dropsonde without an origin"))
		})

		It("Sets the origin information on emitter.DefaultEmitter", func() {
			fakeEmitter := emitter.NewFake(origin)
			emitter.DefaultEmitter = fakeEmitter

			dropsonde.Initialize(origin)
			Expect(fakeEmitter.Origin).To(Equal(origin))
		})
	})
})
