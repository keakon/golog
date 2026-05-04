package golog

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"
)

const dateTimeBufSize = 10 // length of "YYYY-mm-DD"

// now is the package-level clock. It is replaced by tests via setNowFunc.
var now = time.Now

type fastTimerBeforeStoreHook func()

// fastTimerHook is a test hook used to pause update() right before publishing a
// snapshot so tests can exercise stop/start lifecycle edges deterministically.
var fastTimerHook atomic.Pointer[fastTimerBeforeStoreHook]

func setNowFunc(nowFunc func() time.Time) {
	now = nowFunc
}

// fastTimerSnapshot is an immutable date+time pair, swapped atomically by FastTimer.
// Storing both fields in a single value lets readers observe a consistent pair even
// when an update lands across the day boundary.
type fastTimerSnapshot struct {
	date string
	time string
}

// FastTimer is a 1Hz cached clock. Reading is one atomic pointer load instead of a
// full time.Now() + format pass, saving ~30% on every log call when enabled.
//
// The cached value lags real time by up to 1 second. Concurrent readers may observe
// neighboring snapshots within a few milliseconds of an update, so two log lines
// produced "at the same moment" can carry timestamps differing by 1 second. Within a
// single snapshot, date and time are always consistent (no torn reads across day
// boundaries).
type FastTimer struct {
	snapshot    atomic.Pointer[fastTimerSnapshot]
	stopChan    chan struct{}
	stoppedChan chan struct{}
	controlMu   sync.Mutex
	running     atomic.Bool
}

// load returns the current snapshot or nil if the timer is not running.
func (t *FastTimer) load() *fastTimerSnapshot {
	return t.snapshot.Load()
}

// update computes a new snapshot from tm and stores it atomically.
func (t *FastTimer) update(tm time.Time, buf *bytes.Buffer) {
	buf.Reset()
	year, mon, day := tm.Date()
	buf.Write(uint2Bytes4(year))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(int(mon)))
	buf.WriteByte('-')
	buf.Write(uint2Bytes2(day))
	date := buf.String()

	buf.Reset()
	hour, min, sec := tm.Clock()
	buf.Write(uint2Bytes2(hour))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(min))
	buf.WriteByte(':')
	buf.Write(uint2Bytes2(sec))
	timeStr := buf.String()

	if hook := fastTimerHook.Load(); hook != nil {
		(*hook)()
	}
	t.snapshot.Store(&fastTimerSnapshot{date: date, time: timeStr})
}

func (t *FastTimer) start() {
	t.controlMu.Lock()
	if t.running.Load() {
		t.controlMu.Unlock()
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, dateTimeBufSize))
	t.update(now(), buf)
	t.running.Store(true)
	t.stopChan = make(chan struct{})
	t.stoppedChan = make(chan struct{})
	stopChan := t.stopChan
	stoppedChan := t.stoppedChan
	t.controlMu.Unlock()

	go func() {
		defer close(stoppedChan)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case tm := <-ticker.C:
				t.update(tm, buf)
			case <-stopChan:
				return
			}
		}
	}()
}

func (t *FastTimer) stop() {
	t.controlMu.Lock()
	if !t.running.Load() {
		t.controlMu.Unlock()
		return
	}
	stopChan := t.stopChan
	stoppedChan := t.stoppedChan
	t.running.Store(false)
	close(stopChan)
	<-stoppedChan
	t.snapshot.Store(nil)
	t.stopChan = nil
	t.stoppedChan = nil
	t.controlMu.Unlock()
}

var fastTimer = FastTimer{}

// StartFastTimer starts the package-level FastTimer. Idempotent.
func StartFastTimer() {
	fastTimer.start()
}

// StopFastTimer stops the package-level FastTimer. Idempotent.
func StopFastTimer() {
	fastTimer.stop()
}

// stopTimer drains a time.Timer's channel after stopping it, so a subsequent Reset
// does not race with a stale tick.
func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
