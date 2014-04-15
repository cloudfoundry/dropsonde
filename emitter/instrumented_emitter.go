package emitter

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"sync"
)

type InstrumentedEmitter interface {
	Emitter
	GetHeartbeatEvent() events.Event
}

type instrumentedEmitter struct {
	concreteEmitter        Emitter
	mutex                  *sync.RWMutex
	ReceivedMetricsCounter uint64
	SentMetricsCounter     uint64
	ErrorCounter           uint64
}

func (emitter *instrumentedEmitter) Emit(event events.Event) error {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	emitter.ReceivedMetricsCounter++

	err := emitter.concreteEmitter.Emit(event)
	if err != nil {
		emitter.ErrorCounter++
	} else {
		emitter.SentMetricsCounter++
	}

	return err
}

func NewInstrumentedEmitter(concreteEmitter Emitter) (InstrumentedEmitter, error) {
	if concreteEmitter == nil {
		err := errors.New("Unable to create InstrumentedEmitter from nil emitter implementation")
		return nil, err
	}

	emitter := &instrumentedEmitter{concreteEmitter: concreteEmitter, mutex: &sync.RWMutex{}}
	return emitter, nil
}

func (emitter *instrumentedEmitter) Close() {
	emitter.concreteEmitter.Close()
}

func (emitter *instrumentedEmitter) GetHeartbeatEvent() events.Event {
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()

	return events.NewHeartbeat(emitter.SentMetricsCounter, emitter.ReceivedMetricsCounter, emitter.ErrorCounter)
}
