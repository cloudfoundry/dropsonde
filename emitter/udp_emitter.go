package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"net"
)

type udpEmitter struct {
	udpAddr *net.UDPAddr
	udpConn net.PacketConn
	origin  *events.Origin
}

var DefaultAddress = "localhost:42420"

func NewUdpEmitter() (emitter Emitter, err error) {
	addr, err := net.ResolveUDPAddr("udp", DefaultAddress)
	if err != nil {
		return
	}

	conn, err := net.ListenPacket("udp", "")
	if err != nil {
		return
	}

	emitter = &udpEmitter{udpAddr: addr, udpConn: conn}
	return
}

func (e *udpEmitter) Emit(event events.Event) (err error) {
	envelope, err := Wrap(event, e.origin)
	if err != nil {
		return
	}
	data, err := proto.Marshal(envelope)
	if err != nil {
		return
	}

	_, err = e.udpConn.WriteTo(data, e.udpAddr)
	return
}

func (e *udpEmitter) SetOrigin(origin *events.Origin) {
	e.origin = origin
}

func (e *udpEmitter) Close() {
	e.udpConn.Close()
}
