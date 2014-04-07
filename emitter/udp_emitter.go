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

func NewUdpEmitter(remoteAddr string, origin *events.Origin) (emitter Emitter, err error) {
	addr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		return
	}

	conn, err := net.ListenPacket("udp", "")
	if err != nil {
		return
	}

	emitter = &udpEmitter{udpAddr: addr, udpConn: conn, origin: origin}
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

func (e *udpEmitter) Close() {
	e.udpConn.Close()
}
