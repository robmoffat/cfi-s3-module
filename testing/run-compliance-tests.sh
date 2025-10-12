#!/bin/bash
set -euo pipefail

# CCC CFI Compliance Test Runner
# This script discovers cloud resources and runs compliance tests against them

# Default values
PROVIDER=""
OUTPUT_DIR="output"
FEATURES_PATH="testing/features"
SKIP_PORTS=""
SKIP_SERVICES=""
TIMEOUT="30m"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -p|--provider)
      PROVIDER="$2"
      shift 2
      ;;
    -o|--output)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    -f|--features)
      FEATURES_PATH="$2"
      shift 2
      ;;
    --skip-ports)
      SKIP_PORTS="--skip-ports"
      shift
      ;;
    --skip-services)
      SKIP_SERVICES="--skip-services"
      shift
      ;;
    -t|--timeout)
      TIMEOUT="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  -p, --provider PROVIDER    Cloud provider (aws, azure, or gcp) [REQUIRED]"
      echo "  -o, --output DIR          Output directory for test reports (default: output)"
      echo "  -f, --features PATH       Path to feature files (default: testing/features)"
      echo "  --skip-ports              Skip port tests"
      echo "  --skip-services           Skip service tests"
      echo "  -t, --timeout DURATION    Timeout for all tests (default: 30m)"
      echo "  -h, --help                Show this help message"
      echo ""
      echo "Examples:"
      echo "  $0 --provider aws"
      echo "  $0 --provider azure --output results"
      echo "  $0 --provider gcp --skip-ports"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use -h or --help for usage information"
      exit 1
      ;;
  esac
done

# Validate required parameters
if [ -z "$PROVIDER" ]; then
  echo "Error: --provider is required"
  echo "Use -h or --help for usage information"
  exit 1
fi

if [ "$PROVIDER" != "aws" ] && [ "$PROVIDER" != "azure" ] && [ "$PROVIDER" != "gcp" ]; then
  echo "Error: provider must be 'aws', 'azure', or 'gcp'"
  exit 1
fi

# Check if Steampipe is running
echo "🔍 Checking Steampipe connection..."
if ! steampipe query "SELECT 1" > /dev/null 2>&1; then
  echo "❌ Error: Steampipe is not running or not accessible"
  echo "   Please start Steampipe with: steampipe service start"
  exit 1
fi
echo "✅ Steampipe is running"
echo ""

# Build the Go test runner
echo "🔨 Building test runner..."
cd "$(dirname "$0")"
go build -o run-tests ./runner/main.go
echo "✅ Test runner built"
echo ""

# Run the tests
echo "🚀 Running compliance tests..."
./run-tests \
  --provider "$PROVIDER" \
  --output "$OUTPUT_DIR" \
  --features "$FEATURES_PATH" \
  --timeout "$TIMEOUT" \
  $SKIP_PORTS \
  $SKIP_SERVICES

exit_code=$?

# Cleanup
rm -f run-tests

exit $exit_code

