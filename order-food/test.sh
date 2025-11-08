#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}===== Running Order Food Tests =====${NC}\n"

# Get to the order-food directory
cd "$(dirname "$0")"

# Run tests with coverage
echo -e "${BLUE}Running tests with coverage...${NC}"
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Check if tests passed
if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✓ All tests passed!${NC}\n"
else
    echo -e "\n${RED}✗ Tests failed${NC}\n"
    exit 1
fi

# Display coverage report
echo -e "${BLUE}Coverage Report:${NC}"
go tool cover -func=coverage.out

# Calculate total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
echo -e "\n${BLUE}Total Coverage: ${GREEN}${TOTAL_COVERAGE}${NC}\n"

# Generate HTML coverage report
echo -e "${BLUE}Generating HTML coverage report...${NC}"
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}✓ HTML coverage report generated: coverage.html${NC}\n"

# Check if coverage meets threshold (optional)
COVERAGE_THRESHOLD=70
COVERAGE_NUM=$(echo $TOTAL_COVERAGE | sed 's/%//')
COVERAGE_INT=${COVERAGE_NUM%.*}

if [ "$COVERAGE_INT" -ge "$COVERAGE_THRESHOLD" ]; then
    echo -e "${GREEN}✓ Coverage threshold met ($COVERAGE_INT% >= $COVERAGE_THRESHOLD%)${NC}"
else
    echo -e "${YELLOW}⚠ Coverage below threshold ($COVERAGE_INT% < $COVERAGE_THRESHOLD%)${NC}"
fi

echo -e "\n${BLUE}===== Test Summary =====${NC}"
echo -e "Tests: ${GREEN}PASSED${NC}"
echo -e "Coverage: ${GREEN}${TOTAL_COVERAGE}${NC}"
echo -e "Report: ${BLUE}coverage.html${NC}\n"
