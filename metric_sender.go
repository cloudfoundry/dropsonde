package dropsonde

import (
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
)

// A MetricSender emits metric events.
type MetricSender struct {
	eventEmitter emitter.EventEmitter
}

// NewMetricSender instantiates a MetricSender with the given EventEmitter.
func NewMetricSender(eventEmitter emitter.EventEmitter) *MetricSender {
	return &MetricSender{eventEmitter: eventEmitter}
}

// SendValue sends a metric with the given name, value and unit. See
// http://metrics20.org/spec/#units for a specification of acceptable units.
// Returns an error if one occurs while sending the event.
func (ms *MetricSender) SendValue(name string, value float64, unit string) error {
	return ms.eventEmitter.Emit(&events.ValueMetric{Name: &name, Value: &value, Unit: &unit})
}

// IncrementCounter increments the named counter. Returns an error if one occurs
// while sending the event.
func (ms *MetricSender) IncrementCounter(name string) error {
	return ms.eventEmitter.Emit(&events.CounterEvent{Name: &name})
}
