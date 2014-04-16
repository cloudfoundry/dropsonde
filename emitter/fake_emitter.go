package emitter

import (
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"sync"
)

type envelope struct {
	Event  events.Event
	Origin string
}

type FakeEmitter struct {
	ReturnError error
	Messages    []envelope
	mutex       *sync.RWMutex
	Origin      string
	isClosed    bool
}

func NewFake(origin string) *FakeEmitter {
	return &FakeEmitter{mutex: new(sync.RWMutex), Origin: origin}
}
func (f *FakeEmitter) Emit(e events.Event) (err error) {

	if f.ReturnError != nil {
		err = f.ReturnError
		f.ReturnError = nil
		return
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
