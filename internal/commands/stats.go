package commands

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// handleStats shows comprehensive VM and system statistics
func handleStats(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Defer response to allow time for gathering stats
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return err
	}

	// Gather all system statistics
	statsData, err := gatherSystemStats(s)
	if err != nil {
		return err
	}

	// Create comprehensive embed
	embeds := createStatsEmbeds(statsData)

	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &embeds,
	})

	return err
}

// SystemStats holds all system statistics
type SystemStats struct {
	// Host Information
	Hostname     string
	OS           string
	Platform     string
	Architecture string
	Uptime       time.Duration
	BootTime     time.Time

	// CPU Information
	CPUModel       string
	CPUCores       int
	CPUThreads     int
	CPUUsage       float64
	CPUFrequency   float64
	CPUGovernor    string
	CPUIsolation   string

	// Memory Information
	TotalMemory     uint64
	UsedMemory      uint64
	FreeMemory      uint64
	MemoryPercent   float64
	SwapTotal       uint64
	SwapUsed        uint64

	// Disk Information
	DiskTotal       uint64
	DiskUsed        uint64
	DiskFree        uint64
	DiskPercent     float64

	// Network Information
	NetworkSent     uint64
	NetworkRecv     uint64
	NetworkConnections int

	// Go Runtime
	GoVersion      string
	GoRoutines     int
	MemAlloc       uint64
	TotalAlloc     uint64
	Sys            uint64
	NumGC          uint32

	// Bot Stats
	BotUptime      time.Duration
	Guilds         int
	Latency        time.Duration
}

var botStartTime = time.Now()

// gatherSystemStats collects all system statistics
func gatherSystemStats(s *discordgo.Session) (*SystemStats, error) {
	stats := &SystemStats{}

	// Host Information
	hostInfo, err := host.Info()
	if err == nil {
		stats.Hostname = hostInfo.Hostname
		stats.OS = hostInfo.OS
		stats.Platform = hostInfo.Platform
		stats.Architecture = hostInfo.KernelArch
		stats.Uptime = time.Duration(hostInfo.Uptime) * time.Second
		stats.BootTime = time.Unix(int64(hostInfo.BootTime), 0)
	}

	// CPU Information
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		stats.CPUModel = cpuInfo[0].ModelName
		stats.CPUCores = int(cpuInfo[0].Cores)
	}
	stats.CPUThreads = runtime.NumCPU()

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err == nil && len(cpuPercent) > 0 {
		stats.CPUUsage = cpuPercent[0]
	}

	cpuFreq, err := cpu.Info()
	if err == nil && len(cpuFreq) > 0 {
		stats.CPUFrequency = cpuFreq[0].Mhz
	}

	// Try to read CPU governor
	if data, err := os.ReadFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"); err == nil {
		stats.CPUGovernor = string(data)
	} else {
		stats.CPUGovernor = "unknown"
	}

	// Check CPU isolation
	if data, err := os.ReadFile("/proc/cmdline"); err == nil {
		cmdline := string(data)
		if containsStr(cmdline, "isolcpus") {
			stats.CPUIsolation = "enabled"
		} else {
			stats.CPUIsolation = "disabled"
		}
	} else {
		stats.CPUIsolation = "unknown"
	}

	// Memory Information
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		stats.TotalMemory = memInfo.Total
		stats.UsedMemory = memInfo.Used
		stats.FreeMemory = memInfo.Free
		stats.MemoryPercent = memInfo.UsedPercent
	}

	swapInfo, err := mem.SwapMemory()
	if err == nil {
		stats.SwapTotal = swapInfo.Total
		stats.SwapUsed = swapInfo.Used
	}

	// Disk Information
	diskInfo, err := disk.Usage("/")
	if err == nil {
		stats.DiskTotal = diskInfo.Total
		stats.DiskUsed = diskInfo.Used
		stats.DiskFree = diskInfo.Free
		stats.DiskPercent = diskInfo.UsedPercent
	}

	// Network Information
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		stats.NetworkSent = netIO[0].BytesSent
		stats.NetworkRecv = netIO[0].BytesRecv
	}

	netConns, err := net.Connections("all")
	if err == nil {
		stats.NetworkConnections = len(netConns)
	}

	// Go Runtime Statistics
	stats.GoVersion = runtime.Version()
	stats.GoRoutines = runtime.NumGoroutine()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	stats.MemAlloc = m.Alloc
	stats.TotalAlloc = m.TotalAlloc
	stats.Sys = m.Sys
	stats.NumGC = m.NumGC

	// Bot Statistics
	stats.BotUptime = time.Since(botStartTime)
	stats.Guilds = len(s.State.Guilds)
	stats.Latency = s.HeartbeatLatency()

	return stats, nil
}

// createStatsEmbeds creates formatted embeds for stats display
func createStatsEmbeds(stats *SystemStats) []*discordgo.MessageEmbed {
	embeds := []*discordgo.MessageEmbed{}

	// Main Stats Embed
	mainEmbed := &discordgo.MessageEmbed{
		Title: "📊 VM & System Statistics",
		Color: 0x00BFFF,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🖥️ Host Information",
				Value: fmt.Sprintf("**Hostname:** `%s`\n**OS:** `%s`\n**Platform:** `%s`\n**Architecture:** `%s`\n**Uptime:** `%s`",
					stats.Hostname,
					stats.OS,
					stats.Platform,
					stats.Architecture,
					formatDuration(stats.Uptime)),
				Inline: false,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// CPU Embed
	cpuEmbed := &discordgo.MessageEmbed{
		Title: "⚡ CPU Performance",
		Color: 0xFF8C00,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🔧 CPU Details",
				Value: fmt.Sprintf("**Model:** `%s`\n**Cores:** `%d` physical\n**Threads:** `%d` logical\n**Frequency:** `%.2f MHz`",
					truncateString(stats.CPUModel, 40),
					stats.CPUCores,
					stats.CPUThreads,
					stats.CPUFrequency),
				Inline: true,
			},
			{
				Name:   "📈 CPU Usage",
				Value: fmt.Sprintf("**Current:** `%.2f%%`\n%s",
					stats.CPUUsage,
					createProgressBar(stats.CPUUsage, 100)),
				Inline: true,
			},
			{
				Name:   "⚙️ Optimization",
				Value: fmt.Sprintf("**Governor:** `%s`\n**CPU Isolation:** `%s`",
					trimString(stats.CPUGovernor),
					stats.CPUIsolation),
				Inline: false,
			},
		},
	}

	// Memory Embed
	memEmbed := &discordgo.MessageEmbed{
		Title: "💾 Memory Statistics",
		Color: 0x9370DB,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "🗂️ RAM Usage",
				Value: fmt.Sprintf("**Total:** `%s`\n**Used:** `%s`\n**Free:** `%s`\n**Usage:** `%.2f%%`\n%s",
					formatBytes(stats.TotalMemory),
					formatBytes(stats.UsedMemory),
					formatBytes(stats.FreeMemory),
					stats.MemoryPercent,
					createProgressBar(stats.MemoryPercent, 100)),
				Inline: true,
			},
			{
				Name: "💿 Swap Memory",
				Value: fmt.Sprintf("**Total:** `%s`\n**Used:** `%s`",
					formatBytes(stats.SwapTotal),
					formatBytes(stats.SwapUsed)),
				Inline: true,
			},
		},
	}

	// Disk & Network Embed
	diskNetEmbed := &discordgo.MessageEmbed{
		Title: "💽 Storage & Network",
		Color: 0x32CD32,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "📀 Disk Usage",
				Value: fmt.Sprintf("**Total:** `%s`\n**Used:** `%s`\n**Free:** `%s`\n**Usage:** `%.2f%%`\n%s",
					formatBytes(stats.DiskTotal),
					formatBytes(stats.DiskUsed),
					formatBytes(stats.DiskFree),
					stats.DiskPercent,
					createProgressBar(stats.DiskPercent, 100)),
				Inline: false,
			},
			{
				Name: "🌐 Network I/O",
				Value: fmt.Sprintf("**Sent:** `%s`\n**Received:** `%s`\n**Connections:** `%d`",
					formatBytes(stats.NetworkSent),
					formatBytes(stats.NetworkRecv),
					stats.NetworkConnections),
				Inline: false,
			},
		},
	}

	// Bot & Runtime Embed
	botEmbed := &discordgo.MessageEmbed{
		Title: "🤖 Bot & Runtime Statistics",
		Color: 0xFF1493,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "🚀 Bot Status",
				Value: fmt.Sprintf("**Uptime:** `%s`\n**Guilds:** `%d`\n**Latency:** `%dms` (%dµs)",
					formatDuration(stats.BotUptime),
					stats.Guilds,
					stats.Latency.Milliseconds(),
					stats.Latency.Microseconds()),
				Inline: true,
			},
			{
				Name: "🔷 Go Runtime",
				Value: fmt.Sprintf("**Version:** `%s`\n**Goroutines:** `%d`\n**GC Cycles:** `%d`",
					stats.GoVersion,
					stats.GoRoutines,
					stats.NumGC),
				Inline: true,
			},
			{
				Name: "🧠 Process Memory",
				Value: fmt.Sprintf("**Allocated:** `%s`\n**Total Alloc:** `%s`\n**System:** `%s`",
					formatBytes(stats.MemAlloc),
					formatBytes(stats.TotalAlloc),
					formatBytes(stats.Sys)),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Ultra-Low-Latency Antinuke Engine | Target: 100-300ns detection",
		},
	}

	embeds = append(embeds, mainEmbed, cpuEmbed, memEmbed, diskNetEmbed, botEmbed)
	return embeds
}

// Helper functions

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

func createProgressBar(value, max float64) string {
	percent := (value / max) * 100
	filled := int(percent / 10)
	empty := 10 - filled
	
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	return "`" + bar + "`"
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func trimString(s string) string {
	return fmt.Sprintf("%s", s[:len(s)-1])
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
