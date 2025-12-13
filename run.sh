#!/bin/bash

################################################################################
# Antinuke Bot Build & Run Script
# 1. Install dependencies globally
# 2. Fine-tune CPU
# 3. Delete old build folder
# 4. Create new build
# 5. Run the new build
################################################################################

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_step() { echo -e "\n${CYAN}[STEP $1/5]${NC} $2\n"; }

################################################################################
# STEP 1: INSTALL DEPENDENCIES GLOBALLY
################################################################################

install_dependencies() {
    log_step 1 "Installing Dependencies Globally"
    
    log_info "Checking for Go installation..."
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
        wget -q --show-progress "https://go.dev/dl/${GO_VERSION}.linux-${GOARCH}.tar.gz"
        
        log_info "Installing Go to /usr/local/go..."
        rm -rf /usr/local/go
        tar -C /usr/local -xzf "${GO_VERSION}.linux-${GOARCH}.tar.gz"
        
        # Add Go to PATH
        if ! grep -q "/usr/local/go/bin" /etc/profile; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
            echo 'export PATH=$PATH:$HOME/go/bin' >> /etc/profile
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
