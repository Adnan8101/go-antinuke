package bootstrap

import (
	"fmt"
	"os"
	"runtime/debug"

	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/correlator"
	"go-antinuke-2.0/internal/decision"
	"go-antinuke-2.0/internal/dispatcher"
	"go-antinuke-2.0/internal/forensics"
	"go-antinuke-2.0/internal/ha"
	"go-antinuke-2.0/internal/ingest"
	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/internal/metrics"
	"go-antinuke-2.0/internal/state"
	"go-antinuke-2.0/internal/watchdog"
	"go-antinuke-2.0/pkg/memory"
)

type Bootstrap struct {
	Config      *config.Config
	Components  *Components
	initialized bool
}

type Components struct {
	// Core pipeline components
	RingBuffer     *ingest.RingBuffer
	AlertQueue     *correlator.AlertQueue
	JobQueue       *decision.JobQueue
	GatewayReader  *ingest.GatewayReader
	Correlator     *correlator.Correlator
	DecisionEngine *decision.DecisionEngine
	HTTPPool       *dispatcher.HTTPPool
	RateLimiter    *dispatcher.RateLimitMonitor
	Workers        []*dispatcher.RESTWorker

	// Monitoring and observability
	Metrics     *metrics.MetricsRegistry
	Watchdog    *watchdog.Watchdog
	ForensicLog *forensics.ForensicLogger

	// High availability (optional)
	HACluster   *ha.Cluster
	HAHeartbeat *ha.HeartbeatManager
}

func New() *Bootstrap {
	return &Bootstrap{
		initialized: false,
	}
}

func (b *Bootstrap) Initialize() error {
	if err := b.loadConfig(); err != nil {
		return fmt.Errorf("config load failed: %w", err)
	}

	if err := b.initializeRuntime(); err != nil {
		return fmt.Errorf("runtime init failed: %w", err)
	}

	if err := b.initializeState(); err != nil {
		return fmt.Errorf("state init failed: %w", err)
	}

	if err := b.initializeLogging(); err != nil {
		return fmt.Errorf("logging init failed: %w", err)
	}

	if err := b.wireComponents(); err != nil {
		return fmt.Errorf("component wiring failed: %w", err)
	}

	b.initialized = true
	logging.Info("Bootstrap complete")
	return nil
}

func (b *Bootstrap) loadConfig() error {
	cfg, err := config.Load("config.json")
	if err != nil {
		logging.Warn("Config load failed, using defaults: %v", err)
		cfg = config.DefaultConfig()
	}
	b.Config = cfg
	return nil
}

func (b *Bootstrap) initializeRuntime() error {
	if b.Config.Runtime.DisableGC {
		debug.SetGCPercent(-1)
		logging.Info("GC disabled")
	}

	if b.Config.Runtime.MemoryLock {
		if err := memory.MlockAll(); err != nil {
			logging.Warn("Memory lock failed: %v", err)
		} else {
			logging.Info("Memory locked")
		}
	}

	return nil
}

func (b *Bootstrap) initializeState() error {
	if err := state.InitAll(); err != nil {
		return err
	}

	config.InitThresholds()
	config.InitGuildProfiles()
	forensics.InitRecoveryTracker()
	state.TouchAll()

	logging.Info("State initialized")
	return nil
}

func (b *Bootstrap) initializeLogging() error {
	// Ensure logs directory exists
	if err := ensureLogsDirectory(); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}
	return logging.InitGlobalLogger(logging.LevelInfo, "antinuke.log")
}

func ensureLogsDirectory() error {
	// Create logs directory if it doesn't exist
	_, err := os.Stat("logs")
	if os.IsNotExist(err) {
		return os.Mkdir("logs", 0755)
	}
	return err
}

func (b *Bootstrap) wireComponents() error {
	return Wire(b)
}

func (b *Bootstrap) Start() error {
	if !b.initialized {
		return fmt.Errorf("bootstrap not initialized")
	}

	return StartAll(b.Components)
}

func (b *Bootstrap) Shutdown() error {
	return Shutdown(b.Components)
}
