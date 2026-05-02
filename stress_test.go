package golog

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestConcurrentFileWriterRace(t *testing.T) {
	const goroutines = 100
	const writesPerGoroutine = 100

	path := filepath.Join(t.TempDir(), "test-concurrent-race.log")
	w, err := NewConcurrentFileWriter(path, BufferSize(1024))
	if err != nil {
		t.Fatal(err)
	}

	var expectedSize int64
	for i := 0; i < goroutines; i++ {
		for j := 0; j < writesPerGoroutine; j++ {
			expectedSize += int64(len(fmt.Sprintf("goroutine-%d-write-%d\n", i, j)))
		}
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				data := []byte(fmt.Sprintf("goroutine-%d-write-%d\n", id, j))
				if _, err := w.Write(data); err != nil {
					t.Errorf("Write failed: %v", err)
					return
				}
			}
		}(i)
	}

	wg.Wait()

	if err := w.Close(); err != nil {
		t.Error(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	if stat.Size() != expectedSize {
		t.Errorf("file size %d, expected %d", stat.Size(), expectedSize)
	}
}

func TestConcurrentFileWriterCloseRace(t *testing.T) {
	const iterations = 50

	dir := t.TempDir()
	for iter := 0; iter < iterations; iter++ {
		path := filepath.Join(dir, fmt.Sprintf("test-close-race-%d.log", iter))
		w, err := NewConcurrentFileWriter(path, BufferSize(1024))
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if _, err := w.Write([]byte("data\n")); err != nil {
					if !errors.Is(err, os.ErrClosed) {
						t.Errorf("Write failed: %v", err)
					}
					return
				}
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if _, err := w.Write([]byte("more\n")); err != nil {
					if !errors.Is(err, os.ErrClosed) {
						t.Errorf("Write failed: %v", err)
					}
					return
				}
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
			if err := w.Close(); err != nil {
				t.Errorf("Close failed: %v", err)
			}
		}()

		wg.Wait()
		if err := w.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	}
}

func TestConcurrentFileWriterMultipleClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test-multi-close.log")
	w, err := NewConcurrentFileWriter(path, BufferSize(1024))
	if err != nil {
		t.Fatal(err)
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			if err := w.Close(); err != nil {
				t.Errorf("Close failed: %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestBufferedFileWriterConcurrentWriteClose(t *testing.T) {
	const iterations = 50

	dir := t.TempDir()
	for iter := 0; iter < iterations; iter++ {
		path := filepath.Join(dir, fmt.Sprintf("test-buffered-race-%d.log", iter))
		w, err := NewBufferedFileWriter(path, BufferSize(1024))
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if _, err := w.Write([]byte("test data\n")); err != nil {
					if !errors.Is(err, os.ErrClosed) {
						t.Errorf("Write failed: %v", err)
					}
					return
				}
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
			if err := w.Close(); err != nil {
				t.Errorf("Close failed: %v", err)
			}
		}()

		wg.Wait()
		if err := w.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	}
}

func TestRotatingFileWriterConcurrentWriteClose(t *testing.T) {
	const iterations = 20

	dir := t.TempDir()
	for iter := 0; iter < iterations; iter++ {
		path := filepath.Join(dir, fmt.Sprintf("test-rotating-race-%d.log", iter))
		w, err := NewRotatingFileWriter(path, 1024, 2, BufferSize(512))
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				if _, err := w.Write([]byte("test data for rotation\n")); err != nil {
					if !errors.Is(err, os.ErrClosed) {
						t.Errorf("Write failed: %v", err)
					}
					return
				}
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(2 * time.Millisecond)
			if err := w.Close(); err != nil {
				t.Errorf("Close failed: %v", err)
			}
		}()

		wg.Wait()
		if err := w.Close(); err != nil {
			t.Errorf("Close failed: %v", err)
		}
	}
}

func TestFastTimerConcurrentStartStop(t *testing.T) {
	const goroutines = 100
	const iterations = 10

	for iter := 0; iter < iterations; iter++ {
		var wg sync.WaitGroup
		wg.Add(goroutines * 2)

		for i := 0; i < goroutines; i++ {
			go func() {
				defer wg.Done()
				StartFastTimer()
			}()
			go func() {
				defer wg.Done()
				StopFastTimer()
			}()
		}

		wg.Wait()
	}

	StopFastTimer()
}

func TestConcurrentFileWriterHighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping high load test in short mode")
	}

	const goroutines = 200
	const writesPerGoroutine = 1000
	const dataSize = 100

	path := filepath.Join(t.TempDir(), "test-high-load.log")
	w, err := NewConcurrentFileWriter(path, BufferSize(8192))
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, dataSize)
	for i := 0; i < dataSize-1; i++ {
		data[i] = 'x'
	}
	data[dataSize-1] = '\n'

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				if _, err := w.Write(data); err != nil {
					t.Errorf("Write failed: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	if err := w.Close(); err != nil {
		t.Error(err)
	}

	totalWrites := goroutines * writesPerGoroutine
	writesPerSec := float64(totalWrites) / elapsed.Seconds()
	t.Logf("Completed %d writes in %v (%.0f writes/sec)", totalWrites, elapsed, writesPerSec)

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	expectedSize := int64(goroutines * writesPerGoroutine * dataSize)
	if stat.Size() != expectedSize {
		t.Errorf("file size %d, expected %d", stat.Size(), expectedSize)
	}
}

func TestBufferedFileWriterHighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping high load test in short mode")
	}

	const writes = 100000
	const dataSize = 100

	path := filepath.Join(t.TempDir(), "test-buffered-high-load.log")
	w, err := NewBufferedFileWriter(path, BufferSize(8192))
	if err != nil {
		t.Fatal(err)
	}

	data := make([]byte, dataSize)
	for i := 0; i < dataSize-1; i++ {
		data[i] = 'x'
	}
	data[dataSize-1] = '\n'

	start := time.Now()
	for i := 0; i < writes; i++ {
		if _, err := w.Write(data); err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}
	elapsed := time.Since(start)

	if err := w.Close(); err != nil {
		t.Error(err)
	}

	writesPerSec := float64(writes) / elapsed.Seconds()
	t.Logf("Completed %d writes in %v (%.0f writes/sec)", writes, elapsed, writesPerSec)

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	expectedSize := int64(writes * dataSize)
	if stat.Size() != expectedSize {
		t.Errorf("file size %d, expected %d", stat.Size(), expectedSize)
	}
}
