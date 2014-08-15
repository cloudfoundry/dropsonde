package metrics

import (
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/autowire"
)

var metricSender *dropsonde.MetricSender

func init() {
	Initialize()
}

func Initialize() {
	metricSender = dropsonde.NewMetricSender(autowire.AutowiredEmitter())
}

func SendValue(name string, value float64, unit string) error {
	return metricSender.SendValue(name, value, unit)
}
