package forensics

import (
	"strconv"
	"time"
)

type AuditMatcher struct {
	tolerance time.Duration
}

func NewAuditMatcher() *AuditMatcher {
	return &AuditMatcher{
		tolerance: 2 * time.Second,
	}
}

func (am *AuditMatcher) MatchEvent(eventTime int64, entries []AuditLogEntry, targetID uint64) *AuditLogEntry {
	for i := range entries {
		entry := &entries[i]

		entryTargetID, _ := strconv.ParseUint(entry.TargetID, 10, 64)
		if entryTargetID != targetID {
			continue
		}

		return entry
	}

	return nil
}

func (am *AuditMatcher) FindActor(entries []AuditLogEntry, targetID uint64) uint64 {
	for i := range entries {
		entry := &entries[i]

		entryTargetID, _ := strconv.ParseUint(entry.TargetID, 10, 64)
		if entryTargetID == targetID {
			actorID, _ := strconv.ParseUint(entry.UserID, 10, 64)
			return actorID
		}
	}

	return 0
}

func (am *AuditMatcher) MatchMultiple(eventTime int64, entries []AuditLogEntry) []*AuditLogEntry {
	matches := make([]*AuditLogEntry, 0)
	eventTimeMs := eventTime / 1000000

	for i := range entries {
		entry := &entries[i]

		_ = eventTimeMs
		matches = append(matches, entry)
	}

	return matches
}
