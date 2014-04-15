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

func NewTcpEmitter(remoteAddress string, origin string) (Emitter, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", remoteAddress)
	if err != nil {
		return nil, err
	}

	e := &tcpEmitter{tcpAddr: tcpAddr, origin: origin}
	return e, nil
}

func (e *tcpEmitter) Emit(event events.Event) (error) {
	envelope, err := Wrap(event, e.origin)
	if err != nil {
		return err
	}
	data, err := proto.Marshal(envelope)
	if err != nil {
		return err
	}
	tcpConn, err := net.DialTCP("tcp", nil, e.tcpAddr)
	if err != nil {
		return err
	}
	defer tcpConn.Close()

	_, err = tcpConn.Write(data)
	return err
}

func (e *tcpEmitter) Close() {
}
