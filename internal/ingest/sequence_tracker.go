package ingest

import (
	"sync/atomic"
)

type SequenceTracker struct {
	sequence  uint64
	sessionID string
}

func NewSequenceTracker() *SequenceTracker {
	return &SequenceTracker{}
}

func (st *SequenceTracker) Update(seq uint64) {
	atomic.StoreUint64(&st.sequence, seq)
}

func (st *SequenceTracker) Get() uint64 {
	return atomic.LoadUint64(&st.sequence)
}

func (st *SequenceTracker) SetSessionID(sessionID string) {
	st.sessionID = sessionID
}

func (st *SequenceTracker) GetSessionID() string {
	return st.sessionID
}

func (st *SequenceTracker) Reset() {
	atomic.StoreUint64(&st.sequence, 0)
	st.sessionID = ""
}
