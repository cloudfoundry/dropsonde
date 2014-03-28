package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"net"
	"os"
	"errors"
)

type udpEmitter struct {
	udpAddr *net.UDPAddr
	udpConn net.PacketConn
}

var DefaultAddress = "localhost:42420"

func NewUdpEmitter() (emitter Emitter, err error) {
	if os.Getenv("BOSH_JOB_NAME") == "" || os.Getenv("BOSH_JOB_INSTANCE") == "" {
		err = errors.New("BOSH_JOB_NAME or BOSH_JOB_INSTANCE not set")
		return
	}

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

func (e *udpEmitter) Emit(event Event) (err error) {
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
