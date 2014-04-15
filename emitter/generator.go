package emitter

import (
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"log"
	"time"
)

var HeartbeatInterval = 10 * time.Second

type HeartbeatEventSource interface {
	GetHeartbeatEvent() events.Event
}

func BeginGeneration(dataSource HeartbeatEventSource, destinationEmitter Emitter) (chan struct{}, error) {
	stopChannel := make(chan struct{})
	go heartbeatGeneratingLoop(destinationEmitter, dataSource, stopChannel)
	return stopChannel, nil
}

func heartbeatGeneratingLoop(e Emitter, dataSource HeartbeatEventSource, stopChannel <-chan struct{}) {
	defer e.Close()

	timer := time.NewTimer(HeartbeatInterval)
	for {
		select {
		case <-stopChannel:
			return
		case <-timer.C:
			timer.Reset(HeartbeatInterval)

			event := dataSource.GetHeartbeatEvent()
			err := e.Emit(event)
			if err != nil {
				log.Printf("Problem while emitting event: %v\n", err)
			}
		}
	}
}
