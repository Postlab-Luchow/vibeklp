#!/bin/bash
# Test script for KLP project

set -e

echo "=== Kulturelle Landpartie Test Suite ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse command line options
VERBOSE=false
COVERAGE=false
PACKAGE=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -p|--package)
            PACKAGE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: ./test.sh [options]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose    Run tests in verbose mode"
            echo "  -c, --coverage   Generate coverage report"
            echo "  -p, --package    Run tests for specific package (e.g., ./internal/api)"
            echo "  -h, --help       Show this help message"
            echo ""
            echo "Examples:"
            echo "  ./test.sh                    # Run all tests"
            echo "  ./test.sh -v                 # Run with verbose output"
            echo "  ./test.sh -c                 # Run with coverage"
            echo "  ./test.sh -p ./internal/api  # Test specific package"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Determine what to test
if [ -n "$PACKAGE" ]; then
    TEST_TARGET="$PACKAGE"
else
    TEST_TARGET="./..."
fi

# Build test command
TEST_CMD="go test $TEST_TARGET"

if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$COVERAGE" = true ]; then
    TEST_CMD="$TEST_CMD -cover -coverprofile=coverage.out"
fi

# Run tests
echo "Running tests..."
echo "Command: $TEST_CMD"
echo ""

if eval $TEST_CMD; then
    echo ""
    echo -e "${GREEN}✅ All tests passed!${NC}"
    
    # Show coverage report if requested
    if [ "$COVERAGE" = true ]; then
        echo ""
        echo "=== Coverage Report ==="
        go tool cover -func=coverage.out
        echo ""
        echo -e "${YELLOW}📊 To view detailed coverage HTML report:${NC}"
        echo "   go tool cover -html=coverage.out"
    fi
    
    exit 0
else
    echo ""
    echo -e "${RED}❌ Tests failed!${NC}"
    exit 1
fi
