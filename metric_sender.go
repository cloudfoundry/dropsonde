package dropsonde

import (
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
)

type MetricSender struct {
	eventEmitter emitter.EventEmitter
}

func NewMetricSender(eventEmitter emitter.EventEmitter) *MetricSender {
	return &MetricSender{eventEmitter: eventEmitter}
}

func (ms *MetricSender) SendValue(name string, value float64, unit string) error {
	return ms.eventEmitter.Emit(&events.ValueMetric{Name: &name, Value: &value, Unit: &unit})
}

func (ms *MetricSender) IncrementCounter(name string) error {
	return ms.eventEmitter.Emit(&events.CounterEvent{Name: &name})
}
