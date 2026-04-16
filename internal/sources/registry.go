package sources

import (
	"fmt"
	"sort"
	"sync"
)

// Registry is a thread-safe compile-in set of Source plugins. Sources
// are added via Register at startup (typically from a single bootstrap
// function in main) and looked up by name on every API call.
type Registry struct {
	mu      sync.RWMutex
	entries map[string]Source
}

// NewRegistry constructs an empty Registry.
func NewRegistry() *Registry {
	return &Registry{entries: map[string]Source{}}
}

// Register adds src to the registry. Duplicate names replace existing
// entries — callers control the order of Register calls and get last-
// writer-wins semantics.
func (r *Registry) Register(src Source) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[src.Name()] = src
}

// Get returns the source registered under name.
func (r *Registry) Get(name string) (Source, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	src, ok := r.entries[name]
	if !ok {
		return nil, fmt.Errorf("sources: unknown source %q", name)
	}
	return src, nil
}

// Names returns the sorted list of registered source names.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.entries))
	for k := range r.entries {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
