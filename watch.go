package conform

import (
	"context"
	"reflect"
	"sync"
	"time"
)

// Watcher watches for configuration changes and reloads automatically
type Watcher[T any] struct {
	config   *T
	mu       sync.RWMutex
	opts     []Option
	ctx      context.Context
	cancel   context.CancelFunc
	onReload func(T)
}

// Watch creates a watcher that monitors configuration changes
func Watch[T any](onReload func(T), opts ...Option) (*Watcher[T], error) {
	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher[T]{
		opts:     opts,
		ctx:      ctx,
		cancel:   cancel,
		onReload: onReload,
	}

	// Load initial config
	cfg, err := LoadGeneric[T](opts...)
	if err != nil {
		cancel()
		return nil, err
	}

	w.mu.Lock()
	w.config = cfg
	w.mu.Unlock()

	// Start watching
	go w.watch()

	return w, nil
}

// Get returns the current configuration (thread-safe)
func (w *Watcher[T]) Get() T {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return *w.config
}

// Stop stops watching for changes
func (w *Watcher[T]) Stop() {
	w.cancel()
}

func (w *Watcher[T]) watch() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			// Reload config
			cfg, err := LoadGeneric[T](w.opts...)
			if err != nil {
				// Log error but don't update config
				continue
			}

			// Check if config changed
			w.mu.Lock()
			changed := !reflect.DeepEqual(w.config, cfg)
			if changed {
				w.config = cfg
				w.mu.Unlock()

				if w.onReload != nil {
					w.onReload(*cfg)
				}
			} else {
				w.mu.Unlock()
			}
		}
	}
}
