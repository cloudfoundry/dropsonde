// package metricbatcher provides a mechanism to batch counter updates into a single event.
package metricbatcher

import (
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"sync"
	"time"
)

// MetricBatcher batches counter increment/add calls into periodic, aggregate events.
type MetricBatcher struct {
	metrics      map[string]uint64
	batchTicker  *time.Ticker
	metricSender metric_sender.MetricSender
	lock         sync.Mutex
}

// New instantiates a running MetricBatcher. Eventswill be emitted once per batchDuration. All
// updates to a given counter name will be combined into a single event and sent to metricSender.
func New(metricSender metric_sender.MetricSender, batchDuration time.Duration) *MetricBatcher {
	mb := &MetricBatcher{
		metrics:      make(map[string]uint64),
		batchTicker:  time.NewTicker(batchDuration),
		metricSender: metricSender,
	}

	go func() {
		for {
			<-mb.batchTicker.C
			mb.sendBatch()
		}
	}()

	return mb
}

func (mb *MetricBatcher) sendBatch() {
	localMetrics := mb.resetAndReturnMetrics()

	for name, delta := range localMetrics {
		mb.metricSender.AddToCounter(name, delta)
	}
}

// BatchIncrementCounter increments the named counter by 1, but does not immediately send a
// CounterEvent.
func (mb *MetricBatcher) BatchIncrementCounter(name string) {
	mb.BatchAddCounter(name, 1)
}

// BatchAddCounter increments the named counter by the provided delta, but does not
// immediately send a CounterEvent.
func (mb *MetricBatcher) BatchAddCounter(name string, delta uint64) {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	mb.metrics[name] += delta
}

// Reset clears the MetricBatcher's internal state, so that no counters are tracked.
func (mb *MetricBatcher) Reset() {
	mb.resetAndReturnMetrics()
}

func (mb *MetricBatcher) resetAndReturnMetrics() map[string]uint64 {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	localMetrics := mb.metrics

	mb.metrics = make(map[string]uint64, len(mb.metrics))

	return localMetrics
}
