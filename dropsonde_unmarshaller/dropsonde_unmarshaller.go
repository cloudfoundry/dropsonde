// Package dropsonde_unmarshaller provides a tool for unmarshalling Envelopes
// from Protocol Buffer messages.
//
// Use
//
// Instantiate a Marshaller and run it:
//
//		unmarshaller := dropsonde_unmarshaller.NewDropsondeUnMarshaller(logger)
//		inputChan :=  make(chan []byte) // or use a channel provided by some other source
//		outputChan := make(chan *events.Envelope)
//		go unmarshaller.Run(inputChan, outputChan)
//
// The unmarshaller self-instruments, counting the number of messages
// processed and the number of errors. These can be accessed through the Emit
// function on the unmarshaller.
package dropsonde_unmarshaller

import (
	"unicode"

	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/davecgh/go-spew/spew"
	"github.com/gogo/protobuf/proto"
)

// A DropsondeUnmarshaller is an self-instrumenting tool for converting Protocol
// Buffer-encoded dropsonde messages to Envelope instances.
type DropsondeUnmarshaller interface {
	Run(inputChan <-chan []byte, outputChan chan<- *events.Envelope)
	UnmarshallMessage([]byte) (*events.Envelope, error)
}

// NewDropsondeUnmarshaller instantiates a DropsondeUnmarshaller and logs to the
// provided logger.
func NewDropsondeUnmarshaller(logger *gosteno.Logger) DropsondeUnmarshaller {
	return &dropsondeUnmarshaller{
		logger: logger,
	}
}

type dropsondeUnmarshaller struct {
	logger *gosteno.Logger
}

// Run reads byte slices from inputChan, unmarshalls them to Envelopes, and
// emits the Envelopes onto outputChan. It operates one message at a time, and
// will block if outputChan is not read.
func (u *dropsondeUnmarshaller) Run(inputChan <-chan []byte, outputChan chan<- *events.Envelope) {
	for message := range inputChan {
		envelope, err := u.UnmarshallMessage(message)
		if err != nil {
			continue
		}
		outputChan <- envelope
	}
}

func (u *dropsondeUnmarshaller) UnmarshallMessage(message []byte) (*events.Envelope, error) {
	envelope := &events.Envelope{}
	err := proto.Unmarshal(message, envelope)
	if err != nil {
		u.logger.Debugf("dropsondeUnmarshaller: unmarshal error %v for message %v", err, message)
		metrics.BatchIncrementCounter("dropsondeUnmarshaller.unmarshalErrors")
		return nil, err
	}

	u.logger.Debugf("dropsondeUnmarshaller: received message %v", spew.Sprintf("%v", envelope))

	if envelope.GetEventType() == events.Envelope_LogMessage {
		u.incrementLogMessageReceiveCount(envelope.GetLogMessage().GetAppId())
	} else {
		u.incrementReceiveCount(envelope.GetEventType())
	}

	return envelope, nil
}

func (u *dropsondeUnmarshaller) incrementLogMessageReceiveCount(appID string) {
	metrics.BatchIncrementCounter("dropsondeUnmarshaller.logMessageTotal")
}

func (u *dropsondeUnmarshaller) incrementReceiveCount(eventType events.Envelope_EventType) {
	name, ok := events.Envelope_EventType_name[int32(eventType)]

	if !ok {
		name = "unknownEventType"
	}

	modifiedEventName := []rune(name)
	modifiedEventName[0] = unicode.ToLower(modifiedEventName[0])
	metricName := string(modifiedEventName) + "Received"

	metrics.BatchIncrementCounter("dropsondeUnmarshaller." + metricName)
}
