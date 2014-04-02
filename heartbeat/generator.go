package heartbeat

import (
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"time"
)

var MessageOrigin events.Origin
var HeartbeatInterval = 10 * time.Second

type HeartbeatDataSource interface {
	GetData() events.Event
}

func BeginGeneration(e emitter.Emitter, dataSource HeartbeatDataSource) (stopChannel chan interface{}) {
	stopChannel = make(chan interface{})
	go HeartbeatGeneratingLoop(e, dataSource, stopChannel)
	return
}

/*
	Main heartbeat generation loop.
	Most applications will not want to use this directly, use BeginGeneration instead.
*/
func HeartbeatGeneratingLoop(e emitter.Emitter, dataSource HeartbeatDataSource, stopChannel chan interface{}) {
	timer := time.NewTimer(HeartbeatInterval)
	for {
		select {
		case <-stopChannel:
			return
		case <-timer.C:
			timer.Reset(HeartbeatInterval)

			data := dataSource.GetData()
			e.Emit(data)
		}
	}
}
