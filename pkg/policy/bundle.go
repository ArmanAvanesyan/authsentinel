package policy

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
)

// BundleLoader loads WASM policy bundles from disk with a cache keyed by path and file mtime.
// When the file at path changes (mtime), the bundle is recompiled on next LoadBundle.
type BundleLoader struct {
	mu      sync.RWMutex
	runtime wazero.Runtime
	cache   map[string]cachedBundle
	fallback FallbackConfig
}

type cachedBundle struct {
	compiled wazero.CompiledModule
	modTime  time.Time
}

// NewBundleLoader creates a loader that caches compiled modules by path and mtime.
func NewBundleLoader(fallback FallbackConfig) *BundleLoader {
	return &BundleLoader{
		runtime:  wazero.NewRuntime(context.Background()),
		cache:    make(map[string]cachedBundle),
		fallback: fallback,
	}
}

// LoadBundle loads (or reuses cached) the WASM bundle at path and returns an Engine.
// If the file mtime has changed since the cached version, the bundle is recompiled.
func (b *BundleLoader) LoadBundle(path string) (Engine, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	modTime := info.ModTime()
	wasm, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	b.mu.Lock()
	entry, ok := b.cache[path]
	if ok && entry.modTime.Equal(modTime) {
		b.mu.Unlock()
		return b.instantiateEngine(entry.compiled)
	}
	compiled, err := b.runtime.CompileModule(context.Background(), wasm)
	if err != nil {
		b.mu.Unlock()
		return nil, err
	}
	b.cache[path] = cachedBundle{compiled: compiled, modTime: modTime}
	b.mu.Unlock()
	return b.instantiateEngine(compiled)
}

// instantiateEngine creates a WASMRuntime with the given compiled module (using the loader's runtime).
func (b *BundleLoader) instantiateEngine(compiled wazero.CompiledModule) (Engine, error) {
	w := NewWASMRuntimeWithRuntime(b.runtime, b.fallback)
	if err := w.loadCompiled(compiled); err != nil {
		return nil, err
	}
	return w, nil
}
