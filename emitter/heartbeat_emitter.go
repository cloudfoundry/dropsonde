package emitter

import (
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var HeartbeatInterval = 10 * time.Second

func init() {
	intervalOverride, err := strconv.ParseFloat(os.Getenv("DROPSONDE_HEARTBEAT_INTERVAL_SECS"), 64)
	if err == nil {
		HeartbeatInterval = time.Duration(intervalOverride*1000) * time.Millisecond
	}
}

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
