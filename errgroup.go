package utilz

import "golang.org/x/sync/errgroup"

// RunBatchFunc is the function signature for RunBatch.
type RunBatchFunc func() error

// RunBatch runs all functions simultaneously and waits until
// execution has completed or an error is encountered.
func RunBatch(fn ...RunBatchFunc) error {
	eg := errgroup.Group{}
	for _, f := range fn {
		eg.Go(f)
	}
	return eg.Wait()
}
func GoFuncErrChan(f func() error) chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- f()
	}()
	return ch
}
