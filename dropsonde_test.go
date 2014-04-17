package dropsonde_test

import (
	"github.com/cloudfoundry-incubator/dropsonde"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
	"time"
)

var _ = Describe("Dropsonde", func() {
	var origin = "awesome-job-name/42"
	var heartbeatEmitter *emitter.FakeEmitter
	var existingDefaultEmitterRemoteAddr string

	Describe("Initialize", func() {
		BeforeEach(func() {
			heartbeatEmitter = emitter.NewFake(origin)
			emitter.HeartbeatInterval = 10 * time.Millisecond
			existingDefaultEmitterRemoteAddr = dropsonde.DefaultEmitterRemoteAddr
		})

		AfterEach(func() {
			dropsonde.DefaultEmitterRemoteAddr = existingDefaultEmitterRemoteAddr
		})

		It("errors if passed an origin with empty job name", func() {
			malformedOriginString := ""

			err := dropsonde.Initialize(malformedOriginString)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Cannot initialize dropsonde without an origin"))
		})

		It("errors if DefaultEmitterRemoteAddr is invalid", func() {
			dropsonde.DefaultEmitterRemoteAddr = "localhost"
			err := dropsonde.Initialize(origin)
			Expect(err).To(HaveOccurred())
		})

		Context("succesfully initialized", func() {

			BeforeEach(func() {
				err := dropsonde.Initialize(origin)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Sets the emitter.DefaultEmitter to be a HearbeatEmitter", func() {
				Expect(reflect.TypeOf(emitter.DefaultEmitter).Elem().Name()).To(Equal("heartbeatEmitter"))
			})
		})
	})
})
