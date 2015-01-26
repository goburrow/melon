package gows

import (
	"fmt"
	"net/http"
	"runtime"
)

// handleAdminRuntime displays runtime statistics.
func handleAdminRuntime(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "must-revalidate,no-cache,no-store")
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "NumCPU: %d\nNumCgoCall: %d\nNumGoroutine: %d\n", runtime.NumCPU(), runtime.NumCgoCall(), runtime.NumGoroutine())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// General statistics
	fmt.Fprintf(w, "MemStats:\n\tAlloc: %d\n\tTotalAlloc: %d\n\tSys: %d\n\tLookups: %d\n\tMallocs: %d\n\tFrees: %d\n", m.Alloc, m.TotalAlloc, m.Sys, m.Lookups, m.Mallocs, m.Frees)
	// Main allocation heap statistics
	fmt.Fprintf(w, "\tHeapAlloc: %d\n\tHeapSys: %d\n\tHeapIdle: %d\n\tHeapInuse: %d\n\tHeapReleased: %d\n\tHeapObjects: %d\n", m.HeapAlloc, m.HeapSys, m.HeapIdle, m.HeapInuse, m.HeapReleased, m.HeapObjects)
	// Low-level fixed-size structure allocator statistics
	fmt.Fprintf(w, "\tStackInuse: %d\n\tStackSys: %d\n\tMSpanInuse: %d\n\tMSpanSys: %d\n\tMCacheInuse: %d\n\tMCacheSys: %d\n\tBuckHashSys: %d\n\tGCSys: %d\n\tOtherSys: %d\n", m.StackInuse, m.StackSys, m.MSpanInuse, m.MSpanSys, m.MCacheInuse, m.MCacheSys, m.BuckHashSys, m.GCSys, m.OtherSys)
	// Garbage collector statistics
	fmt.Fprintf(w, "\tNextGC: %d\n\tLastGC: %d\n\tPauseTotalNs: %d\n\tNumGC: %d\n\tEnableGC: %t\n\tDebugGC: %t\n", m.NextGC, m.LastGC, m.PauseTotalNs, m.NumGC, m.EnableGC, m.DebugGC)

	fmt.Fprintf(w, "Version: %s\n", runtime.Version())
}
