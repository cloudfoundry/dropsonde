package dropsonde_unmarshaller

import (
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
		go func() {
			defer waitGroup.Done()
			unmarshaller.Run(inputChan, outputChan)
		}()
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
	var metrics []instrumentation.Metric

	for _, u := range u.unmarshallers {
		internalMetrics = append(internalMetrics, u.Emit().Metrics...)
	}

	metricsByName := make(map[string][]instrumentation.Metric)
	for _, metric := range internalMetrics {
		metricsEntry := metricsByName[metric.Name]
		metricsByName[metric.Name] = append(metricsEntry, metric)
	}

	concatTotalLogMessages(&metricsByName, &metrics)
	concatLogMessagesReceivedPerApp(&metricsByName, &metrics)
	concatOtherEventTypes(&metricsByName, &metrics)

	return metrics
}

func concatTotalLogMessages(metricsByName *map[string][]instrumentation.Metric, metrics *[]instrumentation.Metric) {
	totalLogs := uint64(0)
	for _, metric := range (*metricsByName)["logMessageTotal"] {
		totalLogs += metric.Value.(uint64)
	}

	*metrics = append(*metrics, instrumentation.Metric{Name: "logMessageTotal", Value: totalLogs})
}

func concatLogMessagesReceivedPerApp(metricsByName *map[string][]instrumentation.Metric, metrics *[]instrumentation.Metric) {
	logsReceivedPerApp := make(map[string]uint64)
	for _, metric := range (*metricsByName)["logMessageReceived"] {
		appId := metric.Tags["appId"].(string)
		logsReceivedPerApp[appId] += metric.Value.(uint64)
	}

	for appId, count := range logsReceivedPerApp {
		tags := make(map[string]interface{})
		tags["appId"] = appId
		*metrics = append(*metrics, instrumentation.Metric{Name: "logMessageReceived", Value: count, Tags: tags})
	}
}

func concatOtherEventTypes(metricsByName *map[string][]instrumentation.Metric, metrics *[]instrumentation.Metric) {
	metricsByEventType := make(map[string]uint64)

	for eventType, eventTypeMetrics := range *metricsByName {
		if eventType == "logMessageTotal" || eventType == "logMessageReceived" {
			continue
		}

		for _, metric := range eventTypeMetrics {
			metricsByEventType[eventType] += metric.Value.(uint64)
		}
	}

	for eventType, count := range metricsByEventType {
		*metrics = append(*metrics, instrumentation.Metric{Name: eventType, Value: count})
	}
}
