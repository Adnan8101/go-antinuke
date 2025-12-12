#!/bin/bash

################################################################################
# CPU Jitter Auto-Fix Script
# Automatically fixes common issues causing high CPU latency
################################################################################

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_section() { echo -e "\n${CYAN}═══ $1 ═══${NC}\n"; }

# Check root
if [[ $EUID -ne 0 ]]; then
   log_error "This script must be run as root (use sudo)"
   exit 1
fi

################################################################################
# FIX 1: CPU GOVERNOR
################################################################################

fix_cpu_governor() {
    log_section "FIXING CPU GOVERNOR"
    
    # Install cpufrequtils if not present
    if ! command -v cpupower &> /dev/null; then
        log_info "Installing CPU frequency tools..."
        
        if command -v apt-get &> /dev/null; then
            apt-get update -qq
            apt-get install -y linux-tools-common linux-tools-$(uname -r) 2>/dev/null || {
                apt-get install -y cpufrequtils
            }
        elif command -v yum &> /dev/null; then
            yum install -y kernel-tools
        fi
    fi
    
    # Set performance governor
    log_info "Setting CPU governor to performance..."
    
    # Method 1: Using cpupower
    if command -v cpupower &> /dev/null; then
        cpupower frequency-set -g performance >/dev/null 2>&1 || true
    fi
    
    # Method 2: Direct sysfs write
    for cpu in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor; do
        if [[ -w $cpu ]]; then
            echo "performance" > "$cpu" 2>/dev/null || true
        fi
    done
    
    # Verify
    if [[ -f /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor ]]; then
        local governor=$(cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor)
        if [[ "$governor" == "performance" ]]; then
            log_success "CPU governor set to performance"
        else
            log_error "Failed to set CPU governor"
            return 1
        fi
    else
        log_warning "CPU frequency scaling not available on this system"
    fi
    
    # Disable CPU frequency scaling (if intel_pstate)
    if [[ -f /sys/devices/system/cpu/intel_pstate/no_turbo ]]; then
        echo 0 > /sys/devices/system/cpu/intel_pstate/no_turbo 2>/dev/null || true
        log_info "Intel turbo boost enabled"
    fi
}

################################################################################
# FIX 2: CPU ISOLATION
################################################################################

fix_cpu_isolation() {
    log_section "FIXING CPU ISOLATION"
    
    local cmdline=$(cat /proc/cmdline)
    
    if echo "$cmdline" | grep -q "isolcpus="; then
        log_success "CPU isolation already configured"
        return 0
    fi
    
    log_warning "CPU isolation requires GRUB configuration and reboot"
    log_info "Adding boot parameters..."
    
    # Backup GRUB config
    if [[ -f /etc/default/grub ]]; then
        cp /etc/default/grub /etc/default/grub.backup.$(date +%s)
        
        # Add isolcpus parameters
        if grep -q "GRUB_CMDLINE_LINUX=" /etc/default/grub; then
            sed -i 's/GRUB_CMDLINE_LINUX="\(.*\)"/GRUB_CMDLINE_LINUX="\1 isolcpus=1 nohz_full=1 rcu_nocbs=1"/' /etc/default/grub
            log_success "GRUB configuration updated"
        else
            log_error "Could not find GRUB_CMDLINE_LINUX in /etc/default/grub"
            return 1
        fi
        
        # Update GRUB
        if command -v update-grub &> /dev/null; then
            update-grub
        elif command -v grub2-mkconfig &> /dev/null; then
            grub2-mkconfig -o /boot/grub2/grub.cfg
        else
            log_error "Could not find GRUB update command"
            return 1
        fi
        
        log_warning "⚠️  REBOOT REQUIRED for CPU isolation to take effect"
    else
        log_error "GRUB configuration file not found"
        return 1
    fi
}

################################################################################
# FIX 3: IRQ AFFINITY
################################################################################

fix_irq_affinity() {
    log_section "FIXING IRQ AFFINITY"
    
    log_info "Moving all IRQs to CPU0..."
    
    local fixed_count=0
    for irq in /proc/irq/*/smp_affinity; do
        if [[ -w $irq ]]; then
            echo "1" > "$irq" 2>/dev/null && fixed_count=$((fixed_count + 1)) || true
        fi
    done
    
    log_success "Configured $fixed_count IRQ affinities to CPU0"
    
    # Verify
    log_info "Verifying IRQ distribution..."
    sleep 2
    
    local cpu1_irqs=$(awk '/CPU1/ {sum+=$2} END {print sum}' /proc/interrupts 2>/dev/null || echo "0")
    log_info "CPU1 interrupt count: $cpu1_irqs"
}

################################################################################
# FIX 4: BACKGROUND SERVICES
################################################################################

fix_background_services() {
    log_section "FIXING BACKGROUND SERVICES"
    
    local jitter_services=(
        "snapd"
        "unattended-upgrades"
        "packagekit"
        "ModemManager"
        "bluetooth"
        "cups"
        "avahi-daemon"
    )
    
    log_info "Disabling jitter-causing services..."
    
    local disabled_count=0
    for service in "${jitter_services[@]}"; do
        if systemctl is-active "$service" &>/dev/null; then
            systemctl stop "$service" 2>/dev/null || true
            systemctl disable "$service" 2>/dev/null || true
            log_info "Disabled: $service"
            disabled_count=$((disabled_count + 1))
        fi
    done
    
    if [[ $disabled_count -gt 0 ]]; then
        log_success "Disabled $disabled_count jitter-causing services"
    else
        log_info "No jitter-causing services were active"
    fi
}

################################################################################
# FIX 5: TRANSPARENT HUGE PAGES
################################################################################

fix_thp() {
    log_section "FIXING TRANSPARENT HUGE PAGES"
    
    if [[ -f /sys/kernel/mm/transparent_hugepage/enabled ]]; then
        echo never > /sys/kernel/mm/transparent_hugepage/enabled
        echo never > /sys/kernel/mm/transparent_hugepage/defrag
        log_success "Transparent Huge Pages disabled"
    else
        log_info "THP not available on this system"
    fi
}

################################################################################
# FIX 6: NETWORK TUNING
################################################################################

fix_network_tuning() {
    log_section "OPTIMIZING NETWORK STACK"
    
    log_info "Finding primary network interface..."
    local primary_nic=$(ip route | grep default | awk '{print $5}' | head -n1)
    
    if [[ -n "$primary_nic" ]]; then
        log_info "Primary NIC: $primary_nic"
        
        # Increase ring buffers
        ethtool -G "$primary_nic" rx 4096 tx 4096 2>/dev/null && \
            log_success "Increased ring buffer sizes" || \
            log_warning "Could not modify ring buffers"
        
        # Set interrupt coalescing for low latency
        ethtool -C "$primary_nic" rx-usecs 0 tx-usecs 0 2>/dev/null && \
            log_success "Disabled interrupt coalescing (low latency mode)" || \
            log_warning "Could not modify interrupt coalescing"
    else
        log_warning "Could not detect primary network interface"
    fi
}

################################################################################
# FIX 7: PROCESS PRIORITY
################################################################################

fix_process_priority() {
    log_section "CONFIGURING PROCESS PRIORITIES"
    
    log_info "Creating user limits configuration..."
    
    cat > /etc/security/limits.d/99-antinuke-realtime.conf <<EOF
# Real-time priorities for antinuke bot
* soft rtprio 99
* hard rtprio 99
* soft nice -20
* hard nice -20
* soft memlock unlimited
* hard memlock unlimited
EOF
    
    log_success "Real-time priority limits configured"
}

################################################################################
# FIX ALL
################################################################################

fix_all() {
    log_info "Running all available fixes..."
    echo ""
    
    fix_cpu_governor
    fix_irq_affinity
    fix_background_services
    fix_thp
    fix_network_tuning
    fix_process_priority
    
    # CPU isolation requires reboot, so do it last
    fix_cpu_isolation
}

################################################################################
# USAGE
################################################################################

show_usage() {
    cat <<EOF
Usage: $0 [OPTION]

Auto-fix CPU jitter issues for ultra-low-latency antinuke bot

Options:
    --fix-governor      Fix CPU governor (set to performance)
    --fix-isolation     Fix CPU isolation (requires reboot)
    --fix-irq           Fix IRQ affinity (pin to CPU0)
    --fix-services      Disable jitter-causing background services
    --fix-thp           Disable Transparent Huge Pages
    --fix-network       Optimize network stack for low latency
    --fix-priority      Configure real-time process priorities
    --fix-all           Apply all fixes
    --help              Show this help message

Example:
    sudo $0 --fix-all

EOF
}

################################################################################
# MAIN
################################################################################

main() {
    if [[ $# -eq 0 ]]; then
        show_usage
        exit 1
    fi
    
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  CPU Jitter Auto-Fix Script"
    echo "  Target: <500ns latency for antinuke bot"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    case "$1" in
        --fix-governor)
            fix_cpu_governor
            ;;
        --fix-isolation)
            fix_cpu_isolation
            ;;
        --fix-irq)
            fix_irq_affinity
            ;;
        --fix-services)
            fix_background_services
            ;;
        --fix-thp)
            fix_thp
            ;;
        --fix-network)
            fix_network_tuning
            ;;
        --fix-priority)
            fix_process_priority
            ;;
        --fix-all)
            fix_all
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
    
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    log_success "FIX COMPLETE"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    log_info "Run diagnostics again: ./scripts/diagnose.sh"
    echo ""
}

main "$@"
