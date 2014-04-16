package emitter_test

import (
	"bytes"
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"github.com/cloudfoundry-incubator/dropsonde/factories"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"log"
	"time"
)

var _ = Describe("HeartbeatEmitter", func() {
	var (
		wrappedEmitter *emitter.FakeEmitter
	)

	BeforeEach(func() {
		emitter.HeartbeatInterval = 10 * time.Millisecond
		origin := "testHeartbeatEmitter/0"
		wrappedEmitter = emitter.NewFake(origin)
	})

	Describe("NewHeartbeatEmitter", func() {
		It("requires non-nil args", func() {
			hbEmitter, err := emitter.NewHeartbeatEmitter(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("wrappedEmitter is nil"))
			Expect(hbEmitter).To(BeNil())
		})

		It("starts periodic heartbeat emission", func() {
			hbEmitter, err := emitter.NewHeartbeatEmitter(wrappedEmitter)
			Expect(err).NotTo(HaveOccurred())
			Expect(hbEmitter).NotTo(BeNil())

			Eventually(func() int { return len(wrappedEmitter.GetMessages()) }).Should(BeNumerically(">=", 2))
		})

		It("logs an error when heartbeat emission fails", func() {
			wrappedEmitter.ReturnError = errors.New("fake error")

			logWriter := new(bytes.Buffer)
			log.SetOutput(logWriter)

			hbEmitter, _ := emitter.NewHeartbeatEmitter(wrappedEmitter)

			Eventually(func() int { return len(wrappedEmitter.GetMessages()) }).Should(BeNumerically(">=", 2))

			loggedText := string(logWriter.Bytes())
			expectedText := "fake error"
			Expect(loggedText).To(ContainSubstring(expectedText))
			hbEmitter.Close()
		})
	})

	Describe("Emit", func() {
		var (
			hbEmitter emitter.Emitter
			testEvent events.Event
		)
		BeforeEach(func() {
			hbEmitter, _ = emitter.NewHeartbeatEmitter(wrappedEmitter)
			testEvent = factories.NewHeartbeat(42, 0, 0)
		})

		It("delegates to the wrapped emitter", func() {
			hbEmitter.Emit(testEvent)

			messages := wrappedEmitter.GetMessages()
			Expect(messages).To(HaveLen(1))
			Expect(messages[0].Event).To(Equal(testEvent))
		})

		It("increments the heartbeat counter", func() {
			hbEmitter.Emit(testEvent)

			Eventually(func() bool {
				messages := wrappedEmitter.GetMessages()

				for _, message := range messages {
					hbEvent, ok := message.Event.(*events.Heartbeat)
					if ok && hbEvent.GetReceivedCount() == 1 {
						return true
					}
				}

				return false
			}).Should(BeTrue())
		})
	})

	Describe("Close", func() {
		var hbEmitter emitter.Emitter

		BeforeEach(func() {
			hbEmitter, _ = emitter.NewHeartbeatEmitter(wrappedEmitter)
		})

		It("eventually delegates to the inner heartbeat emitter", func() {
			hbEmitter.Close()
			Eventually(wrappedEmitter.IsClosed).Should(BeTrue())
		})

		It("can be called more than once", func() {
			hbEmitter.Close()
			Expect(hbEmitter.Close).ToNot(Panic())
		})
	})
})
