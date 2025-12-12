package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"syscall"

	"go-antinuke-2.0/internal/bot"
	"go-antinuke-2.0/internal/commands"
	"go-antinuke-2.0/internal/config"
	"go-antinuke-2.0/internal/correlator"
	"go-antinuke-2.0/internal/database"
	"go-antinuke-2.0/internal/decision"
	"go-antinuke-2.0/internal/dispatcher"
	"go-antinuke-2.0/internal/ingest"
	"go-antinuke-2.0/internal/logging"
	"go-antinuke-2.0/internal/notifier"
	"go-antinuke-2.0/internal/state"
	"go-antinuke-2.0/pkg/memory"
)

func main() {
	fmt.Println("Starting Ultra-Low-Latency 1µs Anti-Nuke Engine")

	cfg := loadConfig()

	if err := initializeRuntime(cfg); err != nil {
		panic(err)
	}

	if err := initializeDatabase(); err != nil {
		panic(err)
	}

	if err := initializeState(); err != nil {
		panic(err)
	}

	if err := initializeLogging(); err != nil {
		panic(err)
	}

	components := startComponents(cfg)

	// Initialize bot AFTER creating ring buffer
	if err := initializeBot(cfg.Bot.Token, components.ringBuffer); err != nil {
		panic(err)
	}

	logging.Info("All components started successfully")
	logging.Info("Correlator running on CPU %d", cfg.Runtime.CorrelatorCPU)
	logging.Info("Detection target: <1µs, Execution target: <200ms")
	logging.Info("Discord bot connected and commands registered")

	waitForShutdown()

	stopComponents(components)

	database.Close()
	bot.GetSession().Close()

	logging.Info("Shutdown complete")
}

func loadConfig() *config.Config {
	cfg, err := config.Load("config.json")
	if err != nil {
		fmt.Printf("Config load failed, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}

	if cfg.Bot.Token == "" {
		cfg.Bot.Token = os.Getenv("DISCORD_TOKEN")
	}

	return cfg
}

func initializeRuntime(cfg *config.Config) error {
	if cfg.Runtime.DisableGC {
		debug.SetGCPercent(-1)
		fmt.Println("GC disabled for hot path performance")
	}

	runtime.GOMAXPROCS(8)

	if cfg.Runtime.MemoryLock {
		if err := memory.MlockAll(); err != nil {
			fmt.Printf("Warning: Memory lock not available on this platform (%v)\n", err)
			fmt.Println("Note: Memory locking is typically not required on macOS")
		} else {
			fmt.Println("Memory locked to prevent page faults ✓")
		}
	}

	return nil
}

func initializeState() error {
	fmt.Println("Initializing preallocated state...")

	if err := state.InitAll(); err != nil {
		return err
	}

	config.InitThresholds()
	config.InitGuildProfiles()

	state.TouchAll()

	fmt.Println("State initialization complete")
	return nil
}

func initializeLogging() error {
	return logging.InitGlobalLogger(logging.LevelInfo, "antinuke.log")
}

func initializeDatabase() error {
	fmt.Println("Initializing SQLite database...")

	if err := database.Initialize("antinuke.db"); err != nil {
		return err
	}

	if database.IsConnected() {
		fmt.Println("Database initialized and connection verified ✓")
	} else {
		fmt.Println("Database initialized but connection verification failed")
	}

	return nil
}

func initializeBot(token string, ringBuffer *ingest.RingBuffer) error {
	fmt.Println("Initializing Discord bot...")

	if err := bot.Initialize(token); err != nil {
		return err
	}

	session := bot.GetSession()

	// Setup event handlers BEFORE connecting
	session.SetupEventHandlers(ringBuffer)

	if err := session.Connect(); err != nil {
		return err
	}

	// Sync all guild configurations from database to in-memory state
	fmt.Println("Syncing guild configurations from database...")
	db := database.GetDB()
	if err := session.SyncGuildsFromDatabase(db); err != nil {
		fmt.Printf("Warning: Guild sync failed: %v\n", err)
	} else {
		fmt.Println("Guild configurations synced successfully ✓")
	}

	// Set notifier session for Discord logging
	notifier.SetSession(session.GetDiscord())

	// Initialize and register commands
	if err := commands.Initialize(session); err != nil {
		return err
	}

	fmt.Println("Discord bot initialized successfully")
	return nil
}

type Components struct {
	ringBuffer     *ingest.RingBuffer
	alertQueue     *correlator.AlertQueue
	jobQueue       *decision.JobQueue
	correlatorInst *correlator.Correlator
	decisionEngine *decision.DecisionEngine
	httpPool       *dispatcher.HTTPPool
	rateLimiter    *dispatcher.RateLimitMonitor
	workers        []*dispatcher.RESTWorker
}

func startComponents(cfg *config.Config) *Components {
	ringBuffer := ingest.NewRingBuffer(65536)
	alertQueue := correlator.NewAlertQueue(32768)
	jobQueue := decision.NewJobQueue(16384)

	// Gateway intents:
	// 1<<0 = GUILDS
	// 1<<1 = GUILD_MEMBERS
	// 1<<7 = GUILD_AUDIT_LOG (CRITICAL for actor detection)
	// 1<<9 = GUILD_MESSAGES
	intents := 1<<0 | 1<<1 | 1<<7 | 1<<9

	// NOTE: We use discordgo's built-in websocket connection instead of a custom GatewayReader
	// The bot session handles the websocket connection and event processing
	_ = intents // Intents are set on the bot session, not the custom gateway reader

	// gatewayReader := ingest.NewGatewayReader(cfg.Bot.Token, intents, ringBuffer, cfg.Runtime.IngestCPU)
	// if err := gatewayReader.Connect(); err != nil {
	// 	panic(fmt.Sprintf("Gateway connection failed: %v", err))
	// }
	// go gatewayReader.ReadLoop()

	correlatorInst := correlator.NewCorrelator(ringBuffer, alertQueue, cfg.Runtime.CorrelatorCPU)
	go correlatorInst.Start()

	decisionEngine := decision.NewDecisionEngine(alertQueue, jobQueue, cfg.Runtime.DecisionCPU)
	go decisionEngine.Start()

	httpPool := dispatcher.NewHTTPPool(cfg.Network.HTTPPoolSize)
	httpPool.Warmup()

	rateLimiter := dispatcher.NewRateLimitMonitor()

	workers := make([]*dispatcher.RESTWorker, cfg.Network.WorkerCount)
	for i := 0; i < cfg.Network.WorkerCount; i++ {
		worker := dispatcher.NewRESTWorker(jobQueue, httpPool, rateLimiter, i, cfg.Runtime.DispatcherCPU)
		workers[i] = worker
		go worker.Start()
	}

	return &Components{
		ringBuffer:     ringBuffer,
		alertQueue:     alertQueue,
		jobQueue:       jobQueue,
		correlatorInst: correlatorInst,
		decisionEngine: decisionEngine,
		httpPool:       httpPool,
		rateLimiter:    rateLimiter,
		workers:        workers,
	}
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("\nShutdown signal received")
}

func stopComponents(components *Components) {
	components.correlatorInst.Stop()
	components.decisionEngine.Stop()

	for _, worker := range components.workers {
		worker.Stop()
	}

	// Note: Gateway connection is handled by discordgo session
}
