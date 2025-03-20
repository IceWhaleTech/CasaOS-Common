package utils

import (
	"sync/atomic"
	"time"
)

func CompareSlices(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

func CompareStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func Ptr[T any](v T) *T {
	return &v
}

type Throttle struct {
	lastCall time.Time
	limit    time.Duration
	timer    atomic.Pointer[time.Timer]
}

func NewThrottle(limit time.Duration) *Throttle {
	return &Throttle{limit: limit}
}

func (t *Throttle) Do(f func()) {
	now := time.Now()
	if now.Sub(t.lastCall) >= t.limit {
		t.lastCall = now
		f()
		if old := t.timer.Swap(nil); old != nil {
			old.Stop()
		}
	} else {
		lastFunc := f
		if old := t.timer.Swap(time.AfterFunc(t.limit-now.Sub(t.lastCall), func() {
			t.lastCall = time.Now()
			lastFunc()
		})); old != nil {
			old.Stop()
		}
	}
}
