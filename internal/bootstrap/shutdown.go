package bootstrap

import (
	"time"

	"go-antinuke-2.0/internal/logging"
)

func Shutdown(c *Components) error {
	logging.Info("Starting graceful shutdown...")

	// Stop HA components first
	if c.HAHeartbeat != nil {
		logging.Info("Stopping HA heartbeat...")
		c.HAHeartbeat.Stop()
	}

	// Stop watchdog
	if c.Watchdog != nil {
		logging.Info("Stopping watchdog...")
		c.Watchdog.Stop()
	}

	logging.Info("Stopping gateway reader...")
	c.GatewayReader.Close()

	logging.Info("Stopping correlator...")
	c.Correlator.Stop()
	time.Sleep(10 * time.Millisecond)

	logging.Info("Stopping decision engine...")
	c.DecisionEngine.Stop()
	time.Sleep(10 * time.Millisecond)

	logging.Info("Stopping REST workers...")
	for i, worker := range c.Workers {
		worker.Stop()
		logging.Info("Worker %d stopped", i)
	}

	// Close forensic logger
	if c.ForensicLog != nil {
		logging.Info("Closing forensic logger...")
		c.ForensicLog.Close()
	}

	logging.Info("Draining queues...")
	time.Sleep(50 * time.Millisecond)

	logging.Info("Graceful shutdown complete")
	return nil
}

func EmergencyShutdown(c *Components) {
	logging.Critical("Emergency shutdown initiated")

	if c.HAHeartbeat != nil {
		c.HAHeartbeat.Stop()
	}
	if c.Watchdog != nil {
		c.Watchdog.Stop()
	}
	if c.Correlator != nil {
		c.Correlator.Stop()
	}
	if c.DecisionEngine != nil {
		c.DecisionEngine.Stop()
	}
	if c.GatewayReader != nil {
		c.GatewayReader.Close()
	}
	if c.ForensicLog != nil {
		c.ForensicLog.Close()
	}

	logging.Critical("Emergency shutdown complete")
}
