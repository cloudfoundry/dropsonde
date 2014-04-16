package emitter

import (
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"log"
	"runtime"
	"sync"
	"time"
)

var HeartbeatInterval = 10 * time.Second

type heartbeatEmitter struct {
	instrumentedEmitter InstrumentedEmitter
	innerHbEmitter      Emitter
	stopChan            chan struct{}
	sync.Mutex
	closed bool
}

func NewHeartbeatEmitter(emitter Emitter) (Emitter, error) {
	instrumentedEmitter, err := NewInstrumentedEmitter(emitter)
	if err != nil {
		return nil, err
	}

	hbEmitter := &heartbeatEmitter{
		instrumentedEmitter: instrumentedEmitter,
		innerHbEmitter:      emitter,
		stopChan:            make(chan struct{}),
	}

	go hbEmitter.generateHeartbeats(HeartbeatInterval)
	runtime.SetFinalizer(hbEmitter, (*heartbeatEmitter).Close)

	return hbEmitter, nil
}

func (e *heartbeatEmitter) Emit(event events.Event) error {
	return e.instrumentedEmitter.Emit(event)
}

func (e *heartbeatEmitter) Close() {
	e.Lock()
	defer e.Unlock()

	if e.closed {
		return
	}

	e.closed = true
	close(e.stopChan)
}

func (e *heartbeatEmitter) generateHeartbeats(heartbeatInterval time.Duration) {
	defer e.instrumentedEmitter.Close()

	timer := time.NewTimer(heartbeatInterval)
	for {
		select {
		case <-e.stopChan:
			return
		case <-timer.C:
			timer.Reset(heartbeatInterval)

			event := e.instrumentedEmitter.GetHeartbeatEvent()
			err := e.innerHbEmitter.Emit(event)
			if err != nil {
				log.Printf("Problem while emitting heartbeat event: %v\n", err)
			}
		}
	}
}
