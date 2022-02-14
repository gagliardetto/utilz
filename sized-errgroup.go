package utilz

import (
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

func NewSizedGroup(cNum int64) *SizedGroup {
	gr := new(SizedGroup)
	gr.sem = semaphore.NewWeighted(cNum)
	return gr
}

type SizedGroup struct {
	cancel func()

	wg  sync.WaitGroup
	sem *semaphore.Weighted

	errOnce sync.Once
	err     error
}

func SizedGroupWithContext(ctx context.Context) (*SizedGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &SizedGroup{cancel: cancel}, ctx
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (g *SizedGroup) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

// Go calls the given function in a new goroutine.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (g *SizedGroup) Go(f func() error) {
	g.wg.Add(1)

	if err := g.sem.Acquire(context.Background(), 1); err != nil {
		panic(err)
	}
	go func() {
		defer g.wg.Done()
		defer g.sem.Release(1)

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
