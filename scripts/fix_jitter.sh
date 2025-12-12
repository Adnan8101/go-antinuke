#!/usr/bin/env bash
# fix_jitter.sh
# Purpose: apply jitter / IRQ / CPU isolation fixes for antinuke VM
# Usage:
#   sudo ./fix_jitter.sh --fix-all        # apply everything (no automatic reboot)
#   sudo ./fix_jitter.sh --fix-irq        # only set IRQ affinity -> CPU0
#   sudo ./fix_jitter.sh --fix-isolation  # add GRUB boot params (requires reboot)
#   sudo ./fix_jitter.sh --undo           # attempt to roll back changes
#   ./fix_jitter.sh --dry-run             # show what would be done (no change)

set -euo pipefail
SELF="$(readlink -f "$0" 2>/dev/null || realpath "$0" 2>/dev/null || echo "$0")"
DRY=0
DO_FIX_IRQ=0
DO_FIX_ISOLATION=0
DO_FIX_ALL=0
DO_UNDO=0
DO_REBOOT=0

# Config - adjust if your service name or preferred core numbers differ
SERVICE_NAME="antinuke.service"
BOT_USER="byte_main_1"
BOT_BIN="/home/${BOT_USER}/antinuke"
ISOLATED_CPU=1     # CPU reserved for bot
SYSTEM_CPU=0       # CPU to put IRQs on
RPS_CPU_MASK="1"   # CPU mask for RPS/XPS (binary mask, '1' -> CPU0)

usage() {
  cat <<EOF
fix_jitter.sh - auto-fix jitter / IRQ affinity / CPU isolation for antinuke

Options:
  --fix-all          Apply IRQ affinity + systemd CPU pin + suggest GRUB isolation (no reboot)
  --fix-irq          Only set IRQ affinity to CPU${SYSTEM_CPU} and apply RPS/XPS
  --fix-isolation    Only update GRUB to include isolcpus/nohz_full/rcu_nocbs (requires reboot)
  --reboot           Reboot after modifying GRUB (use with --fix-isolation)
  --undo             Attempt to undo changes made by this script
  --dry-run          Show actions but don't change anything
  -h, --help         Show this help

Example:
  sudo ./fix_jitter.sh --fix-all
EOF
  exit 1
}

# parse args
while (( "$#" )); do
  case "$1" in
    --fix-all) DO_FIX_ALL=1 ;;
    --fix-irq) DO_FIX_IRQ=1 ;;
    --fix-isolation) DO_FIX_ISOLATION=1 ;;
    --undo) DO_UNDO=1 ;;
    --dry-run) DRY=1 ;;
    --reboot) DO_REBOOT=1 ;;
    -h|--help) usage ;;
    *) echo "Unknown arg: $1"; usage ;;
  esac
  shift
done

if [[ $DO_UNDO -eq 1 && ( $DO_FIX_ALL -eq 1 || $DO_FIX_IRQ -eq 1 || $DO_FIX_ISOLATION -eq 1 ) ]]; then
  echo "Cannot combine --undo with other actions"; exit 1
fi
if [[ $DO_FIX_ALL -eq 0 && $DO_FIX_IRQ -eq 0 && $DO_FIX_ISOLATION -eq 0 && $DO_UNDO -eq 0 ]]; then
  usage
fi

echo "=== fix_jitter.sh starting ==="
echo "Dry run: $DRY"
if (( EUID != 0 )); then
  echo "This script must be run as root (sudo). Exiting."; exit 1
fi

# helper to run or print
run() {
  if [[ $DRY -eq 1 ]]; then
    echo "[DRY] $*"
  else
    echo "[RUN] $*"
    eval "$@"
  fi
}

# ensure bc for diagnostic math (diagnose.sh expected it)
if ! command -v bc >/dev/null 2>&1; then
  echo "[INFO] Installing bc for diagnostics..."
  run apt-get update -y >/dev/null 2>&1 || true
  run apt-get install -y bc >/dev/null 2>&1 || true
fi

# -------------------------
# Function: disable irqbalance
# -------------------------
disable_irqbalance() {
  echo "[STEP] Disabling irqbalance service (if present)"
  if systemctl list-unit-files | grep -q irqbalance; then
    run systemctl stop irqbalance || true
    run systemctl disable irqbalance || true
    # also comment /etc/default/irqbalance if exists
    if [[ -f /etc/default/irqbalance ]]; then
      run cp -n /etc/default/irqbalance /etc/default/irqbalance.bak || true
      run sed -i -E 's/^(IRQBALANCE_BANNED_CPUS=).*/\1"$(printf "%d" $((1<<'"$ISOLATED_CPU"')) )"/' /etc/default/irqbalance 2>/dev/null || true
    fi
  else
    echo "[INFO] irqbalance not installed; skipping"
  fi
}

# -------------------------
# Function: set IRQ affinity to SYSTEM_CPU (cpu0)
# -------------------------
fix_irq_affinity() {
  echo "[STEP] Setting IRQ affinity: move device IRQs to CPU${SYSTEM_CPU}"
  # list IRQs that look like virtio / network controllers
  # We'll prioritize PCI-MSIX lines and network-related IRQs
  # parse /proc/interrupts lines containing virtio or eth or enp or ens
  mapfile -t irq_lines < <(grep -E "virtio|eth|enp|ens|net" /proc/interrupts || true)
  if [[ ${#irq_lines[@]} -eq 0 ]]; then
    # fallback: scan numeric IRQ ids and try to detect associated handlers
    mapfile -t irq_lines < <(awk '/PCI-MSIX|virtio|eth|enp|ens|net/ {print NR ":" $0}' /proc/interrupts || true)
  fi

  echo "[INFO] Found ${#irq_lines[@]} irq-related lines to consider"
  for line in "${irq_lines[@]}"; do
    # line begins with IRQ number
    irqnum=$(echo "$line" | awk '{print $1}' | tr -d ':')
    # sanity check numeric
    if ! [[ "$irqnum" =~ ^[0-9]+$ ]]; then continue; fi
    # build cpu mask for SYSTEM_CPU (hex)
    # mask = 1 << SYSTEM_CPU
    mask=$(printf "%x" $((1 << SYSTEM_CPU)))
    if [[ $DRY -eq 1 ]]; then
      echo "[DRY] Would set /proc/irq/$irqnum/smp_affinity -> $mask"
    else
      if [[ -w /proc/irq/$irqnum/smp_affinity ]]; then
        echo "$mask" > /proc/irq/$irqnum/smp_affinity || echo "[WARN] failed to write affinity for irq $irqnum"
        echo "[OK] IRQ $irqnum -> cpu${SYSTEM_CPU} (mask $mask)"
      else
        echo "[WARN] cannot write to /proc/irq/$irqnum/smp_affinity; skipping"
      fi
    fi
  done

  # Additionally set any existing virtio IRQ groups to CPU0 mask
  for irq in $(awk '/PCI-MSIX/ {print $1}' /proc/interrupts | tr -d ':'); do
    mask=$(printf "%x" $((1 << SYSTEM_CPU)))
    if [[ -w /proc/irq/$irq/smp_affinity ]]; then
      run bash -c "echo $mask > /proc/irq/$irq/smp_affinity" || true
    fi
  done

  echo "[INFO] IRQ affinity applied. Verify via: cat /proc/interrupts"
}

# -------------------------
# Function: set RPS/XPS
# -------------------------
fix_rps_xps() {
  echo "[STEP] Setting RPS/XPS to route network processing to CPU${SYSTEM_CPU}"
  for iface in $(ls /sys/class/net | grep -v lo); do
    # set all rx queues rps_cpus
    for q in /sys/class/net/$iface/queues/rx-*; do
      if [[ -f $q/rps_cpus ]]; then
        run bash -c "echo $RPS_CPU_MASK > $q/rps_cpus" || true
      fi
    done
    # set all tx queues xps_cpus
    for q in /sys/class/net/$iface/queues/tx-*; do
      if [[ -f $q/xps_cpus ]]; then
        run bash -c "echo $RPS_CPU_MASK > $q/xps_cpus" || true
      fi
    done
  done
  echo "[INFO] RPS/XPS applied. Check /sys/class/net/<iface>/queues/"
}

# -------------------------
# Function: systemd drop-in for CPU affinity & realtime
# -------------------------
apply_systemd_dropin() {
  echo "[STEP] Creating systemd drop-in to pin $SERVICE_NAME to CPU${ISOLATED_CPU}"
  dropin_dir="/etc/systemd/system/${SERVICE_NAME}.d"
  dropin_file="${dropin_dir}/cpu.conf"
  run mkdir -p "$dropin_dir"
  cat <<EOF > "$dropin_file".tmp
[Service]
# Pin the service (MainPID) to a single CPU
CPUAffinity=${ISOLATED_CPU}
# Try to give higher scheduling priority (fallback to Nice if CPU Scheduling not available)
CPUSchedulingPolicy=fifo
CPUSchedulingPriority=50
# Lower nice to reduce interference
Nice=-10
# Restart policy
Restart=on-failure
RestartSec=2
EOF
  run mv "$dropin_file".tmp "$dropin_file"
  run systemctl daemon-reload
  echo "[INFO] systemd drop-in created at $dropin_file"
}

# -------------------------
# Function: suggest/update GRUB for isolcpus
# -------------------------
update_grub_isolation() {
  echo "[STEP] Configure GRUB to add isolcpus=nohz_full=rcu_nocbs (requires reboot)"
  # only suggest or write to /etc/default/grub depending on dry-run
  GRUB_FILE="/etc/default/grub"
  if [[ ! -f $GRUB_FILE ]]; then
    echo "[ERROR] $GRUB_FILE not found; cannot update GRUB"
    return 1
  fi
  current_cmdline=$(grep '^GRUB_CMDLINE_LINUX' "$GRUB_FILE" || true)
  echo "[INFO] Current GRUB_CMDLINE_LINUX: $current_cmdline"
  # Build expected params
  params="isolcpus=${ISOLATED_CPU} nohz_full=${ISOLATED_CPU} rcu_nocbs=${ISOLATED_CPU}"
  if grep -q "$params" "$GRUB_FILE"; then
    echo "[INFO] GRUB already appears to contain isolation params"
    return 0
  fi

  # backup
  run cp -n "$GRUB_FILE" "${GRUB_FILE}.bak.fix_jitter" || true

  if [[ $DRY -eq 1 ]]; then
    echo "[DRY] Would append $params to GRUB_CMDLINE_LINUX in $GRUB_FILE"
  else
    # insert params into GRUB_CMDLINE_LINUX line
    # safe append within quotes
    tmpfile="${GRUB_FILE}.tmp"
    awk -v p="$params" '
      BEGIN{done=0}
      /^GRUB_CMDLINE_LINUX=/ {
        line=$0
        sub(/^GRUB_CMDLINE_LINUX=/,"",line)
        # remove surrounding quotes
        gsub(/^"/,"",line); gsub(/"$/,"",line)
        if (index(line,p)==0) {
          line = line " " p
        }
        print "GRUB_CMDLINE_LINUX=\"" line "\""
        done=1
        next
      }
      {print}
      END{
        if(done==0) {
          print "GRUB_CMDLINE_LINUX=\"" p "\""
        }
      }
    ' "$GRUB_FILE" > "$tmpfile"
    run mv "$tmpfile" "$GRUB_FILE"
    echo "[INFO] GRUB updated. To apply changes, run: sudo update-grub && reboot"
    if [[ $DO_REBOOT -eq 1 ]]; then
      echo "[INFO] Reboot requested; rebooting now..."
      run update-grub
      run reboot
    fi
  fi
}

# -------------------------
# Function: undo changes (best-effort)
# -------------------------
undo_changes() {
  echo "[STEP] Undoing changes (best-effort)."
  # Attempt to re-enable irqbalance if backup exists
  if [[ -f /etc/default/irqbalance.bak ]]; then
    run mv /etc/default/irqbalance.bak /etc/default/irqbalance || true
  fi
  # remove systemd drop-in
  dropin_dir="/etc/systemd/system/${SERVICE_NAME}.d"
  if [[ -d "$dropin_dir" ]]; then
    echo "[INFO] Removing drop-in $dropin_dir"
    run rm -rf "$dropin_dir" || true
    run systemctl daemon-reload || true
  fi
  echo "[INFO] Manual rollback may be required for IRQ masks in /proc/irq/*"
}

# -------------------------
# Execute selected actions
# -------------------------
if [[ $DO_UNDO -eq 1 ]]; then
  undo_changes
  echo "[DONE] Undo attempted. Re-run diagnostics."
  exit 0
fi

if [[ $DO_FIX_ALL -eq 1 ]]; then
  DO_FIX_IRQ=1
  DO_FIX_ISOLATION=1
fi

if [[ $DO_FIX_IRQ -eq 1 ]]; then
  disable_irqbalance
  fix_irq_affinity
  fix_rps_xps
  apply_systemd_dropin
fi

if [[ $DO_FIX_ISOLATION -eq 1 ]]; then
  update_grub_isolation
fi

echo
echo "=== VERIFY: quick checks ==="
echo "- /proc/interrupts snapshot (top 10 lines):"
if [[ $DRY -eq 1 ]]; then
  echo "[DRY] cat /proc/interrupts | head -n 20"
else
  cat /proc/interrupts | head -n 20
fi

echo
echo "- systemd status for $SERVICE_NAME (show CPU Affinity):"
if [[ $DRY -eq 1 ]]; then
  echo "[DRY] systemctl show $SERVICE_NAME --property=CPUAffinity,CPUSchedulingPolicy,ExecStart"
else
  systemctl show "$SERVICE_NAME" --property=CPUAffinity,CPUSchedulingPolicy,ExecStart || true
fi

echo
echo "[DONE] Actions complete. Re-run ./scripts/diagnose.sh to validate improvements."
if [[ $DO_FIX_ISOLATION -eq 1 ]]; then
  echo "Note: GRUB changes require update-grub + reboot to take effect."
  if [[ $DO_REBOOT -eq 1 ]]; then
    echo "You asked for reboot; system should be rebooting."
  else
    echo "Run: sudo update-grub && sudo reboot"
  fi
fi

exit 0
