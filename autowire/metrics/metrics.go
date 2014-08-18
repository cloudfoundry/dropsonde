// Package metrics provides a simple API for sending value and counter metrics
// through the dropsonde system.
//
// Use
//
// See the documentation for package autowire for details on configuring through
// environment variables.
//
// Import the package (note that you do not need to additionally import
// autowire). The package self-initializes; to send metrics use
//
//		metrics.SendValue(name, value, unit)
//
// for sending known quantities, and
//
//		metrics.IncrementCounter(name)
//
// to increment a counter. (Note that the value of the counter is maintained by
// the receiver of the counter events, not the application that includes this
// package.)
package metrics

import (
	"github.com/cloudfoundry/dropsonde"
	"github.com/cloudfoundry/dropsonde/autowire"
)

var metricSender *dropsonde.MetricSender

func init() {
	Initialize()
}

// Initialize prepares the metrics package for use with the automatic Emitter
// from dropsonde/autowire. This function is called by the package's init
// method, so should only be explicitly called to reset the default
// MetricSender, e.g. in tests.
func Initialize() {
	metricSender = dropsonde.NewMetricSender(autowire.AutowiredEmitter())
}

// SendValue sends a value event for the named metric. See
// http://metrics20.org/spec/#units for the specifications on allowed units.
func SendValue(name string, value float64, unit string) error {
	return metricSender.SendValue(name, value, unit)
}

// IncrementCounter sends an increment event for the named counter. Maintaining
// the value of the counter is the responsibility of the receiver of the event,
// not the process that includes this package.
func IncrementCounter(name string) error {
	return metricSender.IncrementCounter(name)
}
