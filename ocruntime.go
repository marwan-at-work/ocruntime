package ocruntime

import (
	"context"
	"runtime"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

// Runtime Views
var (
	grM           = stats.Int64("num_gr", "number of goroutines", stats.UnitDimensionless)
	GoroutineView = &view.View{
		Name:        "process/cpu_goroutines",
		Description: "number of running goroutines",
		Measure:     grM,
		Aggregation: view.LastValue(),
	}

	haM           = stats.Int64("ham", "bytes allocated", stats.UnitBytes)
	HeapAllocView = &view.View{
		Name:        "process/heap_alloc",
		Description: "total bytes of allocated heap",
		Measure:     haM,
		Aggregation: view.LastValue(),
	}

	hsM            = stats.Int64("hsM", "bytes given by os", stats.UnitBytes)
	HeapSystemView = &view.View{
		Name:        "process/sys_heap",
		Description: "Bytes of heap memory obtained from the OS",
		Measure:     hsM,
		Aggregation: view.LastValue(),
	}

	pnsM        = stats.Int64("pnsM", "nanoseconds of recent pause", stats.UnitMilliseconds)
	PauseNSView = &view.View{
		Name:        "process/pause_ns",
		Description: "most recent stop the world duration",
		Measure:     pnsM,
		Aggregation: view.LastValue(),
	}
)

// Start starts an infinite loop that
// profiles the process and reports the views
func Start(ctx context.Context, dur time.Duration) (close func()) {
	ctx, cancel := context.WithCancel(ctx)
	go start(ctx, dur)
	return cancel
}

func start(ctx context.Context, dur time.Duration) {
	if dur < time.Second {
		dur = time.Second
	}
	ticker := time.NewTicker(dur)
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			ticker.Stop()
			return
		}
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		stats.Record(
			context.Background(),
			grM.M(int64(runtime.NumGoroutine())),
			haM.M(int64(ms.HeapAlloc)),
			hsM.M(int64(ms.Sys)),
			pnsM.M(int64(ms.PauseNs[(ms.NumGC+255)%256])),
		)
	}
}
