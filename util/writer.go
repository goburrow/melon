package util

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/goburrow/gol"
)

var (
	writerLoggerName = "gomelon/util/writer"
)

// AsyncWriter writes asynchronously to the given writers.
type AsyncWriter struct {
	// Timeout is maximum duration before timing out flush a channel.
	Timeout time.Duration

	writers []io.Writer
	chans   []chan []byte

	wg     sync.WaitGroup
	finish chan struct{}

	logger gol.Logger
}

var _ io.WriteCloser = (*AsyncWriter)(nil)

// NewAsyncWriter allocates and returns a new AsyncWriter.
// Start() must be called before writing data.
// IMPORTANT: don't use fmt.Fprintf() directly on this writer since the printf
// buffer might be freed/reused and the data is contaminated before the
// underline writers receive it. Use fmt.Sprintf() then Write() instead.
func NewAsyncWriter(bufferSize int, writers ...io.Writer) *AsyncWriter {
	a := &AsyncWriter{
		writers: writers,
		chans:   make([]chan []byte, len(writers)),
		finish:  make(chan struct{}),
		logger:  gol.GetLogger(writerLoggerName),

		Timeout: 10 * time.Second,
	}
	for i, _ := range a.writers {
		a.chans[i] = make(chan []byte, bufferSize)
	}
	return a
}

// Start starts gorotines for each writer.
func (a *AsyncWriter) Start() error {
	a.wg.Add(len(a.chans))
	for i, c := range a.chans {
		go a.listen(c, a.writers[i])
	}
	return nil
}

// Stop stops and wait until all goroutines exit.
func (a *AsyncWriter) Stop() error {
	close(a.finish)
	// TODO: close all channels.
	a.wg.Wait()
	return nil
}

// Write sends data to all writers and returns error if a channel if full.
func (a *AsyncWriter) Write(b []byte) (int, error) {
	for i, c := range a.chans {
		select {
		case c <- b:
			// Succeed.
		default:
			// This channel is full.
			// TODO: Retry policy.
			err := fmt.Errorf("util: writer channel full %T", a.writers[i])
			return 0, err
		}
	}
	return len(b), nil
}

func (a *AsyncWriter) Close() error {
	return a.Stop()
}

// Listen retrieves data from the channel and write to the writer.
func (a *AsyncWriter) listen(c chan []byte, w io.Writer) {
	defer a.wg.Done()

	for {
		select {
		case <-a.finish:
			a.flush(c, w)
			return
		case b := <-c:
			if _, err := w.Write(b); err != nil {
				a.logger.Error("error writing %T: %v", w, err)
			}
		}
	}
}

// Flush sends all pending data in the channel to writer or timeout after
// maximum of a.Timeout and the writer timeout.
func (a *AsyncWriter) flush(c chan []byte, w io.Writer) {
	timeout := time.After(a.Timeout)
	for {
		// Timeout channel has higher priority.
		select {
		case <-timeout:
			// Timed out.
			a.logger.Warn("timeout flushing %T", w)
			return
		default:
		}
		select {
		case b := <-c:
			// Succeed.
			if _, err := w.Write(b); err != nil {
				a.logger.Error("error writing %T: %v", w, err)
				return
			}
		default:
			// Channel is empty.
			return
		}
	}
}
