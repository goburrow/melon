package async

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/goburrow/gol"
)

func init() {
	// Mute logger
	gol.GetLogger("melon/server").(*gol.DefaultLogger).SetLevel(gol.Off)
}

// chanWriter is used for testing async writer,
type chanWriter chan []byte

func (c chanWriter) Write(b []byte) (int, error) {
	c <- b
	return len(b), nil
}

type slowWriter struct {
	writeTime time.Duration
}

func (w *slowWriter) Write(b []byte) (int, error) {
	time.Sleep(w.writeTime)
	return len(b), nil
}

type errorWriter struct{}

func (w *errorWriter) Write(b []byte) (int, error) {
	return 0, errors.New("error")
}

func TestWriter(t *testing.T) {
	buffers := [...]chanWriter{
		make(chanWriter),
		make(chanWriter),
		make(chanWriter),
	}

	writer := NewWriter(1, buffers[0], buffers[1], buffers[2])
	err := writer.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Stop()
	_, err = writer.Write([]byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	for _, buf := range buffers {
		select {
		case b := <-buf:
			if "data" != string(b) {
				t.Fatalf("unexpected data: %s", b)
			}
		}
	}
}

func TestWriterFlush(t *testing.T) {
	buffers := [...]*bytes.Buffer{
		&bytes.Buffer{},
		&bytes.Buffer{},
	}

	count := 5
	writer := NewWriter(count, buffers[0], buffers[1])
	err := writer.Start()
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= count; i++ {
		data := string('0' + i)
		_, err = writer.Write([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
	}
	err = writer.Stop()
	if err != nil {
		t.Fatal(err)
	}
	for _, buf := range buffers {
		if "12345" != buf.String() {
			t.Fatalf("unexpected data: %s", buf.String())
		}
	}
}

func TestWriterFull(t *testing.T) {
	sw := &slowWriter{10 * time.Millisecond}

	count := 3
	writer := NewWriter(count, sw)
	writer.DrainTimeout = 1 * time.Millisecond

	err := writer.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Stop()
	for i := 1; i <= count; i++ {
		data := string('0' + i)
		_, err = writer.Write([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err = writer.Write([]byte("full"))
	if err != nil {
		t.Fatal(err)
	}
}
