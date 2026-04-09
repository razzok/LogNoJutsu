// Package native provides a registry for Go-native technique implementations.
// Techniques with type: go are dispatched through this registry instead of
// spawning a child process (powershell.exe / cmd.exe).
package native

import "sync"

// NativeResult is the outcome of a native Go technique execution.
// Its fields map directly to playbooks.ExecutionResult.Output/ErrorOutput/Success.
type NativeResult struct {
	Output      string
	ErrorOutput string
	Success     bool
}

// NativeFunc is the signature of a Go technique implementation.
type NativeFunc func() (NativeResult, error)

// CleanupFunc is the signature of a Go technique cleanup function.
type CleanupFunc func() error

// entry holds a technique's implementation and optional cleanup function.
type entry struct {
	fn      NativeFunc
	cleanup CleanupFunc
}

var (
	mu       sync.RWMutex
	registry = make(map[string]entry)
)

// Register stores a NativeFunc (and optional CleanupFunc) for the given technique ID.
// Calling Register with the same ID twice overwrites the previous registration.
func Register(id string, fn NativeFunc, cleanup CleanupFunc) {
	mu.Lock()
	defer mu.Unlock()
	registry[id] = entry{fn: fn, cleanup: cleanup}
}

// Lookup returns the NativeFunc registered for id, or nil if not registered.
func Lookup(id string) NativeFunc {
	mu.RLock()
	defer mu.RUnlock()
	e, ok := registry[id]
	if !ok {
		return nil
	}
	return e.fn
}

// LookupCleanup returns the CleanupFunc registered for id, or nil if not registered
// or if no cleanup was provided at registration time.
func LookupCleanup(id string) CleanupFunc {
	mu.RLock()
	defer mu.RUnlock()
	e, ok := registry[id]
	if !ok {
		return nil
	}
	return e.cleanup
}
