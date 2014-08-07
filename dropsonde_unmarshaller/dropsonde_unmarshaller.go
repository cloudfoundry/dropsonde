package dropsonde_unmarshaller

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/davecgh/go-spew/spew"
	"sync/atomic"
	"unicode"
)

type DropsondeUnmarshaller interface {
	instrumentation.Instrumentable
	Run(inputChan <-chan []byte, outputChan chan<- *events.Envelope)
}

func NewDropsondeUnmarshaller(logger *gosteno.Logger) DropsondeUnmarshaller {
	receiveCounts := make(map[events.Envelope_EventType]*uint64)
	for key := range events.Envelope_EventType_name {
		var count uint64
		receiveCounts[events.Envelope_EventType(key)] = &count
	}

	return &dropsondeUnmarshaller{
		logger:        logger,
		receiveCounts: receiveCounts,
	}
}

type dropsondeUnmarshaller struct {
	logger              *gosteno.Logger
	receiveCounts       map[events.Envelope_EventType]*uint64
	unmarshalErrorCount uint64
}

func (u *dropsondeUnmarshaller) Run(inputChan <-chan []byte, outputChan chan<- *events.Envelope) {
	for message := range inputChan {
		envelope := &events.Envelope{}
		err := proto.Unmarshal(message, envelope)
		if err != nil {
			u.logger.Debugf("dropsondeUnmarshaller: unmarshal error %v for message %v", err, message)
			incrementCount(&u.unmarshalErrorCount)
			continue
		}

		u.logger.Debugf("dropsondeUnmarshaller: received message %v", spew.Sprintf("%v", envelope))

		u.incrementReceiveCount(envelope.GetEventType())
		outputChan <- envelope
	}
}

func (u *dropsondeUnmarshaller) incrementReceiveCount(eventType events.Envelope_EventType) {
	incrementCount(u.receiveCounts[eventType])
}

func incrementCount(count *uint64) {
	atomic.AddUint64(count, 1)
}

func (m *dropsondeUnmarshaller) metrics() []instrumentation.Metric {
	var metrics []instrumentation.Metric

	for eventType, eventName := range events.Envelope_EventType_name {
		modifiedEventName := []rune(eventName)
		modifiedEventName[0] = unicode.ToLower(modifiedEventName[0])
		metricName := string(modifiedEventName) + "Received"

		metricValue := atomic.LoadUint64(m.receiveCounts[events.Envelope_EventType(eventType)])
		metrics = append(metrics, instrumentation.Metric{Name: metricName, Value: metricValue})
	}

	metrics = append(metrics, instrumentation.Metric{
		Name:  "unmarshalErrors",
		Value: atomic.LoadUint64(&m.unmarshalErrorCount),
	})

	return metrics
}

func (m *dropsondeUnmarshaller) Emit() instrumentation.Context {
	return instrumentation.Context{
		Name:    "dropsondeUnmarshaller",
		Metrics: m.metrics(),
	}
}
