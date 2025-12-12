#!/bin/bash

################################################################################
# VM Performance Diagnostic Tool
# Analyzes why CPU latency is high and provides fix recommendations
################################################################################

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Thresholds
CPU_LATENCY_THRESHOLD_NS=500
STEAL_TIME_THRESHOLD=1.0
IRQ_COUNT_THRESHOLD=100

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[✓]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }
log_section() { echo -e "\n${CYAN}═══ $1 ═══${NC}\n"; }

ISSUES_FOUND=0
FIXES_AVAILABLE=()

################################################################################
# DIAGNOSTIC A: CPU GOVERNOR CHECK
################################################################################

check_cpu_governor() {
    log_section "CPU GOVERNOR STATUS"
    
    if [[ ! -d /sys/devices/system/cpu/cpu0/cpufreq ]]; then
        log_error "CPU frequency scaling not available on this system"
        log_warning "This VM may not support cpufreq"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
        return
    fi
    
    local governor=""
    if [[ -f /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor ]]; then
        governor=$(cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor)
        log_info "Current governor: $governor"
        
        if [[ "$governor" == "performance" ]]; then
            log_success "CPU governor set to performance mode"
        else
            log_error "CPU governor is NOT in performance mode"
            log_warning "Current: $governor | Required: performance"
            ISSUES_FOUND=$((ISSUES_FOUND + 1))
            FIXES_AVAILABLE+=("fix_cpu_governor")
        fi
    else
        log_error "Cannot read CPU governor status"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    fi
    
    # Check available governors
    if [[ -f /sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors ]]; then
        local available=$(cat /sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors)
        log_info "Available governors: $available"
    fi
}

################################################################################
# DIAGNOSTIC B: CPU ISOLATION CHECK
################################################################################

check_cpu_isolation() {
    log_section "CPU ISOLATION STATUS"
    
    local cmdline=$(cat /proc/cmdline)
    log_info "Boot parameters: $cmdline"
    
    if echo "$cmdline" | grep -q "isolcpus="; then
        local isolcpus=$(echo "$cmdline" | grep -o "isolcpus=[^ ]*")
        log_success "CPU isolation enabled: $isolcpus"
    else
        log_error "CPU isolation NOT configured in boot parameters"
        log_warning "Add 'isolcpus=1 nohz_full=1 rcu_nocbs=1' to GRUB"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
        FIXES_AVAILABLE+=("fix_cpu_isolation")
    fi
    
    # Check nohz_full
    if echo "$cmdline" | grep -q "nohz_full="; then
        log_success "Tickless kernel enabled"
    else
        log_warning "nohz_full not configured (optional but recommended)"
    fi
}

################################################################################
# DIAGNOSTIC C: IRQ AFFINITY CHECK
################################################################################

check_irq_affinity() {
    log_section "IRQ AFFINITY ANALYSIS"
    
    log_info "Checking interrupt distribution across CPUs..."
    
    # Count interrupts per CPU
    local cpu0_irqs=0
    local cpu1_irqs=0
    
    # Parse /proc/interrupts
    while IFS= read -r line; do
        if [[ $line =~ ^[[:space:]]*[0-9]+: ]]; then
            # Extract CPU columns (skip first column which is IRQ number)
            local counts=($(echo "$line" | awk '{for(i=2; i<=NF && $i ~ /^[0-9]+$/; i++) print $i}'))
            if [[ ${#counts[@]} -ge 2 ]]; then
                cpu0_irqs=$((cpu0_irqs + ${counts[0]}))
                cpu1_irqs=$((cpu1_irqs + ${counts[1]}))
            fi
        fi
    done < /proc/interrupts
    
    log_info "CPU0 total interrupts: $cpu0_irqs"
    log_info "CPU1 total interrupts: $cpu1_irqs"
    
    if [[ $cpu1_irqs -gt $IRQ_COUNT_THRESHOLD ]]; then
        log_error "CPU1 has too many interrupts: $cpu1_irqs (threshold: $IRQ_COUNT_THRESHOLD)"
        log_warning "CPU1 should be isolated from IRQs"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
        FIXES_AVAILABLE+=("fix_irq_affinity")
    else
        log_success "CPU1 interrupt count acceptable: $cpu1_irqs"
    fi
    
    # Check specific NIC interrupts
    log_info "\nNetwork interface interrupts:"
    grep -E "virtio|eth|ens" /proc/interrupts | head -5 || log_info "No network IRQs found"
}

################################################################################
# DIAGNOSTIC D: STEAL TIME CHECK
################################################################################

check_steal_time() {
    log_section "CPU STEAL TIME ANALYSIS"
    
    log_info "Measuring CPU steal time (5 second sample)..."
    
    # Use vmstat to check steal time
    if command -v vmstat &> /dev/null; then
        local steal_time=$(vmstat 1 5 | tail -1 | awk '{print $16}')
        log_info "CPU steal time: ${steal_time}%"
        
        if (( $(echo "$steal_time > $STEAL_TIME_THRESHOLD" | bc -l) )); then
            log_error "High CPU steal time detected: ${steal_time}%"
            log_warning "VM is experiencing hypervisor contention"
            log_warning "Consider migrating to different zone or CPU type"
            ISSUES_FOUND=$((ISSUES_FOUND + 1))
        else
            log_success "CPU steal time acceptable: ${steal_time}%"
        fi
    else
        log_warning "vmstat not available, cannot check steal time"
    fi
}

################################################################################
# DIAGNOSTIC E: BACKGROUND SERVICES CHECK
################################################################################

check_background_services() {
    log_section "BACKGROUND SERVICES ANALYSIS"
    
    local jitter_services=(
        "snapd"
        "unattended-upgrades"
        "packagekit"
        "ModemManager"
        "bluetooth"
        "cups"
        "avahi-daemon"
    )
    
    local found_services=0
    for service in "${jitter_services[@]}"; do
        if systemctl is-active "$service" &>/dev/null; then
            log_warning "Jitter-causing service active: $service"
            found_services=$((found_services + 1))
        fi
    done
    
    if [[ $found_services -gt 0 ]]; then
        log_error "Found $found_services active jitter-causing services"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
        FIXES_AVAILABLE+=("fix_background_services")
    else
        log_success "No major jitter-causing services detected"
    fi
}

################################################################################
# DIAGNOSTIC F: TRANSPARENT HUGE PAGES CHECK
################################################################################

check_thp() {
    log_section "TRANSPARENT HUGE PAGES STATUS"
    
    if [[ -f /sys/kernel/mm/transparent_hugepage/enabled ]]; then
        local thp_status=$(cat /sys/kernel/mm/transparent_hugepage/enabled | grep -o '\[.*\]' | tr -d '[]')
        log_info "THP status: $thp_status"
        
        if [[ "$thp_status" == "never" ]]; then
            log_success "Transparent Huge Pages disabled (optimal for low latency)"
        else
            log_warning "THP enabled: $thp_status (can cause latency spikes)"
            FIXES_AVAILABLE+=("fix_thp")
        fi
    else
        log_info "THP not available on this system"
    fi
}

################################################################################
# DIAGNOSTIC G: CPU LATENCY BENCHMARK
################################################################################

check_cpu_latency() {
    log_section "CPU LATENCY BENCHMARK"
    
    log_info "Running CPU latency test (10,000 iterations)..."
    
    local start=$(date +%s%N)
    for i in {1..10000}; do
        :
    done
    local end=$(date +%s%N)
    
    local cpu_latency=$(( (end - start) / 10000 ))
    
    log_info "Average CPU latency: ${cpu_latency}ns"
    
    if [[ $cpu_latency -gt $CPU_LATENCY_THRESHOLD_NS ]]; then
        log_error "CPU latency HIGH: ${cpu_latency}ns (threshold: ${CPU_LATENCY_THRESHOLD_NS}ns)"
        log_warning "Target for antinuke: 100-500ns"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        log_success "CPU latency acceptable: ${cpu_latency}ns"
    fi
}

################################################################################
# DIAGNOSTIC H: SYSTEM LOAD CHECK
################################################################################

check_system_load() {
    log_section "SYSTEM LOAD ANALYSIS"
    
    local load=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | tr -d ',')
    local cpu_count=$(nproc)
    
    log_info "System load (1m): $load"
    log_info "CPU count: $cpu_count"
    
    if (( $(echo "$load > $cpu_count" | bc -l) )); then
        log_warning "System load high: $load (CPUs: $cpu_count)"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        log_success "System load normal: $load"
    fi
}

################################################################################
# DIAGNOSTIC I: MEMORY PRESSURE CHECK
################################################################################

check_memory_pressure() {
    log_section "MEMORY PRESSURE ANALYSIS"
    
    local mem_available=$(grep MemAvailable /proc/meminfo | awk '{print $2}')
    local mem_total=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    local mem_percent=$((100 * mem_available / mem_total))
    
    log_info "Available memory: $((mem_available / 1024))MB / $((mem_total / 1024))MB"
    log_info "Memory available: ${mem_percent}%"
    
    if [[ $mem_percent -lt 20 ]]; then
        log_error "Low memory: ${mem_percent}% available"
        ISSUES_FOUND=$((ISSUES_FOUND + 1))
    else
        log_success "Memory pressure acceptable: ${mem_percent}% available"
    fi
    
    # Check swap usage
    local swap_used=$(grep SwapTotal /proc/meminfo | awk '{print $2}')
    if [[ $swap_used -gt 0 ]]; then
        local swap_free=$(grep SwapFree /proc/meminfo | awk '{print $2}')
        local swap_in_use=$((swap_used - swap_free))
        if [[ $swap_in_use -gt 0 ]]; then
            log_warning "Swap in use: $((swap_in_use / 1024))MB"
        fi
    fi
}

################################################################################
# DIAGNOSTIC J: KERNEL VERSION CHECK
################################################################################

check_kernel_version() {
    log_section "KERNEL VERSION INFO"
    
    local kernel_version=$(uname -r)
    log_info "Kernel version: $kernel_version"
    
    # Check for real-time patches
    if uname -r | grep -q "rt"; then
        log_success "Real-time kernel detected"
    else
        log_info "Standard kernel (RT patches not detected)"
    fi
}

################################################################################
# SUMMARY AND RECOMMENDATIONS
################################################################################

show_summary() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  DIAGNOSTIC SUMMARY"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    if [[ $ISSUES_FOUND -eq 0 ]]; then
        log_success "NO ISSUES FOUND - System is optimized!"
        echo ""
        log_info "Your VM is ready for ultra-low-latency operations"
    else
        log_error "FOUND $ISSUES_FOUND ISSUES"
        echo ""
        log_warning "The following issues need to be addressed:"
        echo ""
        
        if [[ ${#FIXES_AVAILABLE[@]} -gt 0 ]]; then
            log_info "Available automatic fixes:"
            for fix in "${FIXES_AVAILABLE[@]}"; do
                case $fix in
                    fix_cpu_governor)
                        echo "  • CPU Governor → Run: ./scripts/fix_jitter.sh --fix-governor"
                        ;;
                    fix_cpu_isolation)
                        echo "  • CPU Isolation → Run: ./scripts/fix_jitter.sh --fix-isolation"
                        ;;
                    fix_irq_affinity)
                        echo "  • IRQ Affinity → Run: ./scripts/fix_jitter.sh --fix-irq"
                        ;;
                    fix_background_services)
                        echo "  • Background Services → Run: ./scripts/fix_jitter.sh --fix-services"
                        ;;
                    fix_thp)
                        echo "  • Transparent Huge Pages → Run: ./scripts/fix_jitter.sh --fix-thp"
                        ;;
                esac
            done
            echo ""
            log_info "Or run all fixes: sudo ./scripts/fix_jitter.sh --fix-all"
        fi
    fi
    
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
}

################################################################################
# MAIN EXECUTION
################################################################################

main() {
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "  VM Performance Diagnostic Tool"
    echo "  Target: <500ns CPU latency for antinuke bot"
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    check_cpu_governor
    check_cpu_isolation
    check_irq_affinity
    check_steal_time
    check_background_services
    check_thp
    check_cpu_latency
    check_system_load
    check_memory_pressure
    check_kernel_version
    
    show_summary
    
    exit $ISSUES_FOUND
}

main "$@"
