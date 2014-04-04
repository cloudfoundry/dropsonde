package dropsonde

import (
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"github.com/cloudfoundry-incubator/dropsonde/heartbeat"
	"log"
	"sync"
)

const DefaultEmitterRemoteAddr = "localhost:42420"
const DefaultHeartbeatEmitterRemoteAddr = "localhost:42420"

var heartbeatState struct {
	sync.Mutex
	stopChannel chan<- interface{}
}

func Initialize(origin *events.Origin) (err error) {
	if emitter.DefaultEmitter == nil {
		udpEmitter, err := emitter.NewUdpEmitter(DefaultEmitterRemoteAddr)
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
		if heartbeat.HeartbeatEmitter == nil {
			heartbeat.HeartbeatEmitter, err = emitter.NewTcpEmitter(DefaultHeartbeatEmitterRemoteAddr)
			if err != nil {
				log.Fatalf("WARNING: failed to create tcpEmitter: %v\n", err)
			}
		}

		heartbeatState.stopChannel, err = heartbeat.BeginGeneration(heartbeatEventSource, origin)
		if err != nil {
			log.Fatalf("WARNING: failed to start HeartbeatGenerator: %v\n", err)
		}
	}

	return
}

func Cleanup() {
	heartbeatState.Lock()
	defer heartbeatState.Unlock()

	if heartbeatState.stopChannel != nil {
		close(heartbeatState.stopChannel)
		heartbeatState.stopChannel = nil
	}
}
