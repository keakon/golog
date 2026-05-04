package golog

import (
	"runtime"
	"sync"
	_ "unsafe"
)

// callers is linkname'd to runtime.callers to declare //go:noescape, which keeps the
// caller's [1]uintptr buffer on the stack. The public runtime.Callers cannot be
// annotated this way from outside the runtime package, so it would force the array
// to escape to the heap, costing one allocation per Caller() invocation. This is the
// only purpose of the linkname.
//
// If a future Go version blocks this linkname:
//   - Build with -ldflags=-checklinkname=0 (Go 1.23+).
//   - Or fall back to runtime.Callers and accept the heap allocation.
//
//go:noescape
//go:linkname callers runtime.callers
func callers(skip int, pcbuf []uintptr) int

// frameCache memoizes runtime.Frame lookups by program counter. The set of PCs in a
// running Go program is bounded by the code segment size, so the map is effectively
// stable in size after warm-up. We use sync.Map because the workload is read-mostly
// once warm; see OPTIMIZATION_LESSONS.md (Optimization C) for why mutex-based caches
// regress concurrent throughput by ~10x here.
var frameCache sync.Map

// Caller returns the file path and line number of the caller, skipping the given
// number of stack frames (similar to runtime.Caller but cached).
//
// Inspired by https://zhuanlan.zhihu.com/p/403417640.
func Caller(skip int) (file string, line int) {
	rpc := [1]uintptr{}
	n := callers(skip+1, rpc[:])
	if n < 1 {
		return
	}

	pc := rpc[0]
	if f, ok := frameCache.Load(pc); ok {
		frame := f.(runtime.Frame)
		return frame.File, frame.Line
	}
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	frameCache.Store(pc, frame)
	return frame.File, frame.Line
}
