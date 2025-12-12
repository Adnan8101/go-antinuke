/**
 * PM2 Ecosystem Configuration for Ultra-Low-Latency Antinuke Bot
 * 
 * This configuration enables:
 * - CPU affinity management
 * - Performance-optimized Node.js flags
 * - Automatic restart policies
 * - Persistent logging
 * - Resource limits
 */

module.exports = {
  apps: [{
    name: 'antinuke-bot',
    
    // Use compiled binary (Go application)
    script: './bin/antinuke',
    
    // Alternative for Node.js wrapper if needed
    // script: './cmd/main.go',
    // interpreter: 'go',
    // interpreter_args: 'run',
    
    // Working directory
    cwd: '/opt/antinuke',
    
    // Instance configuration
    instances: 1,  // Single instance for CPU pinning
    exec_mode: 'fork',  // Fork mode for Go binary
    
    // Auto-restart configuration
    autorestart: true,
    watch: false,  // Disable watch in production
    max_restarts: 10,
    min_uptime: '10s',
    restart_delay: 2000,
    
    // Resource limits
    max_memory_restart: '2G',
    
    // Environment variables
    env: {
      NODE_ENV: 'production',
      GO_ENV: 'production',
      GOMAXPROCS: '1',  // Use single CPU core
      GOGC: '50',  // More aggressive GC for consistent latency
      GODEBUG: 'gctrace=0',
    },
    
    // Logging configuration
    error_file: '/var/log/antinuke/error.log',
    out_file: '/var/log/antinuke/output.log',
    log_file: '/var/log/antinuke/combined.log',
    log_date_format: 'YYYY-MM-DD HH:mm:ss.SSS Z',
    merge_logs: true,
    
    // Log rotation
    log_type: 'json',
    
    // Process priority
    nice: -20,  // Highest priority (requires root)
    
    // CPU affinity (will be set by external script)
    // This is a placeholder - actual affinity set via taskset
    
    // Health check
    health_check: {
      enabled: true,
      interval: 30000,  // 30 seconds
      timeout: 5000,
    },
    
    // Advanced PM2 options
    kill_timeout: 5000,
    listen_timeout: 10000,
    shutdown_with_message: false,
    
    // Post-deploy hooks
    post_start: './scripts/post_start_hook.sh',
    pre_stop: './scripts/pre_stop_hook.sh',
  }],
  
  // Deployment configuration (optional)
  deploy: {
    production: {
      user: 'antinuke',
      host: 'localhost',
      ref: 'origin/main',
      repo: 'git@github.com:your-repo/antinuke.git',
      path: '/opt/antinuke',
      'pre-deploy': 'git fetch --all',
      'post-deploy': './scripts/deploy.sh',
      'pre-setup': './scripts/vm_optimize.sh',
    }
  }
};
