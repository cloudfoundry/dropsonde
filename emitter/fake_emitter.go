package emitter

import "errors"

type FakeEmitter struct {
	ReturnError bool
	Messages []Event
}

func NewFake() *FakeEmitter {
	return &FakeEmitter{}
}

func (f *FakeEmitter) Emit(e Event) (err error) {
	if f.ReturnError {
		f.ReturnError = false
		return errors.New("Returning error as requested")
	}
	f.Messages = append(f.Messages, e)
	return
}
