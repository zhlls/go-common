package utils

import (
	"errors"
	"time"
)

var ErrTimeout = errors.New("func timeout")

// DoWithTimeout runs f and returns its error.  If the deadline d elapses first,
// it returns a grpc DeadlineExceeded error instead.
func DoWithTimeout(f func() error, d time.Duration) error {
	errChan := make(chan error, 1)
	go func () {
		errChan <- f()
		close(errChan)
	}()
	t := time.NewTimer(d)
	select {
	case <-t.C:
		return ErrTimeout
	case err := <-errChan:
		if !t.Stop() {
			<-t.C
		}
		return err
	}
}
