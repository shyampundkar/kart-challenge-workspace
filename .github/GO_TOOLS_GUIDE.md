# Go Tools and Code Quality Guide

Comprehensive guide to all Go tools integrated into the CI pipeline.

## Overview

The CI pipeline includes multiple Go tools to ensure code quality, security, and maintainability:

| Tool | Purpose | Severity | Auto-Fix |
|------|---------|----------|----------|
| **go fmt** | Code formatting | Error | Yes |
| **goimports** | Import organization | Error | Yes |
| **go vet** | Static analysis | Error | Manual |
| **staticcheck** | Advanced linting | Error | Some |
| **ineffassign** | Unused assignments | Warning | Manual |
| **misspell** | Spelling checker | Warning | Yes |
| **gosec** | Security scanner | Warning | Manual |
| **golangci-lint** | Meta-linter | Error | Some |
| **go test -race** | Race detector | Error | Manual |
| **go test -cover** | Coverage analysis | Info | N/A |

## Tool Details

### 1. go fmt

**Purpose:** Enforces standard Go code formatting

**What it checks:**
- Indentation (tabs vs spaces)
- Line breaks
- Spacing around operators
- Braces positioning

**Usage:**
```bash
# Check formatting
go fmt ./...

# Format code in-place
gofmt -s -w .
```

**CI Integration:**
- Runs on every push/PR
- Fails if code is not formatted
- Shows which files need formatting

**Example Error:**
```
Code is not formatted. Run 'go fmt ./...':
internal/handler/product_handler.go
internal/service/order_service.go
```

**Fix:**
```bash
go fmt ./...
# or
gofmt -s -w .
```

---

### 2. goimports

**Purpose:** Automatically manages import statements

**What it checks:**
- Import organization (stdlib, external, internal)
- Unused imports
- Missing imports
- Import grouping

**Usage:**
```bash
# Install
go install golang.org/x/tools/cmd/goimports@latest

# Check and fix
goimports -w .

# Just check
goimports -l .
```

**CI Integration:**
- Runs after go fmt
- Ensures imports are properly organized
- Removes unused imports

**Example:**
```go
// Before
import (
    "github.com/gin-gonic/gin"
    "fmt"
    "github.com/shyampundkar/project/internal/models"
)

// After (goimports)
import (
    "fmt"

    "github.com/gin-gonic/gin"

    "github.com/shyampundkar/project/internal/models"
)
```

---

### 3. go vet

**Purpose:** Detects suspicious code constructs

**What it checks:**
- Printf format errors
- Unreachable code
- Struct tag issues
- Shadowed variables
- Incorrect composite literals
- Invalid assembly code

**Usage:**
```bash
# Run vet
go vet ./...

# With verbose output
go vet -v ./...
```

**CI Integration:**
- Runs on all modules
- Must pass for build to succeed
- Shows detailed error messages

**Common Issues:**
```go
// Printf format mismatch
fmt.Printf("%s", 42)  // wants string, got int

// Unreachable code
return
fmt.Println("never runs")

// Struct tag issues
type User struct {
    Name string `json:"name,omitempty`  // missing closing quote
}
```

**Fix:** Follow error messages and fix manually

---

### 4. staticcheck

**Purpose:** Advanced static analysis for Go code

**What it checks:**
- Deprecated API usage
- Inefficient code patterns
- Potential bugs
- Code smells
- Style issues

**Usage:**
```bash
# Install
go install honnef.co/go/tools/cmd/staticcheck@latest

# Run all checks
staticcheck -checks=all ./...

# Run specific checks
staticcheck -checks=SA1000,SA1001 ./...
```

**CI Integration:**
- Runs all checks (`-checks=all`)
- More comprehensive than go vet
- Catches subtle bugs

**Example Issues:**
```go
// SA1019: deprecated function
ioutil.ReadFile(filename)  // use os.ReadFile instead

// SA4006: value is never used
x := 5
x = 10  // x was assigned but never used before reassignment

// SA5007: ineffective append
slice = append(slice)  // appending nothing
```

---

### 5. ineffassign

**Purpose:** Detects ineffective assignments

**What it checks:**
- Variables assigned but never used
- Assignments before reassignment
- Dead stores

**Usage:**
```bash
# Install
go install github.com/gordonklaus/ineffassign@latest

# Run
ineffassign ./...
```

**CI Integration:**
- Runs on all modules
- Helps identify dead code

**Example:**
```go
func process() error {
    err := doSomething()  // ineffective assignment
    err = doSomethingElse()
    return err
}

// Fix: remove first assignment if not needed
func process() error {
    _ = doSomething()  // explicitly ignore
    err := doSomethingElse()
    return err
}
```

---

### 6. misspell

**Purpose:** Finds and fixes common spelling mistakes

**What it checks:**
- Comments
- String literals
- Variable names
- Documentation

**Usage:**
```bash
# Install
go install github.com/client9/misspell/cmd/misspell@latest

# Find misspellings
misspell .

# Fix misspellings
misspell -w .

# Error on misspellings
misspell -error .
```

**CI Integration:**
- Runs on all files
- Fails on misspellings
- Helps maintain professionalism

**Common Fixes:**
- "recieve" → "receive"
- "occured" → "occurred"
- "seperator" → "separator"

---

### 7. gosec

**Purpose:** Go security scanner

**What it checks:**
- SQL injection vulnerabilities
- Command injection
- Path traversal
- Weak crypto usage
- Insecure random numbers
- Hardcoded credentials
- Insecure TLS/SSL

**Usage:**
```bash
# Install
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run scan
gosec ./...

# Generate JSON report
gosec -fmt=json -out=report.json ./...

# Generate HTML report
gosec -fmt=html -out=report.html ./...
```

**CI Integration:**
- Runs on order-food (main service)
- Generates JSON report
- Continues on error (warnings only)
- Report uploaded as artifact

**Example Issues:**
```go
// G304: Potential file inclusion via variable
filename := r.URL.Query().Get("file")
data, _ := ioutil.ReadFile(filename)  // DANGEROUS

// G201: SQL injection
query := "SELECT * FROM users WHERE id = " + userId
db.Query(query)  // DANGEROUS

// G401: Weak crypto
h := md5.New()  // Use SHA256 or better

// Fix: Use parameterized queries, validate input, strong crypto
```

---

### 8. golangci-lint

**Purpose:** Meta-linter that runs multiple linters

**What it includes:**
- All the above tools
- Plus 30+ additional linters
- Configurable via `.golangci.yml`

**Usage:**
```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run
golangci-lint run ./...

# With config
golangci-lint run --config=.golangci.yml ./...
```

**CI Integration:**
- Runs as separate job
- Uses project `.golangci.yml` config
- Comprehensive code quality check

**Configuration:** See `.golangci.yml` in project root

---

### 9. go test -race

**Purpose:** Detects data races in concurrent code

**What it checks:**
- Concurrent read/write without synchronization
- Races between goroutines
- Unsafe concurrent map access

**Usage:**
```bash
# Run with race detector
go test -race ./...

# With verbose output
go test -v -race ./...
```

**CI Integration:**
- Runs on all tests
- Must pass for build to succeed
- Critical for concurrent code

**Example Race:**
```go
// RACE CONDITION
var counter int
go func() { counter++ }()
go func() { counter++ }()

// FIX with mutex
var counter int
var mu sync.Mutex
go func() { mu.Lock(); counter++; mu.Unlock() }()
go func() { mu.Lock(); counter++; mu.Unlock() }()

// OR use atomic
var counter int64
go func() { atomic.AddInt64(&counter, 1) }()
go func() { atomic.AddInt64(&counter, 1) }()
```

---

### 10. go test -cover

**Purpose:** Measures test coverage

**What it reports:**
- Line coverage percentage
- Function coverage
- Branch coverage (with -covermode=atomic)

**Usage:**
```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Coverage by function
go tool cover -func=coverage.out
```

**CI Integration:**
- Runs with race detector
- Generates coverage.out
- Creates HTML report
- Uploads to Codecov
- Comments coverage on PRs
- Warns if below threshold (70%)

**Output Example:**
```
github.com/user/project/internal/handler/product_handler.go:15:  ListProducts    100.0%
github.com/user/project/internal/handler/product_handler.go:20:  GetProduct      85.7%
github.com/user/project/internal/service/product_service.go:10:  NewService      100.0%
total:                                                            (statements)    92.3%
```

---

## Coverage Reports

### What's Generated

1. **coverage.out** - Raw coverage data
2. **coverage.html** - Interactive HTML report
3. **Codecov upload** - Integration with codecov.io
4. **PR Comments** - Coverage percentage on PRs
5. **GitHub Artifacts** - Downloadable reports

### Coverage Thresholds

| Module | Threshold | Current |
|--------|-----------|---------|
| database-migration | - | TBD |
| database-load | - | TBD |
| order-food | 70% | TBD |

### Viewing Coverage

**In CI:**
1. Go to Actions → CI Pipeline → Workflow run
2. Click on module job (e.g., build-order-food)
3. Scroll to "Calculate coverage percentage"
4. Download "coverage-report" artifact

**Locally:**
```bash
# Generate and open
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Codecov:**
- Visit https://codecov.io/gh/<username>/<repo>
- View detailed coverage by file
- Track coverage trends over time

---

## Benchmark Testing

**Purpose:** Performance testing and regression detection

**Usage:**
```bash
# Run benchmarks
go test -bench=. ./...

# With memory stats
go test -bench=. -benchmem ./...

# Specific benchmarks
go test -bench=BenchmarkMyFunc -benchmem ./...
```

**CI Integration:**
- Runs on order-food module
- Generates benchmark.txt
- Uploaded as artifact
- Continues on error

**Example Benchmark:**
```go
func BenchmarkListProducts(b *testing.B) {
    repo := repository.NewProductRepository()
    service := service.NewProductService(repo)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.ListProducts()
    }
}
```

**Output:**
```
BenchmarkListProducts-8   100000   12345 ns/op   4096 B/op   64 allocs/op
```

---

## Local Development Workflow

### Before Committing

```bash
# 1. Format code
go fmt ./...
goimports -w .

# 2. Run linters
go vet ./...
staticcheck ./...
ineffassign ./...
misspell -w .

# 3. Run tests
go test -v -race ./...

# 4. Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# 5. Run golangci-lint
golangci-lint run ./...

# 6. Security scan
gosec ./...
```

### Makefile Targets

Add to project Makefile:
```makefile
.PHONY: lint test coverage security

lint: ## Run all linters
	go fmt ./...
	goimports -w .
	go vet ./...
	staticcheck ./...
	ineffassign ./...
	misspell -w .
	golangci-lint run ./...

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage: test ## View coverage report
	go tool cover -html=coverage.out

security: ## Run security scan
	gosec -fmt=html -out=gosec-report.html ./...

all: lint test ## Run all checks
```

---

## CI Pipeline Flow

```
┌─────────────┐
│  Push/PR    │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ Install Tools   │
│ - staticcheck   │
│ - ineffassign   │
│ - misspell      │
│ - goimports     │
│ - gosec         │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Verify Deps     │
│ - go mod verify │
│ - go mod tidy   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Format Checks   │
│ - goimports     │
│ - go fmt        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Static Analysis │
│ - go vet        │
│ - staticcheck   │
│ - ineffassign   │
│ - misspell      │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Security Scan   │
│ - gosec         │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Build           │
│ - Normal        │
│ - Optimized     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Test & Coverage │
│ - go test -race │
│ - coverage.out  │
│ - coverage.html │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Benchmarks      │
│ - go test -bench│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Upload Reports  │
│ - Codecov       │
│ - Artifacts     │
│ - PR Comments   │
└─────────────────┘
```

---

## Troubleshooting

### "goimports: command not found"

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

### "staticcheck: command not found"

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Coverage Below Threshold

1. Add more tests
2. Test error paths
3. Test edge cases
4. Remove dead code

### Race Condition Detected

1. Identify shared variables
2. Add synchronization (mutex, channels)
3. Use atomic operations
4. Redesign to avoid sharing

### Security Issues Found

1. Review gosec report
2. Fix high-severity issues first
3. Use secure alternatives
4. Add input validation

---

## Best Practices

### Writing Testable Code

```go
// Good: testable with interfaces
type ProductService struct {
    repo ProductRepository
}

// Bad: hard to test
type ProductService struct {
    db *sql.DB
}
```

### Coverage Goals

- **Minimum:** 70% for production code
- **Target:** 80-90% for critical paths
- **Not required:** 100% (diminishing returns)

### What to Test

✅ **Always test:**
- Business logic
- Error paths
- Edge cases
- API handlers
- Data validation

❌ **Don't need to test:**
- Generated code
- Third-party libraries
- Simple getters/setters

### Security

✅ **Always check:**
- User input validation
- SQL parameterization
- Path traversal protection
- Authentication/authorization
- Crypto usage

---

## Additional Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [staticcheck docs](https://staticcheck.io/docs/)
- [golangci-lint docs](https://golangci-lint.run/)
- [gosec rules](https://github.com/securego/gosec#available-rules)

---

## Quick Reference

```bash
# Format and organize
go fmt ./... && goimports -w .

# Lint
go vet ./... && staticcheck ./...

# Test
go test -v -race -coverprofile=coverage.out ./...

# Coverage
go tool cover -html=coverage.out

# Security
gosec ./...

# All-in-one
golangci-lint run ./...
```
