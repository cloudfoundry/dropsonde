package dropsonde

import (
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/emitter"
)

var DefaultEmitterRemoteAddr = "localhost:42420"

func Initialize(origin string) error {
	if len(origin) == 0 {
		return errors.New("Cannot initialize dropsonde without an origin")
	}

	udpEmitter, err := emitter.NewUdpEmitter(DefaultEmitterRemoteAddr, origin)
	if err != nil {
		return err
	}

	e, err := emitter.NewHeartbeatEmitter(udpEmitter)
	if err != nil {
		return err
	}

	emitter.DefaultEmitter = e

	return nil
}
