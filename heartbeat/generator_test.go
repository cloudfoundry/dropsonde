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

func (fds *fakeDataSource) GetData() events.Event {
	return &events.DropsondeStatus{}
}

var _ = Describe("HeartbeatGenerator", func() {
	Describe("GenerateHeartbeats", func() {
		It("periodically emits heartbeats", func() {
			fakeEmitter := emitter.NewFake()
			heartbeatDataSource := &fakeDataSource{}

			heartbeat.HeartbeatInterval = 10 * time.Millisecond

			stopChannel := make(chan interface{})
			heartbeatsStopped := make(chan interface{})
			go func() {
				heartbeat.HeartbeatGeneratingLoop(fakeEmitter, heartbeatDataSource, stopChannel)
				close(heartbeatsStopped)
			}()

			Eventually(func() int { return len(fakeEmitter.GetMessages()) }).Should(BeNumerically(">=", 2))
			close(stopChannel)
			Eventually(heartbeatsStopped).Should(BeClosed())
		})
	})

})
