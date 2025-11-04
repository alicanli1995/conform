package conform

import (
	"context"
	"reflect"
	"sync"
	"time"
)

type Watcher[T any] struct {
	config   *T
	mu       sync.RWMutex
	opts     []Option
	ctx      context.Context
	cancel   context.CancelFunc
	onReload func(T)
}

func Watch[T any](onReload func(T), opts ...Option) (*Watcher[T], error) {
	ctx, cancel := context.WithCancel(context.Background())

	w := &Watcher[T]{
		opts:     opts,
		ctx:      ctx,
		cancel:   cancel,
		onReload: onReload,
	}

	cfg, err := LoadGeneric[T](opts...)
	if err != nil {
		cancel()
		return nil, err
	}

	w.mu.Lock()
	w.config = cfg
	w.mu.Unlock()

	go w.watch()

	return w, nil
}

func (w *Watcher[T]) Get() T {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return *w.config
}

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
			cfg, err := LoadGeneric[T](w.opts...)
			if err != nil {
				continue
			}

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
