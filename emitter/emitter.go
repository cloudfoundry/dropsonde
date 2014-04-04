package emitter

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
)

type Emitter interface {
	Emit(events.Event) error
	SetOrigin(*events.Origin)
	Close()
}

var DefaultEmitter Emitter

func Emit(e events.Event) error {
	if DefaultEmitter != nil {
		return DefaultEmitter.Emit(e)
	}

	return errors.New("Default emitter not set")
}
