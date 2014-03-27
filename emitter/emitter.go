package emitter

import "os"

type Event interface {
	ProtoMessage()
}

type Emitter interface {
	Emit(Event)
}

var DefaultEmitter Emitter

func init() {
	if os.Getenv("BOSH_JOB_NAME") == "" || os.Getenv("BOSH_JOB_INSTANCE") == "" {
		// warnings on stdout or stderr?
	}
	DefaultEmitter = new(UdpEmitter)
}

func Emit(e Event) {
	DefaultEmitter.Emit(e)
}
