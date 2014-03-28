package emitter

type Fake struct {
	Messages []Event
}

func NewFake() *Fake {
	return new(Fake)
}

func (f *Fake) Emit(e Event) error {
	f.Messages = append(f.Messages, e)
	return nil
}
