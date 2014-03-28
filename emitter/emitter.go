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
	var err error
	DefaultEmitter, err = NewUdpEmitter()
	if err != nil {
		log.Printf("WARNING: failed to create default emitter: %v\n", err)
	}
}


func Emit(e Event) error {
	if DefaultEmitter != nil {
		return DefaultEmitter.Emit(e)
	}

	return errors.New("Default emitter not set")
}
