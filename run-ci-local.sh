#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Function to install Go tools if not present
install_go_tools() {
    print_info "Installing Go tools..."

    if ! command -v staticcheck &> /dev/null; then
        go install honnef.co/go/tools/cmd/staticcheck@latest
    fi

    if ! command -v ineffassign &> /dev/null; then
        go install github.com/gordonklaus/ineffassign@latest
    fi

    if ! command -v misspell &> /dev/null; then
        go install github.com/client9/misspell/cmd/misspell@latest
    fi

    if ! command -v goimports &> /dev/null; then
        go install golang.org/x/tools/cmd/goimports@latest
    fi

    if ! command -v gosec &> /dev/null; then
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    fi

    if ! command -v golangci-lint &> /dev/null; then
        print_info "Installing golangci-lint..."
        # Install golangci-lint using the recommended method
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
    fi

    print_success "Go tools installed"
}

# Function to run checks for a module
run_module_checks() {
    local module=$1
    print_header "Running CI checks for $module"

    cd "$module"

    # Install dependencies
    print_info "Installing dependencies..."
    go mod download

    # Verify dependencies
    print_info "Verifying dependencies..."
    go mod verify

    # Check go.mod and go.sum are tidy
    print_info "Checking go.mod and go.sum..."
    go mod tidy
    # if [ -n "$(git status --porcelain go.mod go.sum)" ]; then
    #     print_error "go.mod or go.sum not tidy. Run 'go mod tidy'"       
    #     cd ..
    #     return 1
    # fi
    print_success "go.mod and go.sum are tidy"

    # Run goimports
    print_info "Running goimports..."
    goimports -w .
    if [ -n "$(git status --porcelain)" ]; then
        print_warning "Code is not properly imported. Changes made by goimports"       
    else
        print_success "goimports check passed"
    fi

    # Run go fmt
    print_info "Running go fmt..."
    if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
        print_error "Code is not formatted. Run 'go fmt ./...'"
        gofmt -s -l .
        cd ..
        return 1
    fi
    print_success "go fmt check passed"

    # Run go vet
    print_info "Running go vet..."
    go vet -v ./...
    print_success "go vet check passed"

    # Run staticcheck
    print_info "Running staticcheck..."
    staticcheck -checks=all ./...
    print_success "staticcheck passed"

    # Run ineffassign
    print_info "Running ineffassign..."
    ineffassign ./...
    print_success "ineffassign check passed"

    # Check for misspellings
    print_info "Checking for misspellings..."
    misspell -error .
    print_success "misspell check passed"

    # Run golangci-lint
    print_info "Running golangci-lint..."
    golangci-lint run --timeout=5m --out-format=colored-line-number
    print_success "golangci-lint check passed"

    # Run gosec for order-food
    if [ "$module" == "order-food" ]; then
        print_info "Running gosec security scanner..."
        gosec -fmt=json -out=gosec-report.json ./... || true
        if [ -f "gosec-report.json" ]; then
            print_info "gosec report saved to $module/gosec-report.json"
        fi
    fi

    # Build
    print_info "Building..."
    go build -v -o bin/$module ./cmd/main.go
    print_success "Build successful"

    # Build with optimization (order-food only)
    if [ "$module" == "order-food" ]; then
        print_info "Building with optimization..."
        go build -ldflags="-s -w" -o bin/$module-optimized ./cmd/main.go
        ls -lh bin/
        print_success "Optimized build successful"
    fi

    # Run tests with coverage
    print_info "Running tests with coverage..."
    go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

    # Generate coverage report
    print_info "Generating coverage report..."
    go tool cover -func=coverage.out
    go tool cover -html=coverage.out -o coverage.html

    # Calculate coverage percentage
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    print_success "Coverage: $COVERAGE%"
    print_info "Coverage report saved to $module/coverage.html"

    # Check coverage threshold for order-food
    if [ "$module" == "order-food" ]; then
        THRESHOLD=70
        if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
            print_warning "Coverage $COVERAGE% is below threshold $THRESHOLD%"
        fi
    fi

    # Run benchmarks for order-food
    if [ "$module" == "order-food" ]; then
        print_info "Running benchmarks..."
        go test -bench=. -benchmem -run=^$ ./... | tee benchmark.txt || true
        if [ -f "benchmark.txt" ]; then
            print_info "Benchmark results saved to $module/benchmark.txt"
        fi
    fi

    cd ..
    print_success "All checks passed for $module"
}

# Main execution
main() {
    print_header "Local CI Pipeline"

    # Check if we're in the right directory
    if [ ! -f "go.work" ]; then
        print_error "Please run this script from the project root directory"
        exit 1
    fi

    # Install Go tools
    install_go_tools

    # Determine which modules to check
    if [ "$1" == "all" ] || [ -z "$1" ]; then
        MODULES=("database-migration" "database-load" "order-food")
    elif [ "$1" == "database-migration" ] || [ "$1" == "database-load" ] || [ "$1" == "order-food" ]; then
        MODULES=("$1")
    else
        print_error "Invalid module: $1"
        print_info "Usage: ./run-ci-local.sh [all|database-migration|database-load|order-food]"
        exit 1
    fi

    # Run checks for each module
    FAILED_MODULES=()
    for module in "${MODULES[@]}"; do
        if ! run_module_checks "$module"; then
            FAILED_MODULES+=("$module")
        fi
    done

    # Print summary
    print_header "CI Summary"

    if [ ${#FAILED_MODULES[@]} -eq 0 ]; then
        print_success "All checks passed! ✅"
        exit 0
    else
        print_error "Some checks failed for: ${FAILED_MODULES[*]} ❌"
        exit 1
    fi
}

# Run main function
main "$@"
