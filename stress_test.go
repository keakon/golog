package golog

import (
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

	path := filepath.Join(os.TempDir(), "test-concurrent-race.log")
	w, err := NewConcurrentFileWriter(path, BufferSize(1024))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

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

	minExpectedSize := int64(goroutines * writesPerGoroutine * 20)
	if stat.Size() < minExpectedSize {
		t.Errorf("file size %d is too small, expected at least %d", stat.Size(), minExpectedSize)
	}
}

func TestConcurrentFileWriterCloseRace(t *testing.T) {
	const iterations = 50

	for iter := 0; iter < iterations; iter++ {
		path := filepath.Join(os.TempDir(), fmt.Sprintf("test-close-race-%d.log", iter))
		w, err := NewConcurrentFileWriter(path, BufferSize(1024))
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				w.Write([]byte("data\n"))
			}
		}()

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				w.Write([]byte("more\n"))
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
			w.Close()
		}()

		wg.Wait()
		w.Close()
		os.Remove(path)
	}
}

func TestConcurrentFileWriterMultipleClose(t *testing.T) {
	path := filepath.Join(os.TempDir(), "test-multi-close.log")
	w, err := NewConcurrentFileWriter(path, BufferSize(1024))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			w.Close()
		}()
	}

	wg.Wait()
}

func TestBufferedFileWriterConcurrentWriteClose(t *testing.T) {
	const iterations = 50

	for iter := 0; iter < iterations; iter++ {
		path := filepath.Join(os.TempDir(), fmt.Sprintf("test-buffered-race-%d.log", iter))
		w, err := NewBufferedFileWriter(path, BufferSize(1024))
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				w.Write([]byte("test data\n"))
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(time.Millisecond)
			w.Close()
		}()

		wg.Wait()
		w.Close()
		os.Remove(path)
	}
}

func TestRotatingFileWriterConcurrentWriteClose(t *testing.T) {
	const iterations = 20

	for iter := 0; iter < iterations; iter++ {
		path := filepath.Join(os.TempDir(), fmt.Sprintf("test-rotating-race-%d.log", iter))
		w, err := NewRotatingFileWriter(path, 1024, 2, BufferSize(512))
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				w.Write([]byte("test data for rotation\n"))
			}
		}()

		go func() {
			defer wg.Done()
			time.Sleep(2 * time.Millisecond)
			w.Close()
		}()

		wg.Wait()
		w.Close()
		os.Remove(path)
		os.Remove(path + ".1")
		os.Remove(path + ".2")
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

	path := filepath.Join(os.TempDir(), "test-high-load.log")
	w, err := NewConcurrentFileWriter(path, BufferSize(8192))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

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

	path := filepath.Join(os.TempDir(), "test-buffered-high-load.log")
	w, err := NewBufferedFileWriter(path, BufferSize(8192))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(path)

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
