package dropsonde

import (
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"github.com/cloudfoundry-incubator/dropsonde/heartbeat"
	"log"
	"sync"
)

var heartbeatState struct {
	sync.Mutex
	stopChannel chan<- interface{}
}

func Initialize(origin *events.Origin) {
	if emitter.DefaultEmitter == nil {
		udpEmitter, err := emitter.NewUdpEmitter()
		if err != nil {
			log.Fatalf("WARNING: failed to create udpEmitter: %v\n", err)
		}

		emitter.DefaultEmitter, err = emitter.NewInstrumentedEmitter(udpEmitter)
		if err != nil {
			log.Fatalf("WARNING: failed to create instrumentedEmitter: %v\n", err)
		}
	}

	emitter.DefaultEmitter.SetOrigin(origin)

	heartbeatState.Lock()
	defer heartbeatState.Unlock()

	if heartbeatState.stopChannel != nil {
		return
	}

	if heartbeatEventSource, ok := emitter.DefaultEmitter.(heartbeat.HeartbeatEventSource); ok {
		heartbeatState.stopChannel = heartbeat.BeginGeneration(heartbeatEventSource, origin)
	}
}

func Cleanup() {
	heartbeatState.Lock()
	defer heartbeatState.Unlock()

	if heartbeatState.stopChannel != nil {
		close(heartbeatState.stopChannel)
		heartbeatState.stopChannel = nil
	}
}
