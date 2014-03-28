package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"errors"
	"github.com/cloudfoundry-incubator/dropsonde/events"
	"os"
	"strconv"
)

func Wrap(e Event) (*events.Envelope, error) {
	jobIndex, err := strconv.Atoi(os.Getenv("BOSH_JOB_INSTANCE"))
	if os.Getenv("BOSH_JOB_NAME") == "" || err != nil {
		return nil, errors.New("Event not emitted due to missing origin information")
	}

	origin := &events.Origin{
		JobName:       proto.String(os.Getenv("BOSH_JOB_NAME")),
		JobInstanceId: proto.Int(jobIndex),
	}
	envelope := &events.Envelope{Origin: origin}

	switch e.(type) {
	case *events.DropsondeStatus:
		envelope.EventType = events.Envelope_DropsondeStatus.Enum()
		envelope.DropsondeStatus = e.(*events.DropsondeStatus)
	case *events.HttpStart:
		envelope.EventType = events.Envelope_HttpStart.Enum()
		envelope.HttpStart = e.(*events.HttpStart)
	case *events.HttpStop:
		envelope.EventType = events.Envelope_HttpStop.Enum()
		envelope.HttpStop = e.(*events.HttpStop)
	default:
		return nil, errors.New("Cannot create envelope for unknown event type")
	}

	return envelope, nil
}
