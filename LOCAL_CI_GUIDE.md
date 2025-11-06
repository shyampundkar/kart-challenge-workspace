# Local CI Guide

## Quick Start

Run all CI checks locally before pushing:

```bash
# Run checks for all modules
./run-ci-local.sh all

# Run checks for specific module
./run-ci-local.sh database-migration
./run-ci-local.sh database-load
./run-ci-local.sh order-food
```

## What Gets Checked

The local CI script runs the same checks as GitHub Actions:

### For All Modules
1. ✅ **Dependency verification** - `go mod download` and `go mod verify`
2. ✅ **go.mod/go.sum tidiness** - Ensures dependencies are clean
3. ✅ **Import formatting** - `goimports` checks
4. ✅ **Code formatting** - `go fmt` checks
5. ✅ **Code analysis** - `go vet` static analysis
6. ✅ **Advanced linting** - `staticcheck` with all checks
7. ✅ **Inefficient assignments** - `ineffassign` checks
8. ✅ **Spelling** - `misspell` checks
9. ✅ **Build** - Compiles the binary
10. ✅ **Tests with coverage** - Runs tests with race detection
11. ✅ **Coverage report** - Generates HTML coverage report

### Additional for order-food
12. ✅ **Security scanning** - `gosec` security checks
13. ✅ **Optimized build** - Builds with `-ldflags="-s -w"`
14. ✅ **Benchmarks** - Runs performance benchmarks
15. ✅ **Coverage threshold** - Warns if below 70%

## Output

The script generates these reports:

```
database-migration/
├── coverage.out        # Coverage data
├── coverage.html       # Coverage HTML report
└── bin/
    └── database-migration

database-load/
├── coverage.out
├── coverage.html
└── bin/
    └── database-load

order-food/
├── coverage.out
├── coverage.html
├── gosec-report.json   # Security scan results
├── benchmark.txt       # Benchmark results
└── bin/
    ├── order-food
    └── order-food-optimized
```

## Manual CI Steps

If you prefer to run checks manually:

### 1. Install Tools

```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/gordonklaus/ineffassign@latest
go install github.com/client9/misspell/cmd/misspell@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
```

### 2. Run Checks for Each Module

```bash
cd database-migration  # or database-load, or order-food

# Dependencies
go mod download
go mod verify
go mod tidy

# Formatting
goimports -w .
go fmt ./...

# Linting
go vet -v ./...
staticcheck -checks=all ./...
ineffassign ./...
misspell -error .

# Security (order-food only)
gosec -fmt=json -out=gosec-report.json ./...

# Build
go build -v -o bin/module-name ./cmd/main.go

# Tests
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Coverage
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Benchmarks (order-food only)
go test -bench=. -benchmem -run=^$ ./...
```

## Using `act` (GitHub Actions locally)

### Install

```bash
# macOS
brew install act

# Linux
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash
```

### Run Workflows

```bash
# List all jobs
act -l

# Run all jobs
act push

# Run specific job
act -j build-database-migration
act -j build-database-load
act -j build-order-food
act -j lint
act -j security-scan

# Run with specific event
act pull_request

# Dry run (just show what would run)
act -n
```

### Configure act

Create `.actrc` in project root:

```bash
cat > .actrc <<EOF
--platform ubuntu-latest=catthehacker/ubuntu:full-latest
--container-architecture linux/amd64
EOF
```

## Quick Checks Before Commit

### Minimal Check (Fast)

```bash
# Just format and build
cd database-migration && go fmt ./... && go build ./cmd/main.go && cd ..
cd database-load && go fmt ./... && go build ./cmd/main.go && cd ..
cd order-food && go fmt ./... && go build ./cmd/main.go && cd ..
```

### Standard Check (Recommended)

```bash
# Format, vet, and test
./run-ci-local.sh all
```

### Full Check (Complete)

```bash
# Run everything including act
./run-ci-local.sh all
act -j lint
act -j security-scan
```

## Pre-commit Hook

Add a pre-commit hook to automatically run checks:

```bash
cat > .git/hooks/pre-commit <<'EOF'
#!/bin/bash

echo "Running pre-commit checks..."

# Run goimports and go fmt for all modules
for module in database-migration database-load order-food; do
    echo "Checking $module..."
    cd "$module"

    # Format code
    goimports -w .
    go fmt ./...

    # Run go vet
    if ! go vet ./...; then
        echo "go vet failed for $module"
        exit 1
    fi

    cd ..
done

# Add formatted files
git add -u

echo "Pre-commit checks passed!"
EOF

chmod +x .git/hooks/pre-commit
```

## Continuous Integration

### GitHub Actions
- Automatically runs on push to `main`, `develop`, `feature/**`
- Runs on pull requests to `main`, `develop`
- Results visible at: https://github.com/[username]/[repo]/actions

### What Happens on GitHub
1. **changes** job detects which modules changed
2. **build-{module}** jobs run in parallel for changed modules
3. **lint** job runs golangci-lint for all modules
4. **security-scan** job runs Trivy vulnerability scanner
5. **ci-summary** job aggregates all results

### Coverage Reports
- Uploaded to Codecov
- HTML reports available as artifacts
- PR comments show coverage percentage

### Artifacts
Available for 7-30 days:
- Coverage reports (HTML)
- Security scan results (JSON)
- Benchmark results (TXT)
- Build binaries

## Troubleshooting

### "go: command not found"
```bash
# Install Go 1.23.2
brew install go@1.23  # macOS
# or download from https://golang.org/dl/
```

### "staticcheck: command not found"
```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### "act: command not found"
```bash
brew install act  # macOS
```

### "Permission denied: ./run-ci-local.sh"
```bash
chmod +x run-ci-local.sh
```

### Coverage reports not generating
```bash
# Ensure you have tests
go test ./...

# Check if coverage.out exists
ls -la */coverage.out
```

## Best Practices

1. **Run locally before pushing**
   ```bash
   ./run-ci-local.sh all
   ```

2. **Fix formatting issues**
   ```bash
   goimports -w .
   go fmt ./...
   ```

3. **Check coverage**
   ```bash
   # Open coverage report in browser
   open database-migration/coverage.html
   open database-load/coverage.html
   open order-food/coverage.html
   ```

4. **Review security issues**
   ```bash
   # Check gosec report
   cat order-food/gosec-report.json | jq
   ```

5. **Monitor benchmarks**
   ```bash
   # Compare benchmark results
   cat order-food/benchmark.txt
   ```

## Summary

| Method | Speed | Coverage | Use Case |
|--------|-------|----------|----------|
| `./run-ci-local.sh` | ⚡⚡ Fast | ✅ Complete | **Recommended** - Before commit |
| `act` | 🐢 Slow | ✅ Complete | Full GitHub Actions simulation |
| Manual steps | ⚡⚡⚡ Fastest | 🔧 Selective | Quick specific checks |
| Pre-commit hook | ⚡⚡⚡ Auto | ⚠️ Basic | Automatic formatting |

**Recommended workflow:**
```bash
# 1. Make changes
# 2. Run local CI
./run-ci-local.sh all

# 3. Review reports
open */coverage.html

# 4. Commit if all passed
git add .
git commit -m "Your message"

# 5. Push
git push
```
