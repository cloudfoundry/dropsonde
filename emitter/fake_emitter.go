package emitter

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"sync"
)

type envelope struct {
	Event  events.Event
	Origin *events.Origin
}

type FakeEmitter struct {
	ReturnError bool
	Messages    []envelope
	mutex       *sync.RWMutex
	Origin      *events.Origin
	isClosed    bool
}

func NewFake(origin *events.Origin) *FakeEmitter {
	return &FakeEmitter{mutex: new(sync.RWMutex), Origin: origin}
}
func (f *FakeEmitter) Emit(e events.Event) (err error) {

	if f.ReturnError {
		f.ReturnError = false
		return errors.New("Returning error as requested")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.Messages = append(f.Messages, envelope{e, f.Origin})
	return
}

func (f *FakeEmitter) GetMessages() (messages []envelope) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	messages = make([]envelope, len(f.Messages))
	copy(messages, f.Messages)
	return
}

func (f *FakeEmitter) Close() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.isClosed = true
}

func (f *FakeEmitter) IsClosed() bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.isClosed
}
