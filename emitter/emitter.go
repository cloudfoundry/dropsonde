package emitter

import (
	"errors"
	"log"
)

type Event interface {
	ProtoMessage()
}

type Emitter interface {
	Emit(Event) error
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

func Emit(e Event) error {
	if DefaultEmitter != nil {
		return DefaultEmitter.Emit(e)
	}

	return errors.New("Default emitter not set")
}
