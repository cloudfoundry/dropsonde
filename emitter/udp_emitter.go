package emitter

import (
	"net"
	"code.google.com/p/gogoprotobuf/proto"
)

type UdpEmitter struct {
//	connection net.PacketConn
}

func (e *UdpEmitter) Emit(event Event) {
	envelope, err := Wrap(event)
	if err != nil {
		return
	}
	data, _ := proto.Marshal(envelope)
	addr, _ := net.ResolveUDPAddr("udp", ":42420")
	e.connection().WriteTo(data, addr)
}

func (e *UdpEmitter) connection() net.PacketConn {
	c, _ := net.DialUDP("udp", nil, nil)
	return c
}
