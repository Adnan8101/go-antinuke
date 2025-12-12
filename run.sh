#!/bin/bash

################################################################################
# Ultra-Low-Latency Antinuke Bot Deployment Script
# Integrates: PM2 + System Optimization + CPU Pinning + Benchmarking
# Target: 100-300ns detection speed with <200μs execution
################################################################################

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
DEDICATED_CPU=1
LOG_DIR="/var/log/antinuke"
BENCHMARK_THRESHOLD_NS=500

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "\n${CYAN}[STEP $1/7]${NC} $2\n"; }

# Check if running as root for system optimizations
IS_ROOT=false
if [[ $EUID -eq 0 ]]; then
    IS_ROOT=true
fi

################################################################################
# TASK 1: VM SYSTEM OPTIMIZATION
################################################################################

task_1_vm_optimization() {
    log_step 1 "VM System Optimization"
    
    if [[ "$IS_ROOT" == false ]]; then
        log_warning "Not running as root - skipping system optimizations"
        log_warning "Run with sudo for full optimization"
        return
    fi
    
    log_info "Applying kernel tuning parameters..."
    
    # Create sysctl config
    cat > /etc/sysctl.d/99-antinuke.conf <<'EOF'
# Network Stack Optimization
net.core.netdev_max_backlog = 50000
net.core.rmem_max = 134217728
net.core.wmem_max = 134217728
net.ipv4.tcp_rmem = 4096 87380 134217728
net.ipv4.tcp_wmem = 4096 65536 134217728
net.ipv4.tcp_congestion_control = bbr
net.ipv4.tcp_fastopen = 3
net.ipv4.tcp_slow_start_after_idle = 0

# Scheduler Tuning
kernel.sched_min_granularity_ns = 1000000
kernel.sched_wakeup_granularity_ns = 1500000
kernel.sched_migration_cost_ns = 500000

# Memory Management
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5

# File Limits
fs.file-max = 2097152
EOF
    
    sysctl -p /etc/sysctl.d/99-antinuke.conf >/dev/null 2>&1 || true
    
    # CPU Governor
    log_info "Setting CPU governor to performance mode..."
    for cpu in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor; do
        [[ -w $cpu ]] && echo "performance" > "$cpu" 2>/dev/null || true
    done
    
    # IRQ Affinity (move all to CPU0)
    log_info "Configuring IRQ affinity..."
    for irq in /proc/irq/*/smp_affinity; do
        [[ -w $irq ]] && echo "1" > "$irq" 2>/dev/null || true
    done
    
    # Disable THP
    if [[ -f /sys/kernel/mm/transparent_hugepage/enabled ]]; then
        echo never > /sys/kernel/mm/transparent_hugepage/enabled 2>/dev/null || true
        echo never > /sys/kernel/mm/transparent_hugepage/defrag 2>/dev/null || true
    fi
    
    log_success "VM optimization complete"
}

################################################################################
# TASK 2: PRE-FLIGHT BENCHMARK TESTS
################################################################################

task_2_benchmark_tests() {
    log_step 2 "Pre-Flight Benchmark Tests"
    
    log_info "Running CPU latency test..."
    
    # Simple CPU timing test
    local start=$(date +%s%N)
    for i in {1..10000}; do
        :
    done
    local end=$(date +%s%N)
    local cpu_latency=$(( (end - start) / 10000 ))
    
    log_info "CPU latency: ${cpu_latency}ns per operation"
    
    if [[ $cpu_latency -gt $BENCHMARK_THRESHOLD_NS ]]; then
        log_error "CPU latency HIGH: ${cpu_latency}ns (threshold: ${BENCHMARK_THRESHOLD_NS}ns)"
        echo ""
        log_warning "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        log_warning "  VM PERFORMANCE BELOW THRESHOLD"
        log_warning "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        log_info "Target: 100-500ns | Current: ${cpu_latency}ns"
        echo ""
        log_info "Possible causes:"
        echo "  • CPU governor not set to 'performance'"
        echo "  • CPU isolation not configured"
        echo "  • IRQ interference on CPU1"
        echo "  • Background services causing jitter"
        echo "  • High CPU steal time (hypervisor contention)"
        echo ""
        log_info "🔧 Run diagnostics: ${GREEN}./scripts/diagnose.sh${NC}"
        log_info "🔧 Auto-fix issues: ${GREEN}sudo ./scripts/fix_jitter.sh --fix-all${NC}"
        echo ""
        log_warning "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo ""
        
        if [[ "$IS_ROOT" == true ]]; then
            read -p "Continue anyway? (y/N) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_error "Deployment aborted due to high latency"
                exit 1
            fi
        else
            log_error "Deployment aborted - latency too high for antinuke engine"
            exit 1
        fi
    else
        log_success "CPU latency acceptable: ${cpu_latency}ns"
    fi
    
    # Check CPU governor
    if [[ -f /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor ]]; then
        local governor=$(cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor 2>/dev/null || echo "unknown")
        log_info "CPU governor: $governor"
        if [[ "$governor" == "performance" ]]; then
            log_success "Performance mode active"
        else
            log_warning "CPU governor not in performance mode: $governor"
        fi
    fi
    
    # Memory check
    local mem_available=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
    log_info "Available memory: $((mem_available / 1024))MB"
    
    log_success "Benchmark tests complete"
}

################################################################################
# TASK 3: BUILD BOT
################################################################################

task_3_build_bot() {
    log_step 3 "Build Antinuke Bot"
    
    log_info "Loading environment variables..."
    if [[ -f .env ]]; then
        export $(cat .env | grep -v '^#' | xargs)
        log_success "Environment loaded"
    fi
    
    log_info "Tidying dependencies..."
    go mod tidy
    
    log_info "Building optimized binary..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w" \
        -gcflags="-l=4" \
        -o bin/antinuke \
        ./cmd/main.go
    
    if [[ $? -eq 0 ]]; then
        log_success "Build successful!"
        ls -lh bin/antinuke
    else
        log_error "Build failed!"
        exit 1
    fi
}

################################################################################
# TASK 4: CPU PINNING SETUP
################################################################################

task_4_cpu_pinning() {
    log_step 4 "CPU Pinning & Affinity Setup"
    
    # Create CPU pinning script
    cat > scripts/apply_cpu_affinity.sh <<'CPUPIN'
#!/bin/bash
# Apply CPU affinity to antinuke process

set -euo pipefail

PROCESS_NAME="antinuke"
DEDICATED_CPU=1

# Find process PID
PID=$(pgrep -f "$PROCESS_NAME" | head -n1)

if [[ -z "$PID" ]]; then
    echo "Process $PROCESS_NAME not found"
    exit 1
fi

# Apply CPU affinity
taskset -cp $DEDICATED_CPU $PID

# Set real-time priority (if root)
if [[ $EUID -eq 0 ]]; then
    chrt -f -p 99 $PID
    echo "Applied RT priority to PID $PID"
fi

echo "CPU affinity set: PID $PID → CPU $DEDICATED_CPU"
CPUPIN
    
    chmod +x scripts/apply_cpu_affinity.sh
    log_success "CPU pinning script created"
}

################################################################################
# TASK 5: PM2 INTEGRATION
################################################################################

task_5_pm2_integration() {
    log_step 5 "PM2 Integration Setup"
    
    # Check if PM2 is installed
    if ! command -v pm2 &> /dev/null; then
        log_warning "PM2 not installed - installing..."
        if command -v npm &> /dev/null; then
            npm install -g pm2
        else
            log_error "npm not found - cannot install PM2"
            log_info "Install Node.js and PM2 manually, then re-run"
            return
        fi
    fi
    
    log_info "PM2 version: $(pm2 --version)"
    
    # Create log directory
    if [[ "$IS_ROOT" == true ]]; then
        mkdir -p "$LOG_DIR"
        chmod 755 "$LOG_DIR"
        log_success "Log directory created: $LOG_DIR"
    else
        mkdir -p logs
        log_info "Using local logs directory (not root)"
    fi
    
    # Create PM2 startup hooks
    cat > scripts/post_start_hook.sh <<'POSTSTART'
#!/bin/bash
# Post-start hook for PM2

sleep 2

# Apply CPU affinity
if [[ -f ./scripts/apply_cpu_affinity.sh ]]; then
    ./scripts/apply_cpu_affinity.sh || true
fi

echo "Post-start hook complete"
POSTSTART
    
    cat > scripts/pre_stop_hook.sh <<'PRESTOP'
#!/bin/bash
# Pre-stop hook for PM2

echo "Gracefully shutting down antinuke bot..."
PRESTOP
    
    chmod +x scripts/post_start_hook.sh scripts/pre_stop_hook.sh
    
    log_success "PM2 integration configured"
}

################################################################################
# TASK 6: SYSTEMD INTEGRATION
################################################################################

task_6_systemd_integration() {
    log_step 6 "Systemd Integration"
    
    if [[ "$IS_ROOT" == false ]]; then
        log_warning "Not root - skipping systemd integration"
        return
    fi
    
    # Create systemd service
    cat > /etc/systemd/system/antinuke-pm2.service <<SYSTEMD
[Unit]
Description=Antinuke Bot PM2 Manager
After=network.target

[Service]
Type=forking
User=antinuke
WorkingDirectory=/opt/antinuke
Environment=PATH=/usr/local/bin:/usr/bin:/bin
ExecStart=/usr/bin/pm2 start ecosystem.config.js --env production
ExecReload=/usr/bin/pm2 reload ecosystem.config.js --env production
ExecStop=/usr/bin/pm2 stop ecosystem.config.js
Restart=on-failure
RestartSec=10s

# CPU Affinity
CPUAffinity=$DEDICATED_CPU

# Performance
Nice=-20
IOSchedulingClass=realtime
IOSchedulingPriority=0

# Limits
LimitNOFILE=1048576
LimitNPROC=65535

[Install]
WantedBy=multi-user.target
SYSTEMD
    
    systemctl daemon-reload
    log_success "Systemd service created: antinuke-pm2.service"
    log_info "Enable with: systemctl enable antinuke-pm2"
}

################################################################################
# TASK 7: START WITH PM2
################################################################################

task_7_start_pm2() {
    log_step 7 "Start Bot with PM2"
    
    if ! command -v pm2 &> /dev/null; then
        log_warning "PM2 not available - falling back to direct execution"
        log_info "Starting bot directly..."
        
        # Apply CPU affinity if possible
        if [[ "$IS_ROOT" == true ]] && command -v taskset &> /dev/null; then
            taskset -c $DEDICATED_CPU ./bin/antinuke
        else
            ./bin/antinuke
        fi
        return
    fi
    
    # Stop existing instance
    pm2 stop antinuke-bot 2>/dev/null || true
    pm2 delete antinuke-bot 2>/dev/null || true
    
    log_info "Starting with PM2..."
    pm2 start ecosystem.config.js --env production
    
    # Wait for process to start
    sleep 3
    
    # Apply CPU affinity
    if [[ -f ./scripts/apply_cpu_affinity.sh ]]; then
        log_info "Applying CPU affinity..."
        ./scripts/apply_cpu_affinity.sh || true
    fi
    
    # Show status
    pm2 status
    pm2 logs antinuke-bot --lines 20 --nostream
    
    log_success "Bot started with PM2!"
    log_info "Monitor with: pm2 logs antinuke-bot"
    log_info "Status: pm2 status"
}

################################################################################
# MAIN EXECUTION
################################################################################

main() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  Ultra-Low-Latency Antinuke Bot Deployment"
    echo "  Target: 100-300ns detection | <200μs execution"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    task_1_vm_optimization
    task_2_benchmark_tests
    task_3_build_bot
    task_4_cpu_pinning
    task_5_pm2_integration
    task_6_systemd_integration
    task_7_start_pm2
    
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    log_success "DEPLOYMENT COMPLETE!"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
}

main "$@"
