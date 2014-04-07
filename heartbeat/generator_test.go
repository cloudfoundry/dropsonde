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
			jobName := "testHeartbeatEmitter"
			var jobIndex int32

			origin := events.Origin{JobName: &jobName, JobInstanceId: &jobIndex}
			fakeEmitter = emitter.NewFake(&origin)

			heartbeat.HeartbeatInterval = 10 * time.Millisecond
		})

		Context("when HeartbeatEmitter is not set", func() {
			It("returns an error", func() {
				heartbeat.HeartbeatEmitter = nil
				_, err := heartbeat.BeginGeneration(heartbeatEventSource, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when HeartbeatEmitter is set", func() {
			BeforeEach(func() {
				heartbeat.HeartbeatEmitter = fakeEmitter
			})

			It("periodically emits heartbeats", func() {
				stopChannel, _ := heartbeat.BeginGeneration(heartbeatEventSource, nil)
				defer close(stopChannel)

				Eventually(func() int { return len(fakeEmitter.GetMessages()) }).Should(BeNumerically(">=", 2))
			})

			It("closes the emitter after the stopChannel is closed", func() {
				stopChannel, _ := heartbeat.BeginGeneration(heartbeatEventSource, nil)

				close(stopChannel)
				Eventually(func() bool { return fakeEmitter.IsClosed }).Should(BeTrue())
			})
		})

	})
})
