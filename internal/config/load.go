package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Bot       BotConfig       `json:"bot"`
	Detection DetectionConfig `json:"detection"`
	Runtime   RuntimeConfig   `json:"runtime"`
	Network   NetworkConfig   `json:"network"`
	Forensics ForensicsConfig `json:"forensics"`
	HA        HAConfig        `json:"ha"`
}

type BotConfig struct {
	Token    string `json:"token"`
	ClientID string `json:"client_id"`
}

type DetectionConfig struct {
	Enabled       bool   `json:"enabled"`
	DefaultMode   string `json:"default_mode"`
	ThresholdFile string `json:"threshold_file"`
	GuildProfiles string `json:"guild_profiles"`
}

type RuntimeConfig struct {
	DisableGC     bool `json:"disable_gc"`
	CPUIsolation  bool `json:"cpu_isolation"`
	CorrelatorCPU int  `json:"correlator_cpu"`
	IngestCPU     int  `json:"ingest_cpu"`
	DecisionCPU   int  `json:"decision_cpu"`
	DispatcherCPU int  `json:"dispatcher_cpu"`
	MemoryLock    bool `json:"memory_lock"`
	PriorityRT    bool `json:"priority_rt"`
}

type NetworkConfig struct {
	GatewayQueues int    `json:"gateway_queues"`
	HTTPPoolSize  int    `json:"http_pool_size"`
	WorkerCount   int    `json:"worker_count"`
	APIBaseURL    string `json:"api_base_url"`
}

type ForensicsConfig struct {
	Enabled        bool   `json:"enabled"`
	RetentionDays  int    `json:"retention_days"`
	AuditInterval  int    `json:"audit_interval_ms"`
	SnapshotPath   string `json:"snapshot_path"`
	LogCompression bool   `json:"log_compression"`
}

type HAConfig struct {
	Enabled           bool     `json:"enabled"`
	ClusterNodes      []string `json:"cluster_nodes"`
	ReplicationPort   int      `json:"replication_port"`
	ElectionTimeout   int      `json:"election_timeout_ms"`
	HeartbeatInterval int      `json:"heartbeat_interval_ms"`
}

var GlobalConfig *Config

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Override with environment variables if present
	if token := os.Getenv("DISCORD_TOKEN"); token != "" {
		cfg.Bot.Token = token
	}
	if clientID := os.Getenv("CLIENT_ID"); clientID != "" {
		cfg.Bot.ClientID = clientID
	}
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		// Store database path if needed
	}

	GlobalConfig = &cfg
	return &cfg, nil
}

func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		return DefaultConfig()
	}
	return cfg
}

func DefaultConfig() *Config {
	return &Config{
		Bot: BotConfig{},
		Detection: DetectionConfig{
			Enabled:     true,
			DefaultMode: "normal",
		},
		Runtime: RuntimeConfig{
			DisableGC:     true,
			CPUIsolation:  true,
			CorrelatorCPU: 1,
			IngestCPU:     2,
			DecisionCPU:   5,
			DispatcherCPU: 6,
			MemoryLock:    true,
			PriorityRT:    true,
		},
		Network: NetworkConfig{
			GatewayQueues: 4,
			HTTPPoolSize:  8,
			WorkerCount:   8,
			APIBaseURL:    "https://discord.com/api/v10",
		},
		Forensics: ForensicsConfig{
			Enabled:        true,
			RetentionDays:  90,
			AuditInterval:  5000,
			LogCompression: true,
		},
		HA: HAConfig{
			Enabled:           false,
			ElectionTimeout:   3000,
			HeartbeatInterval: 1000,
		},
	}
}

func Get() *Config {
	if GlobalConfig == nil {
		return DefaultConfig()
	}
	return GlobalConfig
}
