package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"net"
)

type UdpEmitter struct {
	udpAddr *net.UDPAddr
	udpConn net.PacketConn
}

var DefaultAddress = "localhost:42420"

func NewUdpEmitter() (emitter *UdpEmitter, err error) {
	addr, _ := net.ResolveUDPAddr("udp", DefaultAddress)
	conn, err := net.ListenPacket("udp", "")

	emitter = &UdpEmitter{udpAddr: addr, udpConn: conn}
	return
}

func (e *UdpEmitter) Emit(event Event) (err error) {
	envelope, err := Wrap(event)
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
