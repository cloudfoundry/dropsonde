package emitter

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"sync"
)

type envelope struct {
	Event  events.Event
	Origin events.Origin
}

type FakeEmitter struct {
	ReturnError bool
	Messages    []envelope
	mutex       *sync.RWMutex
}

func NewFake() *FakeEmitter {
	return &FakeEmitter{mutex: new(sync.RWMutex)}
}
func (f *FakeEmitter) Emit(e events.Event, origin events.Origin) (err error) {

	if f.ReturnError {
		f.ReturnError = false
		return errors.New("Returning error as requested")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Messages = append(f.Messages, envelope{e, origin})
	return
}

func (f *FakeEmitter) GetMessages() (messages []envelope) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	messages = make([]envelope, len(f.Messages))
	copy(messages, f.Messages)
	return
}
