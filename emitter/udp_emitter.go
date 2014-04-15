package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"net"
	"runtime"
)

type udpEmitter struct {
	udpAddr *net.UDPAddr
	udpConn     net.PacketConn
	origin      string
	stopChannel chan struct{}
}

func NewUdpEmitter(remoteAddr string, origin string) (Emitter, error) {
	addr, err := net.ResolveUDPAddr("udp4", remoteAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return nil, err
	}

	rawEmitter := &udpEmitter{udpAddr: addr, udpConn: conn, origin: origin}
	emitter, err := NewInstrumentedEmitter(rawEmitter)

	stopChannel, err := BeginGeneration(emitter, rawEmitter)
	rawEmitter.stopChannel = stopChannel
	if err != nil {
		return nil, err
	}
	runtime.SetFinalizer(emitter, func(e Emitter) { e.Close() })

	return emitter, err
}

func (e *udpEmitter) Emit(event events.Event) error {
	envelope, err := Wrap(event, e.origin)
	if err != nil {
		return err
	}
	data, err := proto.Marshal(envelope)
	if err != nil {
		return err
	}

	_, err = e.udpConn.WriteTo(data, e.udpAddr)
	return err
}

func (e *udpEmitter) Close() {
	select {
	case <-e.stopChannel:
	default:
		close(e.stopChannel)
	}

	e.udpConn.Close()
}
