package dropsonde_marshaller

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/davecgh/go-spew/spew"
	"sync/atomic"
	"unicode"
)

type DropsondeMarshaller interface {
	instrumentation.Instrumentable
	Run(inputChan <-chan *events.Envelope, outputChan chan<- []byte)
}

func NewDropsondeMarshaller(logger *gosteno.Logger) DropsondeMarshaller {
	messageCounts := make(map[events.Envelope_EventType]*uint64)
	for key := range events.Envelope_EventType_name {
		var count uint64
		messageCounts[events.Envelope_EventType(key)] = &count
	}
	return &dropsondeMarshaller{
		logger:        logger,
		messageCounts: messageCounts,
	}
}

type dropsondeMarshaller struct {
	logger            *gosteno.Logger
	messageCounts     map[events.Envelope_EventType]*uint64
	marshalErrorCount uint64
}

func (u *dropsondeMarshaller) Run(inputChan <-chan *events.Envelope, outputChan chan<- []byte) {
	for message := range inputChan {

		messageBytes, err := proto.Marshal(message)
		if err != nil {
			u.logger.Errorf("dropsondeMarshaller: marshal error %v for message %v", err, message)
			incrementCount(&u.marshalErrorCount)
			continue
		}

		u.logger.Debugf("dropsondeMarshaller: marshalled message %v", spew.Sprintf("%v", message))

		u.incrementMessageCount(message.GetEventType())
		outputChan <- messageBytes
	}
}

func (u *dropsondeMarshaller) incrementMessageCount(eventType events.Envelope_EventType) {
	incrementCount(u.messageCounts[eventType])
}

func incrementCount(count *uint64) {
	atomic.AddUint64(count, 1)
}

func (m *dropsondeMarshaller) metrics() []instrumentation.Metric {
	var metrics []instrumentation.Metric

	for eventType, eventName := range events.Envelope_EventType_name {
		modifiedEventName := []rune(eventName)
		modifiedEventName[0] = unicode.ToLower(modifiedEventName[0])
		metricName := string(modifiedEventName) + "Marshalled"

		metricValue := atomic.LoadUint64(m.messageCounts[events.Envelope_EventType(eventType)])
		metrics = append(metrics, instrumentation.Metric{Name: metricName, Value: metricValue})
	}

	metrics = append(metrics, instrumentation.Metric{
		Name:  "marshalErrors",
		Value: atomic.LoadUint64(&m.marshalErrorCount),
	})

	return metrics
}

func (m *dropsondeMarshaller) Emit() instrumentation.Context {
	return instrumentation.Context{
		Name:    "dropsondeMarshaller",
		Metrics: m.metrics(),
	}
}
