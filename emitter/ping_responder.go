package emitter

import (
	"log"
	"runtime"
	"sync"

	"code.google.com/p/gogoprotobuf/proto"
)

type pingResponder struct {
	instrumentedEmitter InstrumentedEmitter
	innerEmitter        ByteEmitter
	origin              string
	sync.Mutex
	closed bool
}

func NewPingResponder(byteEmitter ByteEmitter, origin string) (RespondingByteEmitter, error) {
	instrumentedEmitter, err := NewInstrumentedEmitter(byteEmitter)
	if err != nil {
		return nil, err
	}

	hbEmitter := &pingResponder{
		instrumentedEmitter: instrumentedEmitter,
		innerEmitter:        byteEmitter,
		origin:              origin,
	}

	runtime.SetFinalizer(hbEmitter, (*pingResponder).Close)

	return hbEmitter, nil
}

func (e *pingResponder) Emit(data []byte) error {
	return e.instrumentedEmitter.Emit(data)
}

func (e *pingResponder) Close() {
	e.Lock()
	defer e.Unlock()

	if e.closed {
		return
	}

	e.instrumentedEmitter.Close()
	e.closed = true
}

func (e *pingResponder) RespondToPing() {
	hbEvent := e.instrumentedEmitter.GetHeartbeatEvent()
	hbEnvelope, err := Wrap(hbEvent, e.origin)
	if err != nil {
		log.Printf("Failed to wrap heartbeat event: %v\n", err)
		return
	}

	hbData, err := proto.Marshal(hbEnvelope)
	if err != nil {
		log.Printf("Failed to marshal heartbeat event: %v\n", err)
		return
	}

	err = e.innerEmitter.Emit(hbData)
	if err != nil {
		log.Printf("Problem while emitting heartbeat data: %v\n", err)
	}
}
