package dropsonde_unmarshaller

import (
	metricNames "github.com/cloudfoundry/dropsonde/metrics"

	"github.com/cloudfoundry/dropsonde/events"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"sync"
)

// A DropsondeUnmarshallerCollection is a collection of DropsondeUnmarshaller instances.
type DropsondeUnmarshallerCollection interface {
	instrumentation.Instrumentable
	Run(inputChan <-chan []byte, outputChan chan<- *events.Envelope, waitGroup *sync.WaitGroup)
	Size() int
}

// NewDropsondeUnmarshallerCollection instantiates a DropsondeUnmarshallerCollection,
// creates the specified number of DropsondeUnmarshaller instances and logs to the
// provided logger.
func NewDropsondeUnmarshallerCollection(logger *gosteno.Logger, size int) DropsondeUnmarshallerCollection {
	var unmarshallers []DropsondeUnmarshaller
	for i := 0; i < size; i++ {
		unmarshallers = append(unmarshallers, NewDropsondeUnmarshaller(logger))
	}

	logger.Debugf("dropsondeUnmarshallerCollection: created %v unmarshallers", size)

	return &dropsondeUnmarshallerCollection{
		logger:        logger,
		unmarshallers: unmarshallers,
	}
}

type dropsondeUnmarshallerCollection struct {
	unmarshallers []DropsondeUnmarshaller
	logger        *gosteno.Logger
}

// Returns the number of unmarshallers in its collection.
func (u *dropsondeUnmarshallerCollection) Size() int {
	return len(u.unmarshallers)
}

// Run calls Run on each marshaller in its collection.
// This is done in separate go routines.
func (u *dropsondeUnmarshallerCollection) Run(inputChan <-chan []byte, outputChan chan<- *events.Envelope, waitGroup *sync.WaitGroup) {
	for _, unmarshaller := range u.unmarshallers {
		go func(um DropsondeUnmarshaller) {
			defer waitGroup.Done()
			um.Run(inputChan, outputChan)
		}(unmarshaller)
	}
}

// Emit returns the current metrics the DropsondeMarshallerCollection keeps about itself.
func (u *dropsondeUnmarshallerCollection) Emit() instrumentation.Context {
	return instrumentation.Context{
		Name:    "dropsondeUnmarshaller",
		Metrics: u.metrics(),
	}
}

func (u *dropsondeUnmarshallerCollection) metrics() []instrumentation.Metric {
	var internalMetrics []instrumentation.Metric
	for _, u := range u.unmarshallers {
		internalMetrics = append(internalMetrics, u.Emit().Metrics...)
	}

	metricsByName := make(map[string][]instrumentation.Metric)
	for _, metric := range internalMetrics {
		metricsEntry := metricsByName[metric.Name]
		metricsByName[metric.Name] = append(metricsEntry, metric)
	}

	var metrics []instrumentation.Metric
	metrics = concatTotalLogMessages(metricsByName, metrics)
	metrics = concatLogMessagesReceivedPerApp(metricsByName, metrics)
	metrics = concatOtherEventTypes(metricsByName, metrics)

	return metrics
}

func concatTotalLogMessages(metricsByName map[string][]instrumentation.Metric, metrics []instrumentation.Metric) []instrumentation.Metric {
	totalLogs := uint64(0)
	for _, metric := range metricsByName[metricNames.LogMessageTotal] {
		totalLogs += metric.Value.(uint64)
	}

	return append(metrics, instrumentation.Metric{Name: metricNames.LogMessageTotal, Value: totalLogs})
}

func concatLogMessagesReceivedPerApp(metricsByName map[string][]instrumentation.Metric, metrics []instrumentation.Metric) []instrumentation.Metric {
	logsReceivedPerApp := make(map[string]uint64)
	for _, metric := range metricsByName[metricNames.LogMessageReceived] {
		appId := metric.Tags[metricNames.AppIdTag].(string)
		logsReceivedPerApp[appId] += metric.Value.(uint64)
	}

	for appId, count := range logsReceivedPerApp {
		tags := make(map[string]interface{})
		tags[metricNames.AppIdTag] = appId
		metrics = append(metrics, instrumentation.Metric{Name: metricNames.LogMessageReceived, Value: count, Tags: tags})
	}

	return metrics
}

func concatOtherEventTypes(metricsByName map[string][]instrumentation.Metric, metrics []instrumentation.Metric) []instrumentation.Metric {
	metricsByEventType := make(map[string]uint64)

	for eventType, eventTypeMetrics := range metricsByName {
		if eventType == metricNames.LogMessageTotal || eventType == metricNames.LogMessageReceived {
			continue
		}

		for _, metric := range eventTypeMetrics {
			metricsByEventType[eventType] += metric.Value.(uint64)
		}
	}

	for eventType, count := range metricsByEventType {
		metrics = append(metrics, instrumentation.Metric{Name: eventType, Value: count})
	}

	return metrics
}
