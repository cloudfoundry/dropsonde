package emitter

import (
	"errors"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	"github.com/pivotal-golang/localip"
)

var ErrorMissingOrigin = errors.New("Event not emitted due to missing origin information")
var ErrorMissingDeployment = errors.New("Event not emitted due to missing deployment information")
var ErrorMissingJob = errors.New("Event not emitted due to missing job information")
var ErrorMissingIndex = errors.New("Event not emitted due to missing index information")
var ErrorUnknownEventType = errors.New("Cannot create envelope for unknown event type")

func Wrap(event events.Event, origin, deployment, job, index string) (*events.Envelope, error) {
	if origin == "" {
		return nil, ErrorMissingOrigin
	}
	if deployment == "" {
		return nil, ErrorMissingDeployment
	}
	if job == "" {
		return nil, ErrorMissingJob
	}
	if index == "" {
		return nil, ErrorMissingIndex
	}
	ip, err := localip.LocalIP()
	if err != nil {
		return nil, err
	}

	envelope := &events.Envelope{
		Origin:     proto.String(origin),
		Timestamp:  proto.Int64(time.Now().UnixNano()),
		Deployment: proto.String(deployment),
		Job:        proto.String(job),
		Index:      proto.String(index),
		Ip:         proto.String(ip),
	}

	switch event := event.(type) {
	case *events.HttpStart:
		envelope.EventType = events.Envelope_HttpStart.Enum()
		envelope.HttpStart = event
	case *events.HttpStop:
		envelope.EventType = events.Envelope_HttpStop.Enum()
		envelope.HttpStop = event
	case *events.HttpStartStop:
		envelope.EventType = events.Envelope_HttpStartStop.Enum()
		envelope.HttpStartStop = event
	case *events.ValueMetric:
		envelope.EventType = events.Envelope_ValueMetric.Enum()
		envelope.ValueMetric = event
	case *events.CounterEvent:
		envelope.EventType = events.Envelope_CounterEvent.Enum()
		envelope.CounterEvent = event
	case *events.LogMessage:
		envelope.EventType = events.Envelope_LogMessage.Enum()
		envelope.LogMessage = event
	case *events.ContainerMetric:
		envelope.EventType = events.Envelope_ContainerMetric.Enum()
		envelope.ContainerMetric = event
	default:
		return nil, ErrorUnknownEventType
	}

	return envelope, nil
}
