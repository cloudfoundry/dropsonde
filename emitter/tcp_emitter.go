package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"net"
)

type tcpEmitter struct {
	tcpAddr *net.TCPAddr
	origin  string
}

func NewTcpEmitter(remoteAddress string, origin string) (e Emitter, err error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", remoteAddress)
	if err != nil {
		return
	}

	e = &tcpEmitter{tcpAddr: tcpAddr, origin: origin}
	return
}

func (e *tcpEmitter) Emit(event events.Event) (err error) {
	envelope, err := Wrap(event, e.origin)
	if err != nil {
		return
	}
	data, err := proto.Marshal(envelope)
	if err != nil {
		return
	}
	tcpConn, err := net.DialTCP("tcp", nil, e.tcpAddr)
	if err != nil {
		return
	}
	defer tcpConn.Close()

	_, err = tcpConn.Write(data)
	return
}

func (e *tcpEmitter) Close() {
}
