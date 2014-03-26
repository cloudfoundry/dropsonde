package emitter

type Event interface {

}

type Emitter interface {
	Emit(Event)
}

var DefaultEmitter Emitter = &UdpEmitter{}

func Emit(e Event) {
	DefaultEmitter.Emit(e)
}
