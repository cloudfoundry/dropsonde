package emitter_test

import (
	"bytes"
	"errors"
	"log"

	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/emitter/fake"
	"github.com/cloudfoundry/dropsonde/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PingResponder", func() {
	var (
		wrappedEmitter *fake.FakeByteEmitter
		origin         = "testPingResponder/0"
	)

	BeforeEach(func() {
		wrappedEmitter = fake.NewFakeByteEmitter()
	})

	Describe("NewPingResponder", func() {
		It("requires non-nil args", func() {
			pingResponder, err := emitter.NewPingResponder(nil, origin)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("wrappedEmitter is nil"))
			Expect(pingResponder).To(BeNil())
		})
	})

	Describe("Emit", func() {
		var (
			pingResponder emitter.RespondingByteEmitter
			testData      = []byte("hello")
		)

		BeforeEach(func() {
			pingResponder, _ = emitter.NewPingResponder(wrappedEmitter, origin)
		})

		It("delegates to the wrapped emitter", func() {
			pingResponder.Emit(testData)

			messages := wrappedEmitter.GetMessages()
			Expect(messages).To(HaveLen(1))
			Expect(messages[0]).To(Equal(testData))
		})

		It("increments the heartbeat counter", func() {
			pingResponder.Emit(testData)
			pingResponder.RespondToPing()

			Eventually(wrappedEmitter.GetMessages).Should(HaveLen(2))

			message := wrappedEmitter.GetMessages()[1]
			hbEnvelope := &events.Envelope{}
			err := proto.Unmarshal(message, hbEnvelope)
			Expect(err).NotTo(HaveOccurred())

			hbEvent := hbEnvelope.GetHeartbeat()

			Expect(hbEvent.GetReceivedCount()).To(Equal(uint64(1)))
		})
	})

	Describe("Close", func() {
		var pingResponder emitter.ByteEmitter

		BeforeEach(func() {
			pingResponder, _ = emitter.NewPingResponder(wrappedEmitter, origin)
		})

		It("eventually delegates to the inner heartbeat emitter", func() {
			pingResponder.Close()
			Eventually(wrappedEmitter.IsClosed).Should(BeTrue())
		})

		It("can be called more than once", func() {
			pingResponder.Close()
			Expect(pingResponder.Close).ToNot(Panic())
		})
	})

	Describe("RespondToPing", func() {
		var pingResponder emitter.RespondingByteEmitter

		BeforeEach(func() {
			pingResponder, _ = emitter.NewPingResponder(wrappedEmitter, origin)
		})

		It("creates a Heartbeat message", func() {
			pingResponder.RespondToPing()
			Expect(wrappedEmitter.GetMessages()).To(HaveLen(1))
			hbBytes := wrappedEmitter.GetMessages()[0]

			var heartbeat events.Envelope
			err := proto.Unmarshal(hbBytes, &heartbeat)
			Expect(err).NotTo(HaveOccurred())
		})

		It("logs an error when heartbeat emission fails", func() {
			wrappedEmitter.ReturnError = errors.New("fake error")

			logWriter := new(bytes.Buffer)
			log.SetOutput(logWriter)

			pingResponder.RespondToPing()

			loggedText := string(logWriter.Bytes())
			expectedText := "Problem while emitting heartbeat data: fake error"
			Expect(loggedText).To(ContainSubstring(expectedText))
			pingResponder.Close()
		})
	})
})
