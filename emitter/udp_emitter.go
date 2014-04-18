package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"net"
)

type udpEmitter struct {
	udpAddr *net.UDPAddr
	udpConn net.PacketConn
	origin  string
}

func NewUdpEmitter(remoteAddr string, origin string) (*udpEmitter, error) {
	addr, err := net.ResolveUDPAddr("udp4", remoteAddr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenPacket("udp4", "")
	if err != nil {
		return nil, err
	}

	emitter := &udpEmitter{udpAddr: addr, udpConn: conn, origin: origin}
	return emitter, nil
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
	e.udpConn.Close()
}
