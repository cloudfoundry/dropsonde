package emitter

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"sync"
)

type instrumentedEmitter struct {
	concreteEmitter        Emitter
	mutex                  *sync.RWMutex
	ReceivedMetricsCounter uint64
	SentMetricsCounter     uint64
	ErrorCounter           uint64
}

func (emitter *instrumentedEmitter) Emit(event events.Event) (err error) {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	emitter.ReceivedMetricsCounter++

	err = emitter.concreteEmitter.Emit(event)
	if err != nil {
		emitter.ErrorCounter++
	} else {
		emitter.SentMetricsCounter++
	}

	return
}

func NewInstrumentedEmitter(concreteEmitter Emitter) (emitter Emitter, err error) {
	if concreteEmitter == nil {
		err = errors.New("Unable to create InstrumentedEmitter from nil emitter implementation")
		return
	}

	emitter = &instrumentedEmitter{concreteEmitter: concreteEmitter, mutex: &sync.RWMutex{}}
	return
}

func (emitter *instrumentedEmitter) SetOrigin(origin *events.Origin) {
	emitter.concreteEmitter.SetOrigin(origin)
}

func (emitter *instrumentedEmitter) Close() {
	emitter.concreteEmitter.Close()
}

func (emitter *instrumentedEmitter) GetHeartbeatEvent() events.Event {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	return events.NewHeartbeat(emitter.SentMetricsCounter, emitter.ReceivedMetricsCounter, emitter.ErrorCounter)
}
