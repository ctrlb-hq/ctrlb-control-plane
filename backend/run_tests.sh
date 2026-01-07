#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test packages (directories containing _test.go files)
PACKAGES=(
    "./internal/agent/..."
    "./internal/api/..."
    "./internal/auth/..."
    "./internal/db/..."
    "./internal/frontend/agent/..."
    "./internal/frontend/node/..."
    "./internal/frontend/pipeline/..."
    "./internal/middleware/..."
    "./internal/pkg/configcompiler/..."
    "./internal/pkg/queue/..."
    "./internal/utils/..."
)

TOTAL_PASSED=0
TOTAL_FAILED=0

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}    Backend Go Test Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

for pkg in "${PACKAGES[@]}"; do
    echo -e "${YELLOW}Testing: ${pkg}${NC}"
    echo "----------------------------------------"
    
    # Run tests and capture output
    OUTPUT=$(go test -v -cover "$pkg" 2>&1)
    EXIT_CODE=$?
    
    # Count passed and failed tests
    PASSED=$(echo "$OUTPUT" | grep -c "^--- PASS:")
    FAILED=$(echo "$OUTPUT" | grep -c "^--- FAIL:")
    
    # Extract coverage
    COVERAGE=$(echo "$OUTPUT" | grep -oE "coverage: [0-9]+\.[0-9]+%" | tail -1)
    if [ -z "$COVERAGE" ]; then
        COVERAGE="no test files or no coverage"
    fi
    
    # Update totals
    TOTAL_PASSED=$((TOTAL_PASSED + PASSED))
    TOTAL_FAILED=$((TOTAL_FAILED + FAILED))
    
    # Display results
    if [ $EXIT_CODE -eq 0 ]; then
        echo -e "  ${GREEN}✓ PASSED${NC}: $PASSED tests"
    else
        echo -e "  ${RED}✗ FAILED${NC}: $FAILED tests"
        echo -e "  ${GREEN}✓ PASSED${NC}: $PASSED tests"
    fi
    echo -e "  ${BLUE}Coverage${NC}: $COVERAGE"
    echo ""
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}    SUMMARY${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "  ${GREEN}Total Passed${NC}: $TOTAL_PASSED"
echo -e "  ${RED}Total Failed${NC}: $TOTAL_FAILED"

if [ $TOTAL_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed! ✗${NC}"
    exit 1
fi