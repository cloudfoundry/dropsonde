package runtime_stats

import (
	"log"
	"runtime"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

type EventEmitter interface {
	Emit(events.Event) error
}

type RuntimeStats struct {
	emitter  EventEmitter
	interval time.Duration
}

func NewRuntimeStats(emitter EventEmitter, interval time.Duration) *RuntimeStats {
	return &RuntimeStats{
		emitter:  emitter,
		interval: interval,
	}
}

func (rs *RuntimeStats) Run(stopChan <-chan struct{}) {
	ticker := time.NewTicker(rs.interval)
	defer ticker.Stop()
	for {
		rs.emit("numCPUS", float64(runtime.NumCPU()))
		rs.emit("numGoRoutines", float64(runtime.NumGoroutine()))
		rs.emitMemMetrics()

		select {
		case <-ticker.C:
		case <-stopChan:
			return
		}
	}
}

func (rs *RuntimeStats) emitMemMetrics() {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	toEmit := map[string]float64{
		"memoryStats.numBytesAllocatedHeap":  float64(stats.HeapAlloc),
		"memoryStats.numBytesAllocatedStack": float64(stats.StackInuse),
		"memoryStats.numBytesAllocated":      float64(stats.Alloc),
		"memoryStats.numMallocs":             float64(stats.Mallocs),
		"memoryStats.numFrees":               float64(stats.Frees),
		"memoryStats.lastGCPauseTimeNS":      float64(stats.PauseNs[(stats.NumGC+255)%256]),
	}

	for metric, value := range toEmit {
		rs.emit(metric, value)
	}
}

func (rs *RuntimeStats) emit(name string, value float64) {
	err := rs.emitter.Emit(&events.ValueMetric{
		Name:  &name,
		Value: &value,
		Unit:  proto.String("count"),
	})
	if err != nil {
		log.Printf("RuntimeStats: failed to emit %s: %v", name, err)
	}
}
