#!/bin/bash

# Load Test Script for Bifrost
# Runs a load test against bifrost-http with a mocker provider
# Usage: ./load-test.sh
#
# This script:
# 1. Builds bifrost-http locally if not present
# 2. Creates a config.json with mocker provider (OpenAI-style)
# 3. Starts the mocker server and bifrost-http
# 4. Runs a load test at the configured RPS for 1 minute using Vegeta
# 5. Prints the results

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"
BIFROST_HTTP_DIR="${REPO_ROOT}/transports/bifrost-http"
TRANSPORTS_DIR="${REPO_ROOT}/transports"
WORK_DIR="${SCRIPT_DIR}"
MOCKER_DIR="${REPO_ROOT}/../bifrost-benchmarking/mocker"

BIFROST_PORT=8080
MOCKER_PORT=8000
RATE=1000
MAX_WORKERS=1000
DURATION=60
MOCKER_LATENCY_MS=10000  # 10 seconds base latency from mocker
MAX_LATENCY_US=10000100  # 10 seconds + 100us base latency from mocker
VEGETA_TO_BIFROST_DELAY_US=100
BIFROST_TO_MOCKER_DELAY_US=100

# Results storage for summary table
RESULTS_FILE="${WORK_DIR}/load-test-results.md"
RESULTS_JSON="${WORK_DIR}/load-test-results.json"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
  echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function to kill background processes
cleanup() {
  log_info "Cleaning up..."
  if [ -n "$BIFROST_PID" ] && kill -0 "$BIFROST_PID" 2>/dev/null; then
    kill "$BIFROST_PID" 2>/dev/null || true
    wait "$BIFROST_PID" 2>/dev/null || true
  fi
  if [ -n "$MOCKER_PID" ] && kill -0 "$MOCKER_PID" 2>/dev/null; then
    kill "$MOCKER_PID" 2>/dev/null || true
    wait "$MOCKER_PID" 2>/dev/null || true
  fi
  # Clean up temporary files (keep results files for artifact upload)
  rm -f "${WORK_DIR}/config.json" "${WORK_DIR}/logs.db" "${WORK_DIR}/attack.bin" "${WORK_DIR}/bifrost.log" "${WORK_DIR}/vegeta-target.json" "${WORK_DIR}/vegeta-report.json" 2>/dev/null || true
  log_info "Cleanup complete"
}

trap cleanup EXIT

# Check for required tools
check_dependencies() {
  log_info "Checking dependencies..."

  if ! command -v go &> /dev/null; then
    log_error "Go is not installed. Please install Go 1.24.3 or later."
    exit 1
  fi

  if ! command -v git &> /dev/null; then
    log_error "Git is not installed. Please install Git."
    exit 1
  fi

  log_success "All dependencies found"
}

# Kill any process running on a specific port
kill_port() {
  local port=$1
  local pids=$(lsof -ti ":${port}" 2>/dev/null)
  if [ -n "$pids" ]; then
    log_warn "Killing existing process(es) on port ${port}: ${pids}"
    echo "$pids" | xargs kill -9 2>/dev/null || true
    sleep 1
  fi
}

# Kill processes on required ports before starting
cleanup_ports() {
  log_info "Checking for processes on required ports..."
  kill_port ${MOCKER_PORT}
  kill_port ${BIFROST_PORT}
}

# Install Vegeta if not present
install_vegeta() {
  if ! command -v vegeta &> /dev/null; then
    log_info "Installing Vegeta load testing tool..."
    go install github.com/tsenart/vegeta/v12@latest
    export PATH="$PATH:$(go env GOPATH)/bin"
    if ! command -v vegeta &> /dev/null; then
      log_error "Failed to install Vegeta"
      exit 1
    fi
    log_success "Vegeta installed"
  else
    log_success "Vegeta already installed"
  fi
}

# Build bifrost-http if binary doesn't exist
build_bifrost_http() {
  if [ -f "${REPO_ROOT}/tmp/bifrost-http" ]; then
    log_success "bifrost-http binary already exists at ${REPO_ROOT}/tmp/bifrost-http"
    return 0
  fi

  log_info "Building bifrost-http..."
  cd "${TRANSPORTS_DIR}"

  # Build the binary
  if go build -o ${REPO_ROOT}/tmp/bifrost-http .; then
    log_success "bifrost-http built successfully"
  else
    log_error "Failed to build bifrost-http"
    exit 1
  fi

  cd "${WORK_DIR}"
}

# Clone and setup mocker from bifrost-benchmarking
setup_mocker() {
  if [ -d "${REPO_ROOT}/../bifrost-benchmarking" ]; then
    log_info "Updating bifrost-benchmarking repository..."
    cd "${REPO_ROOT}/../bifrost-benchmarking"
    git pull --quiet || true
    cd "${WORK_DIR}"
  else
    log_info "Cloning bifrost-benchmarking repository..."
    cd "${WORK_DIR}"
    git clone --depth 1 https://github.com/maximhq/bifrost-benchmarking.git
  fi

  log_success "Mocker setup complete"
}

# Create config.json for bifrost with mocker provider
create_config() {
  log_info "Creating config.json..."

  cat > "${WORK_DIR}/config.json" << 'EOF'
{
  "$schema": "https://www.getbifrost.ai/schema",
  "client": {
    "enable_logging": false,
    "initial_pool_size": 10000,
    "drop_excess_requests": false,
    "enable_governance": false,
    "allow_direct_keys": false
  },
  "config_store": {
    "enabled": false
  },
  "logs_store": {
    "enabled": false
  },
  "providers": {
    "openai": {
      "keys": [
        {
          "name": "mocker-key",
          "value": "Bearer mocker-key",
          "weight": 1
        }
      ],
      "network_config": {
        "base_url": "http://localhost:8000",
        "default_request_timeout_in_seconds": 30
      },
      "concurrency_and_buffer_size": {
        "concurrency": 2000,
        "buffer_size": 2500
      },
      "custom_provider_config": {
        "base_provider_type": "openai",
        "allowed_requests": {
          "list_models": false,
          "chat_completion": true,
          "chat_completion_stream": true
        }
      }
    }
  }
}
EOF

  log_success "config.json created"
}

# Start the mocker server
start_mocker() {
  log_info "Starting mocker server on port ${MOCKER_PORT}..."

  cd "${MOCKER_DIR}"
  go run main.go -port ${MOCKER_PORT} -host 0.0.0.0 -latency ${MOCKER_LATENCY_MS}&
  MOCKER_PID=$!
  cd "${WORK_DIR}"

  # Wait for mocker to be ready
  local max_attempts=30
  local attempt=0
  while ! curl -s "http://localhost:${MOCKER_PORT}/v1/chat/completions" -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer mocker-key" \
    -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"test"}]}' > /dev/null 2>&1; do
    sleep 1
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
      log_error "Mocker failed to start within ${max_attempts} seconds"
      exit 1
    fi
  done

  log_success "Mocker server started (PID: ${MOCKER_PID})"
}

# Start bifrost-http server
start_bifrost() {
  log_info "Starting bifrost-http on port ${BIFROST_PORT}..."

  cd "${WORK_DIR}"
  local bifrost_log="${WORK_DIR}/bifrost.log"
  "${REPO_ROOT}/tmp/bifrost-http" -app-dir "${WORK_DIR}" -port "${BIFROST_PORT}" -host "0.0.0.0" -log-level "info" > "${bifrost_log}" 2>&1 &
  BIFROST_PID=$!

  # Wait for bifrost to be fully ready (look for "successfully started bifrost" message)
  local max_attempts=60
  local attempt=0
  while ! grep -q "successfully started bifrost" "${bifrost_log}" 2>/dev/null; do
    sleep 1
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
      log_error "Bifrost failed to start within ${max_attempts} seconds"
      log_error "Bifrost log output:"
      cat "${bifrost_log}" 2>/dev/null || true
      exit 1
    fi
    # Check if process is still running
    if ! kill -0 "$BIFROST_PID" 2>/dev/null; then
      log_error "Bifrost process died unexpectedly"
      log_error "Bifrost log output:"
      cat "${bifrost_log}" 2>/dev/null || true
      exit 1
    fi
  done

  log_success "Bifrost-http started (PID: ${BIFROST_PID})"
}

# Initialize results file with header
init_results_file() {
  cat > "${RESULTS_FILE}" << 'EOF'
# Bifrost Load Test Results

## Latency Summary

| Scenario | RPS | Duration | Success Rate | Min | Mean | P50 | P90 | P95 | P99 | Max |
|----------|-----|----------|--------------|-----|------|-----|-----|-----|-----|-----|
EOF

  # Initialize JSON results
  echo '{"scenarios": [], "timestamp": "'"$(date -u +"%Y-%m-%dT%H:%M:%SZ")"'"}' > "${RESULTS_JSON}"
}

# Parse vegeta output and record raw latencies
# Arguments: $1 = scenario name, $2 = rate, $3 = duration
parse_and_record_results() {
  local scenario_name=$1
  local rate=$2
  local duration=$3

  # Get JSON report from vegeta
  local json_report_file="${WORK_DIR}/vegeta-report.json"
  vegeta report -type=json < "${WORK_DIR}/attack.bin" > "${json_report_file}"

  log_info "Parsing vegeta JSON report..."

  # Check if jq is available for reliable JSON parsing
  if command -v jq &> /dev/null; then
    local latency_min=$(jq '.latencies.min // 0' "${json_report_file}")
    local latency_mean=$(jq '.latencies.mean // 0' "${json_report_file}")
    local latency_50=$(jq '.latencies["50th"] // 0' "${json_report_file}")
    local latency_90=$(jq '.latencies["90th"] // 0' "${json_report_file}")
    local latency_95=$(jq '.latencies["95th"] // 0' "${json_report_file}")
    local latency_99=$(jq '.latencies["99th"] // 0' "${json_report_file}")
    local latency_max=$(jq '.latencies.max // 0' "${json_report_file}")
    local success_rate=$(jq '.success // 0' "${json_report_file}")
  elif command -v python3 &> /dev/null; then
    local latency_min=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('latencies', {}).get('min', 0))")
    local latency_mean=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('latencies', {}).get('mean', 0))")
    local latency_50=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('latencies', {}).get('50th', 0))")
    local latency_90=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('latencies', {}).get('90th', 0))")
    local latency_95=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('latencies', {}).get('95th', 0))")
    local latency_99=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('latencies', {}).get('99th', 0))")
    local latency_max=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('latencies', {}).get('max', 0))")
    local success_rate=$(python3 -c "import json; d=json.load(open('${json_report_file}')); print(d.get('success', 0))")
  else
    log_error "Neither jq nor python3 found. Cannot parse JSON results."
    return 1
  fi

  log_info "  Extracted latencies (ns): min=$latency_min, mean=$latency_mean, p50=$latency_50, p99=$latency_99, max=$latency_max"
  log_info "  Success rate: $success_rate"

  if [ -z "$latency_min" ] || [ "$latency_min" = "0" ] || [ "$latency_min" = "null" ]; then
    log_warn "Failed to extract latency values from vegeta report"
    log_info "Vegeta JSON report contents:"
    cat "${json_report_file}"
    return 1
  fi

  # Subtract network delay (vegeta->bifrost + bifrost->mocker) before converting to µs
  local delay_ns=$(( (VEGETA_TO_BIFROST_DELAY_US + BIFROST_TO_MOCKER_DELAY_US) * 1000 ))
  local us_min=$(printf "%.2f" $(echo "scale=4; ($latency_min - $delay_ns) / 1000" | bc))
  local us_mean=$(printf "%.2f" $(echo "scale=4; ($latency_mean - $delay_ns) / 1000" | bc))
  local us_50=$(printf "%.2f" $(echo "scale=4; ($latency_50 - $delay_ns) / 1000" | bc))
  local us_90=$(printf "%.2f" $(echo "scale=4; ($latency_90 - $delay_ns) / 1000" | bc))
  local us_95=$(printf "%.2f" $(echo "scale=4; ($latency_95 - $delay_ns) / 1000" | bc))
  local us_99=$(printf "%.2f" $(echo "scale=4; ($latency_99 - $delay_ns) / 1000" | bc))
  local us_max=$(printf "%.2f" $(echo "scale=4; ($latency_max - $delay_ns) / 1000" | bc))

  # Convert success rate to percentage
  local success_pct=$(printf "%.2f" $(echo "scale=4; $success_rate * 100" | bc))

  # Append to markdown table
  echo "| ${scenario_name} | ${rate} | ${duration}s | ${success_pct}% | ${us_min}µs | ${us_mean}µs | ${us_50}µs | ${us_90}µs | ${us_95}µs | ${us_99}µs | ${us_max}µs |" >> "${RESULTS_FILE}"

  # Update JSON results
  local tmp_json=$(mktemp)
  cat "${RESULTS_JSON}" | sed 's/\("scenarios": \[\)/\1{"name": "'"${scenario_name}"'", "rate": '"${rate}"', "duration": '"${duration}"', "success_rate": '"${success_pct}"', "latency_us": {"min": '"${us_min}"', "mean": '"${us_mean}"', "p50": '"${us_50}"', "p90": '"${us_90}"', "p95": '"${us_95}"', "p99": '"${us_99}"', "max": '"${us_max}"'}},/' > "${tmp_json}"
  mv "${tmp_json}" "${RESULTS_JSON}"

  rm -f "${json_report_file}"

  log_success "Results recorded for scenario: ${scenario_name}"
  log_info "  Latency - Min: ${us_min}µs, Mean: ${us_mean}µs, P99: ${us_99}µs, Max: ${us_max}µs"

  # Check if any adjusted latency exceeds the threshold
  local failed=0
  local delay_us=$(( VEGETA_TO_BIFROST_DELAY_US + BIFROST_TO_MOCKER_DELAY_US ))
  local labels=("Min" "Mean" "P50" "P90" "P95" "P99" "Max")
  local values=($latency_min $latency_mean $latency_50 $latency_90 $latency_95 $latency_99 $latency_max)
  local extras=()

  for i in "${!values[@]}"; do
    local adjusted_us=$(( values[i] / 1000 - delay_us ))
    if [ "$adjusted_us" -gt "$MAX_LATENCY_US" ]; then
      extras+=("${labels[i]}:$(( adjusted_us - MAX_LATENCY_US ))")
      failed=1
    fi
  done

  if [ "$failed" -eq 1 ]; then
    echo ""
    log_error "FAILED: Latency exceeded ${MAX_LATENCY_US}µs threshold"
    echo ""
    echo -e "${RED}| Bucket | Extra Latency (µs) |${NC}"
    echo -e "${RED}|--------|---------------------|${NC}"
    for entry in "${extras[@]}"; do
      local bucket="${entry%%:*}"
      local extra="${entry##*:}"
      echo -e "${RED}| ${bucket} | +${extra}µs |${NC}"
    done
    echo ""
    exit 1
  fi
}

# Finalize results file
finalize_results() {
  cat >> "${RESULTS_FILE}" << 'EOF'

## Notes

- All latency values are in microseconds (µs)
- P50/P90/P95/P99 represent percentile latencies

---
*Generated by Bifrost Load Test Script*
EOF

  # Fix JSON (remove trailing comma in scenarios array)
  sed -i.bak 's/},\]/}]/' "${RESULTS_JSON}" 2>/dev/null || sed -i '' 's/},\]/}]/' "${RESULTS_JSON}"
  rm -f "${RESULTS_JSON}.bak"

  log_success "Results saved to:"
  log_info "  - Markdown: ${RESULTS_FILE}"
  log_info "  - JSON: ${RESULTS_JSON}"

  echo ""
  echo "╔═══════════════════════════════════════════════════════════╗"
  echo "║              Latency Summary Table                        ║"
  echo "╚═══════════════════════════════════════════════════════════╝"
  echo ""
  cat "${RESULTS_FILE}"
}

# Run a single load test scenario
# Arguments: $1 = scenario name, $2 = rate (optional), $3 = duration (optional)
run_load_test() {
  local scenario_name=${1:-"Default"}
  local rate=${2:-$RATE}
  local duration=${3:-$DURATION}

  log_info "Running load test scenario '${scenario_name}': ${rate} RPS for ${duration} seconds..."
  echo ""

  # Create the target file for Vegeta
  local target_url="http://localhost:${BIFROST_PORT}/v1/chat/completions"
  local target_file="${WORK_DIR}/vegeta-target.json"
  local payload='{"model":"openai/gpt-4o-mini","messages":[{"role":"user","content":"Hello, how are you?"}]}'

  # Write target in Vegeta JSON format
  cat > "${target_file}" << EOF
{"method": "POST", "url": "${target_url}", "header": {"Content-Type": ["application/json"]}, "body": "$(echo -n "${payload}" | base64)"}
EOF

  # Run Vegeta attack and save binary results to file
  vegeta attack \
    -format=json \
    -targets="${target_file}" \
    -rate="${rate}" \
    -duration="${duration}s" \
    -timeout="30s" \
    -workers=500 \
    -max-workers="${MAX_WORKERS}" > "${WORK_DIR}/attack.bin"

  echo ""
  log_info "Attack complete. Generating reports..."

  echo ""
  log_info "Summary report:"
  vegeta report < "${WORK_DIR}/attack.bin"

  echo ""
  log_info "Latency histogram:"
  vegeta report -type=hist[0,1ms,5ms,10ms,50ms,100ms,500ms,1s,5s,10s,15s] < "${WORK_DIR}/attack.bin" || log_warn "Histogram generation failed"

  # Parse results and record to summary
  parse_and_record_results "${scenario_name}" "${rate}" "${duration}"
}

# Run all test scenarios
run_all_scenarios() {
  # Initialize results file
  init_results_file

  echo ""
  echo "╔═══════════════════════════════════════════════════════════╗"
  echo "║                    Load Test Results                      ║"
  echo "╚═══════════════════════════════════════════════════════════╝"
  echo ""

  run_load_test "Load Test (${RATE} RPS)" "${RATE}" "${DURATION}"

  # Finalize and display summary
  finalize_results
}

# Main execution
main() {
  echo ""
  echo "╔═══════════════════════════════════════════════════════════╗"
  echo "║              Bifrost Load Test Script                     ║"
  echo "║           ${RATE} RPS for ${DURATION} seconds                         ║"
  echo "╚═══════════════════════════════════════════════════════════╝"
  echo ""

  echo "Note: As the mocker may respond not exactly in 1000ms, there could be additional latency (+-10/20ms) between the mocker and the bifrost-http server."

  check_dependencies
  install_vegeta
  build_bifrost_http
  setup_mocker
  create_config
  cleanup_ports
  start_mocker
  start_bifrost

  run_all_scenarios
  cleanup_ports
  echo ""
  log_success "Load test completed successfully!"
  log_info "Results files location:"
  log_info "  - ${RESULTS_FILE}"
  log_info "  - ${RESULTS_JSON}"

  # Print final summary table to console
  echo ""
  echo "╔══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╗"
  echo "║                                                         FINAL RESULTS SUMMARY                                                                                    ║"
  echo "╚══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════╝"
  echo ""
  echo "Mocker Latency: ${MOCKER_LATENCY_MS}ms"
  echo ""
  # Print just the table (header + separator + data rows)
  grep -E "^\|" "${RESULTS_FILE}"
  echo ""
  echo "Notes:"
  echo "  - All latency values are in microseconds (µs)"
  echo "  - P50/P90/P95/P99 represent percentile latencies"
  echo ""
}

main "$@"
