package correlator

import (
	"runtime"

	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/pkg/memory"
)

func InitializeCorrelator(c *Correlator) error {
	logging.Info("Initializing correlator...")

	if err := alignMemory(); err != nil {
		return err
	}

	if err := preTouchCounters(); err != nil {
		return err
	}

	if err := bindCPUCore(c.cpuCore); err != nil {
		return err
	}

	logging.Info("Correlator initialized on CPU %d", c.cpuCore)
	return nil
}

func alignMemory() error {
	InitCounters()
	InitThresholds()
	InitVelocity()

	logging.Info("Memory structures aligned")
	return nil
}

func preTouchCounters() error {
	counters := GetCounters()
	for i := range counters {
		counters[i].BanCount = 0
	}

	thresholds := GetThresholds()
	for i := range thresholds {
		thresholds[i].BanThreshold = DefaultThresholds.BanThreshold
	}

	velocity := GetVelocity()
	for i := range velocity {
		velocity[i].velocity = 0
	}

	logging.Info("Counters pre-touched into cache")
	return nil
}

func bindCPUCore(cpuCore int) error {
	runtime.LockOSThread()

	if err := memory.MlockAll(); err != nil {
		logging.Warn("CPU core binding: mlockall failed: %v", err)
	}

	logging.Info("Thread locked to CPU %d", cpuCore)
	return nil
}
