package emitter

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"log"
)

type Event interface {
	ProtoMessage()
}

type Emitter interface {
	Emit(Event, events.Origin) error
}

var DefaultEmitter Emitter

func init() {
	udpEmitter, err := NewUdpEmitter()
	if err != nil {
		log.Printf("WARNING: failed to create udpEmitter: %v\n", err)
	}
	DefaultEmitter, err = NewInstrumentedEmitter(udpEmitter)
	if err != nil {
		log.Printf("WARNING: failed to create instrumentedEmitter: %v\n", err)
	}
}

func Emit(e Event, o events.Origin) error {
	if DefaultEmitter != nil {
		return DefaultEmitter.Emit(e, o)
	}

	return errors.New("Default emitter not set")
}
