package emitter

type Event interface {
	ProtoMessage()
}

type Emitter interface {
	Emit(Event)
}

var DefaultEmitter Emitter = &UdpEmitter{}

func Emit(e Event) {
	DefaultEmitter.Emit(e)
}
