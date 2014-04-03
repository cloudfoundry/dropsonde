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
	Describe("BeginGeneration", func() {
		It("periodically emits heartbeats", func() {
			fakeEmitter := emitter.NewFake()
			heartbeatDataSource := &fakeDataSource{}

			heartbeat.HeartbeatInterval = 10 * time.Millisecond

			heartbeat.HeartbeatEmitter = fakeEmitter
			stopChannel := heartbeat.BeginGeneration(heartbeatDataSource, nil)

			Eventually(func() int { return len(fakeEmitter.GetMessages()) }).Should(BeNumerically(">=", 2))
			close(stopChannel)
		})

		It("closes the emitter after the stopChannel is closed", func() {
			fakeEmitter := emitter.NewFake()
			heartbeatDataSource := &fakeDataSource{}

			heartbeat.HeartbeatInterval = 10 * time.Millisecond

			heartbeat.HeartbeatEmitter = fakeEmitter
			stopChannel := heartbeat.BeginGeneration(heartbeatDataSource, nil)

			close(stopChannel)
			Eventually(func() bool { return fakeEmitter.IsClosed }).Should(BeTrue())
		})
	})
})
