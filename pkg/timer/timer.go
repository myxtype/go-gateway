package timer

import (
	"context"
	"time"
)

type Timer struct {
	d      time.Duration
	f      func()
	ctx    context.Context
	cancel context.CancelFunc
}

func NewTimer(d time.Duration, f func()) *Timer {
	return &Timer{
		d: d,
		f: f,
	}
}

func (t *Timer) Start() {
	t.ctx, t.cancel = context.WithCancel(context.Background())
	tik := time.NewTicker(t.d)

	for {
		select {
		case <-tik.C:
			t.f()
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *Timer) Stop() {
	if t.cancel != nil {
		t.cancel()
	}
}
