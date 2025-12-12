# Ultra-Low-Latency 1µs Anti-Nuke Engine

Military-grade Discord anti-nuke system with sub-microsecond detection and ~200ms response time.

## Architecture

- **1µs Detection Core**: L1-resident counters, branchless operations, zero allocations
- **SPSC Ring Buffers**: Lock-free event pipeline
- **CPU Isolation**: Dedicated cores for correlator, ingestion, decision, dispatch
- **GC-Free Hot Path**: Disabled garbage collection for predictable latency
- **Memory Locked**: Prevents page faults in critical paths

## Features

- Ban spike detection
- Channel/Role deletion detection
- Permission escalation detection (XOR bitmask)
- Multi-actor coordinated attack detection
- Velocity-based anomaly detection
- Auto-ban with configurable thresholds
- Emergency lockdown modes
- Forensic audit log reconciliation
- Snapshot & rollback capabilities

## Quick Start

```bash
# Set your Discord bot token
export DISCORD_TOKEN="your_token_here"

# Build and run
chmod +x run.sh
./run.sh
```

## Configuration

Edit `config.json` to customize:
- Detection thresholds by guild size
- CPU core assignments
- Safety modes (normal, elevated, high, lockdown, emergency)
- HTTP pool size and worker count
- Forensics retention

## Performance Targets

- **Detection Latency**: ≤1µs internal processing
- **Ban Execution**: ~200ms end-to-end (network dependent)
- **Throughput**: 100K+ events/second
- **Memory**: Preallocated, locked, cache-aligned

## System Requirements

- Go 1.23+
- Linux kernel 5.10+ (for optimal CPU isolation)
- Multi-core CPU (recommend 8+ cores)
- Sufficient privileges for memory locking

## License

Proprietary - Internal Use Only
