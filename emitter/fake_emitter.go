package emitter

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
)

type envelope struct {
	Event  Event
	Origin events.Origin
}

type FakeEmitter struct {
	ReturnError bool
	Messages    []envelope
}

func NewFake() *FakeEmitter {
	return &FakeEmitter{}
}

func (f *FakeEmitter) Emit(e Event, origin events.Origin) (err error) {
	if f.ReturnError {
		f.ReturnError = false
		return errors.New("Returning error as requested")
	}
	f.Messages = append(f.Messages, envelope{e, origin})
	return
}
