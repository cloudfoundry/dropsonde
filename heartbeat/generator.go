package heartbeat

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"time"
)

var HeartbeatInterval = 10 * time.Second
var HeartbeatEmitter emitter.Emitter

type HeartbeatEventSource interface {
	GetHeartbeatEvent() events.Event
}

func BeginGeneration(dataSource HeartbeatEventSource, origin *events.Origin) (chan<- interface{}, error) {
	if HeartbeatEmitter == nil {
		return nil, errors.New("HeartbeatEmitter not set")
	}

	HeartbeatEmitter.SetOrigin(origin)
	stopChannel := make(chan interface{})
	go heartbeatGeneratingLoop(HeartbeatEmitter, dataSource, stopChannel)
	return stopChannel, nil
}

/*
	Main heartbeat generation loop.
	Most applications will not want to use this directly, use BeginGeneration instead.
*/
func heartbeatGeneratingLoop(e emitter.Emitter, dataSource HeartbeatEventSource, stopChannel <-chan interface{}) {
	defer e.Close()

	timer := time.NewTimer(HeartbeatInterval)
	for {
		select {
		case <-stopChannel:
			return
		case <-timer.C:
			timer.Reset(HeartbeatInterval)

			event := dataSource.GetHeartbeatEvent()
			e.Emit(event)
		}
	}
}
