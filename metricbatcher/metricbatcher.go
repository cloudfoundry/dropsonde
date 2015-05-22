package metricbatcher

import (
	"github.com/cloudfoundry/dropsonde/metric_sender"
	"sync"
	"time"
)

type MetricBatcher struct {
	metrics      map[string]uint64
	batchTicker  *time.Ticker
	metricSender metric_sender.MetricSender
	lock         sync.RWMutex
}

func New(metricSender metric_sender.MetricSender, batchDuration time.Duration) *MetricBatcher {
	mb := &MetricBatcher{
		metrics:      make(map[string]uint64),
		batchTicker:  time.NewTicker(batchDuration),
		metricSender: metricSender,
	}

	go func() {
		for {
			<-mb.batchTicker.C
			mb.lock.Lock()

			for name, delta := range mb.metrics {
				metricSender.AddToCounter(name, delta)
			}
			mb.unsafeReset()

			mb.lock.Unlock()
		}
	}()

	return mb
}

func (mb *MetricBatcher) BatchIncrementCounter(name string) {
	mb.BatchAddCounter(name, 1)
}

func (mb *MetricBatcher) BatchAddCounter(name string, delta uint64) {
	mb.lock.Lock()
	defer mb.lock.Unlock()

	mb.metrics[name] += delta
}

func (mb *MetricBatcher) Reset() {
    mb.lock.Lock()
    defer mb.lock.Unlock()

    mb.unsafeReset()
}

func (mb *MetricBatcher) unsafeReset() {
    mb.metrics = make(map[string]uint64, len(mb.metrics))
}