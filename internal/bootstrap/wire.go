package bootstrap

import (
	"fmt"
	"os"
	"time"

	"go-antinuke-2.0/internal/correlator"
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/internal/decision"
	"go-antinuke-2.0/internal/dispatcher"
	"go-antinuke-2.0/internal/forensics"
	"go-antinuke-2.0/internal/ha"
	"go-antinuke-2.0/internal/ingest"
	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/internal/metrics"
	"go-antinuke-2.0/internal/watchdog"
)

func Wire(b *Bootstrap) error {
	logging.Info("Wiring components...")

	if !database.IsConnected() {
		return fmt.Errorf("database connection not available")
	}
	logging.Info("Database connection verified")

	metricsRegistry := metrics.NewMetricsRegistry()
	metrics.InitGlobalRegistry()

	// Initialize forensic logger
	forensicLogger, err := forensics.NewForensicLogger("logs/forensic.log")
	if err != nil {
		logging.Warn("Failed to initialize forensic logger: %v", err)
	}

	// Initialize watchdog
	watchdogInst := watchdog.NewWatchdog(5 * time.Second)

	// Core pipeline components
	ringBuffer := ingest.NewRingBuffer(65536)
	alertQueue := correlator.NewAlertQueue(32768)
	jobQueue := decision.NewJobQueue(16384)

	intents := 1<<0 | 1<<1 | 1<<9
	gatewayReader := ingest.NewGatewayReader(
		b.Config.Bot.Token,
		intents,
		ringBuffer,
		b.Config.Runtime.IngestCPU,
	)

	correlatorInst := correlator.NewCorrelator(
		ringBuffer,
		alertQueue,
		b.Config.Runtime.CorrelatorCPU,
	)

	decisionEngine := decision.NewDecisionEngine(
		alertQueue,
		jobQueue,
		b.Config.Runtime.DecisionCPU,
	)

	httpPool := dispatcher.NewHTTPPool(b.Config.Network.HTTPPoolSize)
	rateLimiter := dispatcher.NewRateLimitMonitor()

	workers := make([]*dispatcher.RESTWorker, b.Config.Network.WorkerCount)
	for i := 0; i < b.Config.Network.WorkerCount; i++ {
		workers[i] = dispatcher.NewRESTWorker(
			jobQueue,
			httpPool,
			rateLimiter,
			i,
			b.Config.Runtime.DispatcherCPU,
		)
	}

	// High availability components (optional, can be nil in single-node mode)
	var haCluster *ha.Cluster
	var haHeartbeat *ha.HeartbeatManager

	if enableHA := os.Getenv("ENABLE_HA"); enableHA == "true" {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			nodeID = "node-1"
		}
		haCluster = ha.NewCluster(nodeID)
		haHeartbeat = ha.NewHeartbeatManager(haCluster, 2*time.Second)
		logging.Info("HA mode enabled (Node ID: %s)", nodeID)
	}

	// Register components with watchdog
	watchdogInst.RegisterComponent("correlator", 10*time.Second)
	watchdogInst.RegisterComponent("decision_engine", 10*time.Second)
	watchdogInst.RegisterComponent("dispatcher", 10*time.Second)

	b.Components = &Components{
		RingBuffer:     ringBuffer,
		AlertQueue:     alertQueue,
		JobQueue:       jobQueue,
		GatewayReader:  gatewayReader,
		Correlator:     correlatorInst,
		DecisionEngine: decisionEngine,
		HTTPPool:       httpPool,
		RateLimiter:    rateLimiter,
		Workers:        workers,
		Metrics:        metricsRegistry,
		Watchdog:       watchdogInst,
		ForensicLog:    forensicLogger,
		HACluster:      haCluster,
		HAHeartbeat:    haHeartbeat,
	}

	logging.Info("Component wiring complete")
	return nil
}

func StartAll(c *Components) error {
	logging.Info("Starting components...")

	// Start watchdog first to monitor other components
	if c.Watchdog != nil {
		c.Watchdog.Start()
		logging.Info("Watchdog started")
	}

	// Start HA components if enabled
	if c.HAHeartbeat != nil {
		c.HAHeartbeat.Start()
		logging.Info("HA heartbeat started")
	}

	if err := c.GatewayReader.Connect(); err != nil {
		return fmt.Errorf("gateway connection failed: %w", err)
	}

	go c.GatewayReader.ReadLoop()
	logging.Info("Gateway reader started")

	go c.Correlator.Start()
	logging.Info("Correlator started on isolated CPU")

	go c.DecisionEngine.Start()
	logging.Info("Decision engine started")

	c.HTTPPool.Warmup()
	logging.Info("HTTP pool warmed")

	for i, worker := range c.Workers {
		go worker.Start()
		logging.Info("REST worker %d started", i)
	}

	logging.Info("All components started")
	return nil
}
