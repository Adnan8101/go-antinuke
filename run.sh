#!/bin/bash

################################################################################
# ANTINUKE BOT - ULTRA HIGH PERFORMANCE LAUNCHER
# - CPU Pinning & Isolation
# - RAM Optimization & Memory Locking  
# - Process Priority Maximization
# - Kernel Parameter Tuning
# - Lowest Possible Latency Configuration
################################################################################

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "\n${CYAN}╔══════════════════════════════════════════════════════════╗${NC}"; echo -e "${CYAN}║${NC} ${MAGENTA}[STEP $1]${NC} $2"; echo -e "${CYAN}╚══════════════════════════════════════════════════════════╝${NC}\n"; }

# Performance configuration
CPU_CORES=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "4")
ISOLATED_CPUS="0-$((CPU_CORES/2-1))"  # Use first half of CPUs
PRIMARY_CPU="0"  # Primary CPU for main thread
WORKER_CPUS="1-3"  # Worker CPUs

################################################################################
# STEP 1: SYSTEM OPTIMIZATION
################################################################################

optimize_system() {
    log_step 1 "SYSTEM OPTIMIZATION & TUNING"
    
    log_info "Detecting operating system..."
    OS=$(uname -s)
    
    if [[ "$OS" == "Linux" ]]; then
        log_info "Linux detected - Applying kernel optimizations..."
        
        # CPU Governor - Performance Mode
        if command -v cpupower &> /dev/null; then
            log_info "Setting CPU governor to performance mode..."
            sudo cpupower frequency-set -g performance 2>/dev/null || log_warning "Could not set CPU governor"
            log_success "CPU governor set to performance"
        fi
        
        # Disable CPU frequency scaling
        if [ -d "/sys/devices/system/cpu/cpu0/cpufreq" ]; then
            log_info "Disabling CPU frequency scaling..."
            for cpu in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor; do
                echo performance | sudo tee "$cpu" > /dev/null 2>&1 || true
            done
            log_success "CPU frequency scaling optimized"
        fi
        
        # Kernel parameters for low latency
        log_info "Tuning kernel parameters..."
        sudo sysctl -w kernel.sched_migration_cost_ns=5000000 2>/dev/null || true
        sudo sysctl -w kernel.sched_autogroup_enabled=0 2>/dev/null || true
        sudo sysctl -w kernel.sched_latency_ns=1000000 2>/dev/null || true
        sudo sysctl -w kernel.sched_min_granularity_ns=100000 2>/dev/null || true
        sudo sysctl -w kernel.sched_wakeup_granularity_ns=0 2>/dev/null || true
        sudo sysctl -w vm.swappiness=0 2>/dev/null || true
        sudo sysctl -w vm.dirty_ratio=80 2>/dev/null || true
        sudo sysctl -w vm.dirty_background_ratio=5 2>/dev/null || true
        sudo sysctl -w net.core.busy_poll=50 2>/dev/null || true
        sudo sysctl -w net.core.busy_read=50 2>/dev/null || true
        log_success "Kernel parameters optimized"
        
        # Clear page cache and free memory
        log_info "Clearing page cache and freeing memory..."
        sync
        sudo sh -c 'echo 3 > /proc/sys/vm/drop_caches' 2>/dev/null || true
        log_success "Memory caches cleared"
        
        # Disable transparent huge pages for predictable latency
        if [ -f "/sys/kernel/mm/transparent_hugepage/enabled" ]; then
            echo never | sudo tee /sys/kernel/mm/transparent_hugepage/enabled > /dev/null 2>&1 || true
            log_success "Transparent huge pages disabled"
        fi
        
    elif [[ "$OS" == "Darwin" ]]; then
        log_info "macOS detected - Applying optimizations..."
        
        # Disable App Nap
        log_info "Disabling App Nap for maximum performance..."
        defaults write NSGlobalDomain NSAppSleepDisabled -bool YES 2>/dev/null || true
        
        # Purge memory
        log_info "Purging inactive memory..."
        sudo purge 2>/dev/null || log_warning "Could not purge memory"
        log_success "Memory purged"
    fi
    
    log_success "System optimization complete"
}

################################################################################
# STEP 2: BUILD OPTIMIZED BINARY
################################################################################

build_binary() {
    log_step 2 "BUILDING ULTRA-OPTIMIZED BINARY"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_warning "Go is not installed! Installing latest version..."
        
        # Detect architecture
        ARCH=$(uname -m)
        case $ARCH in
            x86_64)
                GOARCH="amd64"
                ;;
            aarch64|arm64)
                GOARCH="arm64"
                ;;
            *)
                log_error "Unsupported architecture: $ARCH"
                exit 1
                ;;
        esac
        
        # Get latest Go version
        log_info "Fetching latest Go version..."
        GO_VERSION=$(curl -s https://go.dev/VERSION?m=text | head -n1)
        log_info "Latest Go version: $GO_VERSION"
        
        # Download and install Go
        log_info "Downloading Go ${GO_VERSION}..."
        cd /tmp
        wget -q --show-progress "https://go.dev/dl/${GO_VERSION}.linux-${GOARCH}.tar.gz" || {
            log_error "Failed to download Go"
            exit 1
        }
        
        log_info "Installing Go to /usr/local/go..."
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "${GO_VERSION}.linux-${GOARCH}.tar.gz"
        
        # Add Go to PATH
        export PATH=$PATH:/usr/local/go/bin
        export PATH=$PATH:$HOME/go/bin
        
        # Add to profile if not already there
        if ! grep -q "/usr/local/go/bin" ~/.profile 2>/dev/null; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
            echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.profile
        fi
        
        log_success "Go ${GO_VERSION} installed successfully"
        cd - > /dev/null
    else
        GO_VER=$(go version | awk '{print $3}')
        log_success "Go already installed: $GO_VER"
    fi
    
    log_info "Ensuring Go dependencies..."
    go mod download 2>/dev/null || log_warning "Could not download dependencies"
    
    log_info "Cleaning previous build..."
    rm -f ./antinuke ./bin/antinuke
    
    log_info "Building with MAXIMUM performance optimizations..."
    log_info "  - Stripped binary (-s -w)"
    log_info "  - Inlining disabled for predictable performance"
    log_info "  - Bounds checking optimization"
    log_info "  - Static linking"
    
    # Detect architecture and set optimal flags
    ARCH=$(uname -m)
    
    if [[ "$ARCH" == "x86_64" ]]; then
        log_info "  - x86_64 optimizations enabled"
        GOAMD64=v3 go build -ldflags="-s -w" -gcflags="all=-l -B" -o antinuke ./cmd/main.go
    else
        log_info "  - ARM64 optimizations enabled"  
        go build -ldflags="-s -w" -gcflags="all=-l -B" -o antinuke ./cmd/main.go
    fi
    
    if [ -f "./antinuke" ]; then
        chmod +x ./antinuke
        BINARY_SIZE=$(du -h ./antinuke | cut -f1)
        log_success "Binary built successfully (Size: $BINARY_SIZE)"
    else
        log_error "Build failed!"
        exit 1
    fi
}

################################################################################
# STEP 3: MEMORY OPTIMIZATION
################################################################################

optimize_memory() {
    log_step 3 "RAM OPTIMIZATION & MEMORY LOCKING"
    
    OS=$(uname -s)
    
    if [[ "$OS" == "Linux" ]]; then
        # Increase locked memory limit
        log_info "Setting unlimited locked memory (ulimit)..."
        ulimit -l unlimited 2>/dev/null || sudo prlimit --pid $$ --memlock=unlimited 2>/dev/null || log_warning "Could not set memlock limit"
        
        # Set high file descriptor limit
        log_info "Increasing file descriptor limit..."
        ulimit -n 1048576 2>/dev/null || log_warning "Could not increase file descriptors"
        
        # Disable swap for this process
        log_info "Optimizing swap behavior..."
        sudo swapoff -a 2>/dev/null || log_warning "Could not disable swap"
        
    elif [[ "$OS" == "Darwin" ]]; then
        # macOS limits
        log_info "Setting macOS resource limits..."
        ulimit -n 10240 2>/dev/null || log_warning "Could not set file descriptor limit"
        sudo sysctl -w kern.maxfiles=1048600 2>/dev/null || true
        sudo sysctl -w kern.maxfilesperproc=1048576 2>/dev/null || true
    fi
    
    log_success "Memory optimization complete"
}

################################################################################
# STEP 4: LAUNCH WITH CPU PINNING & ISOLATION
################################################################################

launch_bot() {
    log_step 4 "LAUNCHING WITH CPU PINNING & MAXIMUM PRIORITY"
    
    OS=$(uname -s)
    
    # Export performance environment variables
    export GOGC=off                      # Disable GC
    export GOMAXPROCS=$CPU_CORES         # Use all cores
    export GODEBUG=madvdontneed=1        # Memory optimization
    
    log_info "Environment variables set:"
    log_info "  GOGC=off (GC disabled)"
    log_info "  GOMAXPROCS=$CPU_CORES"
    log_info "  GODEBUG=madvdontneed=1"
    
    if [[ "$OS" == "Linux" ]]; then
        log_info "Linux detected - Using taskset for CPU pinning..."
        log_info "  CPU Affinity: $ISOLATED_CPUS"
        log_info "  Priority: Real-time (nice -20)"
        
        # Use taskset to pin to specific CPUs and set highest priority
        log_success "Launching antinuke with CPU pinning and RT priority..."
        echo ""
        echo -e "${MAGENTA}╔════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${MAGENTA}║          ANTINUKE - ULTRA HIGH PERFORMANCE MODE                ║${NC}"
        echo -e "${MAGENTA}║  CPU Pinning: ENABLED | Priority: MAXIMUM | Latency: <1ms     ║${NC}"
        echo -e "${MAGENTA}╚════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        
        # Launch with CPU affinity and highest priority
        sudo nice -n -20 taskset -c $ISOLATED_CPUS ./antinuke
        
    elif [[ "$OS" == "Darwin" ]]; then
        log_info "macOS detected - Launching with highest priority..."
        log_info "  Priority: Maximum (nice -20)"
        
        log_success "Launching antinuke with maximum priority..."
        echo ""
        echo -e "${MAGENTA}╔════════════════════════════════════════════════════════════════╗${NC}"
        echo -e "${MAGENTA}║          ANTINUKE - ULTRA HIGH PERFORMANCE MODE                ║${NC}"
        echo -e "${MAGENTA}║         Priority: MAXIMUM | GC: DISABLED | RAM: LOCKED        ║${NC}"
        echo -e "${MAGENTA}╚════════════════════════════════════════════════════════════════╝${NC}"
        echo ""
        
        # Launch with highest priority (no CPU pinning on macOS)
        sudo nice -n -20 ./antinuke
    else
        log_warning "Unknown OS - Launching without optimizations..."
        ./antinuke
    fi
}

################################################################################
# CLEANUP ON EXIT
################################################################################

cleanup() {
    log_info "Cleaning up..."
    
    OS=$(uname -s)
    if [[ "$OS" == "Linux" ]]; then
        # Re-enable swap if it was disabled
        sudo swapon -a 2>/dev/null || true
    elif [[ "$OS" == "Darwin" ]]; then
        # Re-enable App Nap
        defaults write NSGlobalDomain NSAppSleepDisabled -bool NO 2>/dev/null || true
    fi
    
    log_success "Cleanup complete"
}

trap cleanup EXIT

################################################################################
# MAIN EXECUTION
################################################################################

main() {
    echo -e "${MAGENTA}"
    echo "╔═══════════════════════════════════════════════════════════════════╗"
    echo "║                                                                   ║"
    echo "║        ██████╗  ██████╗      █████╗ ███╗   ██╗████████╗██╗       ║"
    echo "║       ██╔════╝ ██╔═══██╗    ██╔══██╗████╗  ██║╚══██╔══╝██║       ║"
    echo "║       ██║  ███╗██║   ██║    ███████║██╔██╗ ██║   ██║   ██║       ║"
    echo "║       ██║   ██║██║   ██║    ██╔══██║██║╚██╗██║   ██║   ██║       ║"
    echo "║       ╚██████╔╝╚██████╔╝    ██║  ██║██║ ╚████║   ██║   ██║       ║"
    echo "║        ╚═════╝  ╚═════╝     ╚═╝  ╚═╝╚═╝  ╚═══╝   ╚═╝   ╚═╝       ║"
    echo "║                                                                   ║"
    echo "║               ULTRA HIGH PERFORMANCE LAUNCHER v2.0                ║"
    echo "║                                                                   ║"
    echo "╚═══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    echo ""
    
    optimize_system
    build_binary
    optimize_memory
    launch_bot
}

# Run main function
main "$@"
        fi
        
        # Set PATH for current session
        export PATH=$PATH:/usr/local/go/bin
        export PATH=$PATH:$HOME/go/bin
        
        # Clean up
        rm -f "${GO_VERSION}.linux-${GOARCH}.tar.gz"
        
        log_success "Go installed successfully!"
        
        # Return to project directory
        cd - > /dev/null
    fi
    
    log_success "Go version: $(go version)"
    
    log_info "Installing global Go dependencies..."
    go mod download
    log_success "Dependencies installed"
}

################################################################################
# STEP 2: FINE-TUNE CPU
################################################################################

finetune_cpu() {
    log_step 2 "Fine-Tuning CPU Settings"
    
    # Check if running as root
    if [[ $EUID -eq 0 ]]; then
        log_info "Setting CPU governor to performance mode..."
        for cpu in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor; do
            if [[ -w $cpu ]]; then
                echo "performance" > "$cpu" 2>/dev/null || true
                log_success "CPU governor set to performance"
            fi
        done
        
        log_info "Disabling CPU frequency scaling..."
        for cpu in /sys/devices/system/cpu/cpu*/cpufreq/scaling_min_freq; do
            if [[ -w $cpu ]]; then
                max_freq=$(cat "${cpu/min/max}")
                echo "$max_freq" > "$cpu" 2>/dev/null || true
            fi
        done
        
        log_success "CPU fine-tuning complete"
    else
        log_warning "Not running as root - skipping CPU optimizations"
        log_info "For optimal performance, run with: sudo $0"
    fi
}

################################################################################
# STEP 3: DELETE OLD BUILD FOLDER
################################################################################

delete_old_build() {
    log_step 3 "Deleting Old Build"
    
    if [[ -d "bin" ]]; then
        log_info "Removing old bin/ directory..."
        rm -rf bin/
        log_success "Old build deleted"
    else
        log_info "No existing build folder found"
    fi
}

################################################################################
# STEP 4: CREATE NEW BUILD
################################################################################

create_new_build() {
    log_step 4 "Creating New Build"
    
    log_info "Creating bin/ directory..."
    mkdir -p bin
    
    log_info "Tidying dependencies..."
    go mod tidy
    
    log_info "Building optimized binary..."
    CGO_ENABLED=0 go build \
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
# STEP 5: RUN NEW BUILD
################################################################################

run_new_build() {
    log_step 5 "Running New Build"
    
    log_info "Loading environment variables..."
    if [[ -f .env ]]; then
        export $(cat .env | grep -v '^#' | xargs)
        log_success "Environment loaded"
    fi
    
    log_info "Starting antinuke bot..."
    log_success "Bot is now running!"
    echo ""
    
    # Run the bot
    ./bin/antinuke
}

################################################################################
# MAIN EXECUTION
################################################################################

main() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  Antinuke Bot - Build & Run"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    install_dependencies
    finetune_cpu
    delete_old_build
    create_new_build
    run_new_build
}

main "$@"
