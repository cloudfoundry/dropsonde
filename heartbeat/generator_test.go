package heartbeat_test

import (
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"github.com/cloudfoundry-incubator/dropsonde/heartbeat"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

type fakeDataSource struct {
}

func (fds *fakeDataSource) GetHeartbeatEvent() events.Event {
	return events.NewTestEvent(42)
}

var _ = Describe("HeartbeatGenerator", func() {
	Describe("BeginGeneration", func() {
		var (
			fakeEmitter          *emitter.FakeEmitter
			heartbeatEventSource = &fakeDataSource{}
		)

		BeforeEach(func() {
			origin := events.NewOrigin("testHeartbeatEmitter", 0)
			fakeEmitter = emitter.NewFake(origin)

			heartbeat.HeartbeatInterval = 10 * time.Millisecond
		})

		Context("when HeartbeatEmitter is not set", func() {
			It("returns an error", func() {
				heartbeat.HeartbeatEmitter = nil
				stopChan, err := heartbeat.BeginGeneration(heartbeatEventSource)

				Expect(stopChan).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when HeartbeatEmitter is set", func() {
			BeforeEach(func() {
				heartbeat.HeartbeatEmitter = fakeEmitter
			})

			It("periodically emits heartbeats, and the emitter can be closed properly", func() {
				stopChannel, _ := heartbeat.BeginGeneration(heartbeatEventSource)

				Eventually(func() int { return len(fakeEmitter.GetMessages()) }).Should(BeNumerically(">=", 2))

				close(stopChannel)
				Eventually(fakeEmitter.IsClosed).Should(BeTrue())
			})
		})

	})
})
