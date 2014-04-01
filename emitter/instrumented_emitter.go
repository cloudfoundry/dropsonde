package emitter

import (
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"sync"
)

type InstrumentedEmitter struct {
	concreteEmitter        Emitter
	mutex                  *sync.RWMutex
	ReceivedMetricsCounter uint64
	SentMetricsCounter     uint64
	ErrorCounter           uint64
}

func (emitter *InstrumentedEmitter) Emit(event Event, origin events.Origin) (err error) {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	emitter.ReceivedMetricsCounter++

	err = emitter.concreteEmitter.Emit(event, origin)
	if err != nil {
		emitter.ErrorCounter++
	} else {
		emitter.SentMetricsCounter++
	}

	return
}

func NewInstrumentedEmitter(concreteEmitter Emitter) (emitter *InstrumentedEmitter, err error) {
	return &InstrumentedEmitter{concreteEmitter: concreteEmitter, mutex: &sync.RWMutex{}}, nil
}
