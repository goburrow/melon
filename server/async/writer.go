/*
Package async provides generic asynchronous IO for Melon applications.
*/
package async

import (
	"io"
	"sync"
	"time"
)

// Writer writes asynchronously to the given writers.
type Writer struct {
	// DrainTimeout is maximum duration before timing out flush a channel.
	DrainTimeout time.Duration

	writers []io.Writer
	chans   []chan []byte

	wg     sync.WaitGroup
	finish chan struct{}
}

var _ io.WriteCloser = (*Writer)(nil)

// NewWriter allocates and returns a new Writer.
// Start() must be called before writing data.
// IMPORTANT: don't use fmt.Fprintf() directly on this writer since the printf
// buffer might be freed/reused and the data is contaminated before the
// underline writers receive it. Use fmt.Sprintf() then Write() instead.
func NewWriter(bufferSize int, writers ...io.Writer) *Writer {
	a := &Writer{
		writers: writers,
		chans:   make([]chan []byte, len(writers)),
		finish:  make(chan struct{}),

		DrainTimeout: 10 * time.Second,
	}
	for i := range a.writers {
		a.chans[i] = make(chan []byte, bufferSize)
	}
	return a
}

// Start starts gorotines for each writer.
func (a *Writer) Start() error {
	a.wg.Add(len(a.chans))
	for i, c := range a.chans {
		go a.listen(c, a.writers[i])
	}
	return nil
}

// Stop stops and wait until all goroutines exit.
func (a *Writer) Stop() error {
	close(a.finish)
	// TODO: close all channels.
	a.wg.Wait()
	return nil
}

// Write sends data to all writers and returns error if a channel if full.
func (a *Writer) Write(b []byte) (int, error) {
	for _, c := range a.chans {
		// FIXME: still blocked if channel buffer is full.
		c <- b
	}
	return len(b), nil
}

func (a *Writer) Close() error {
	return a.Stop()
}

// Listen retrieves data from the channel and write to the writer.
func (a *Writer) listen(c chan []byte, w io.Writer) {
	defer a.wg.Done()

	for {
		select {
		case <-a.finish:
			a.flush(c, w)
			return
		case b := <-c:
			if _, err := w.Write(b); err != nil {
				logger.Errorf("error writing %T: %v", w, err)
			}
		}
	}
}

// Flush sends all pending data in the channel to writer or timeout after
// maximum of a.Timeout and the writer timeout.
func (a *Writer) flush(c chan []byte, w io.Writer) {
	timeout := time.After(a.DrainTimeout)
	for {
		// Timeout channel has higher priority.
		select {
		case <-timeout:
			// Timed out.
			logger.Warnf("timeout flushing %T", w)
			return
		default:
		}
		select {
		case b := <-c:
			// Succeed.
			if _, err := w.Write(b); err != nil {
				logger.Errorf("error writing %T: %v", w, err)
				return
			}
		default:
			// Channel is empty.
			return
		}
	}
}
