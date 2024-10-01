package errgroup

import (
	"context"
	"sync"
)

// Group represents group of goroutines working on tasks concurrently.
type Group struct {
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
	cancel  func()
}

// WithContext creates new Group with context.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &Group{cancel: cancel}, ctx
}

// Go runs the provided function in a new goroutine.
func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}

// Wait waits for all goroutines to finish and returns first error.
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}
