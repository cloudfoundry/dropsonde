package runtime_stats

import (
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/dropsonde/events"
	"log"
	"runtime"
	"time"
)

type RuntimeStats struct {
	eventEmitter emitter.EventEmitter
	interval     time.Duration
}

func NewRuntimeStats(eventEmitter emitter.EventEmitter, interval time.Duration) *RuntimeStats {
	return &RuntimeStats{
		eventEmitter: eventEmitter,
		interval:     interval,
	}
}

func (rs *RuntimeStats) Run(stopChan <-chan struct{}) {
	ticker := time.NewTicker(rs.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
		case <-stopChan:
			return
		}

		rs.emit("numCPUS", uint64(runtime.NumCPU()))
		rs.emit("numGoRoutines", uint64(runtime.NumGoroutine()))
		rs.emitMemMetrics()
	}
}

func (rs *RuntimeStats) emitMemMetrics() {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	rs.emit("memoryStats.numBytesAllocatedHeap", stats.HeapAlloc)
	rs.emit("memoryStats.numBytesAllocatedStack", stats.StackInuse)
	rs.emit("memoryStats.numBytesAllocated", stats.Alloc)
	rs.emit("memoryStats.numMallocs", stats.Mallocs)
	rs.emit("memoryStats.numFrees", stats.Frees)
	rs.emit("memoryStats.lastGCPauseTimeNS", stats.PauseNs[(stats.NumGC+255)%256])
}

func (rs *RuntimeStats) emit(name string, value uint64) {
	err := rs.eventEmitter.Emit(&events.ValueMetric{
		Name:  &name,
		Value: &value,
	})
	if err != nil {
		log.Printf("RuntimeStats: failed to emit: %v", err)
	}
}
