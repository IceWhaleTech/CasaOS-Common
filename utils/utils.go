package utils

import (
	"sync"
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

type DelayedState struct {
	StartTime time.Duration
	Timer     *time.Timer
}

type DelayedExecutor struct {
	States      sync.Map
	Delay       time.Duration
	MaxInterval time.Duration
}

func NewDelayedExecutor(delay, maxInterval time.Duration) *DelayedExecutor {
	return &DelayedExecutor{
		Delay:       delay,
		MaxInterval: maxInterval,
	}
}

func (d *DelayedExecutor) Do(key string, execFunc func()) {
	now := time.Duration(time.Now().Unix())
	s, loaded := d.States.LoadOrStore(key, &DelayedState{})

	wait := d.Delay
	state := s.(*DelayedState)

	if loaded && state.StartTime > 0 {
		elapsed := now - state.StartTime
		remaining := d.MaxInterval - elapsed

		if remaining <= 0 {
			d.trigger(key, execFunc)
			return
		}

		if remaining < wait {
			wait = remaining
		}

		if state.Timer != nil {
			state.Timer.Reset(wait * time.Second)
		}
	} else {
		state.StartTime = now
		state.Timer = time.AfterFunc(wait*time.Second, func() {
			d.trigger(key, execFunc)
		})
	}
}

func (d *DelayedExecutor) trigger(key string, execFunc func()) {
	_, ok := d.States.LoadAndDelete(key)
	if !ok {
		return
	}

	execFunc()
}
