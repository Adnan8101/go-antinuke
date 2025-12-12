package decision

import (
	"sync"
	"time"
)

type CooldownManager struct {
	mu        sync.RWMutex
	cooldowns map[uint64]map[uint8]int64
	duration  time.Duration
}

func NewCooldownManager(duration time.Duration) *CooldownManager {
	return &CooldownManager{
		cooldowns: make(map[uint64]map[uint8]int64),
		duration:  duration,
	}
}

func (cm *CooldownManager) CanExecute(guildID uint64, actionType uint8) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	guildCooldowns, exists := cm.cooldowns[guildID]
	if !exists {
		return true
	}

	lastExecution, exists := guildCooldowns[actionType]
	if !exists {
		return true
	}

	elapsed := time.Now().UnixNano() - lastExecution
	return elapsed >= int64(cm.duration)
}

func (cm *CooldownManager) RecordExecution(guildID uint64, actionType uint8) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.cooldowns[guildID]; !exists {
		cm.cooldowns[guildID] = make(map[uint8]int64)
	}

	cm.cooldowns[guildID][actionType] = time.Now().UnixNano()
}

func (cm *CooldownManager) Reset(guildID uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.cooldowns, guildID)
}

func (cm *CooldownManager) GetRemainingCooldown(guildID uint64, actionType uint8) time.Duration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	guildCooldowns, exists := cm.cooldowns[guildID]
	if !exists {
		return 0
	}

	lastExecution, exists := guildCooldowns[actionType]
	if !exists {
		return 0
	}

	elapsed := time.Now().UnixNano() - lastExecution
	remaining := int64(cm.duration) - elapsed

	if remaining < 0 {
		return 0
	}

	return time.Duration(remaining)
}
