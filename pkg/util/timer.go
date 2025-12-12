package util

import (
	"time"
	_ "unsafe"
)

//go:noescape
//go:linkname nanotime runtime.nanotime
func nanotime() int64

type MonotonicTimer struct {
	start int64
}

func NewMonotonicTimer() *MonotonicTimer {
	return &MonotonicTimer{start: nanotime()}
}

func (t *MonotonicTimer) Elapsed() int64 {
	return nanotime() - t.start
}

func (t *MonotonicTimer) ElapsedNs() int64 {
	return t.Elapsed()
}

func (t *MonotonicTimer) ElapsedUs() int64 {
	return t.Elapsed() / 1000
}

func (t *MonotonicTimer) ElapsedMs() int64 {
	return t.Elapsed() / 1000000
}

func (t *MonotonicTimer) Reset() {
	t.start = nanotime()
}

func NowMono() int64 {
	return nanotime()
}

func NowMonoTime() time.Time {
	return time.Now()
}

func SinceNs(start int64) int64 {
	return nanotime() - start
}

func SinceUs(start int64) int64 {
	return (nanotime() - start) / 1000
}

func SinceMs(start int64) int64 {
	return (nanotime() - start) / 1000000
}
