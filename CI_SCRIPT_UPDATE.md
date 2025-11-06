# CI Script Update - golangci-lint Integration

## Issue

The local CI script (`./run-ci-local.sh`) was missing the `golangci-lint` check that runs in GitHub Actions CI pipeline. This caused a mismatch where:

- ✅ `./run-ci-local.sh all` would pass locally
- ❌ GitHub Actions would fail with linting errors

### Root Cause

The GitHub Actions workflow (`.github/workflows/ci.yml`) has **two separate linting layers**:

1. **Individual linters** in build jobs:
   - `staticcheck -checks=all ./...`
   - `ineffassign ./...`
   - `misspell -error .`

2. **Comprehensive linter** in separate lint job:
   - `golangci-lint run --timeout=5m`

The local CI script only had layer 1, missing layer 2 entirely.

## What is golangci-lint?

`golangci-lint` is a comprehensive Go linting aggregator that runs multiple linters in parallel:

- Style checkers (ST1000, ST1003, etc.)
- Static analysis tools
- Code smell detectors
- Revive linters
- And many more...

It's configured via `.golangci.yml` and provides:
- Faster execution (parallel linting)
- Unified output format
- Customizable linter selection

## Changes Made

### 1. Updated `run-ci-local.sh`

#### Added golangci-lint Installation

```bash
if ! command -v golangci-lint &> /dev/null; then
    print_info "Installing golangci-lint..."
    # Install golangci-lint using the recommended method
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
fi
```

**Location**: Lines 61-65 in `install_go_tools()` function

#### Added golangci-lint Check

```bash
# Run golangci-lint
print_info "Running golangci-lint..."
golangci-lint run --timeout=5m --out-format=colored-line-number
print_success "golangci-lint check passed"
```

**Location**: Lines 134-137 in `run_module_checks()` function
**Position**: After `misspell` check, before `gosec` check

### 2. Updated `LOCAL_CI_GUIDE.md`

Updated documentation to reflect the new check:

- Added golangci-lint to the "What Gets Checked" list (step 9)
- Updated tool installation instructions
- Added golangci-lint to manual check steps

## Verification

The updated script now matches GitHub Actions CI pipeline exactly:

### Checks Performed (All Modules)
1. ✅ Dependency verification
2. ✅ go.mod/go.sum tidiness
3. ✅ Import formatting (goimports)
4. ✅ Code formatting (go fmt)
5. ✅ Code analysis (go vet)
6. ✅ Advanced linting (staticcheck)
7. ✅ Inefficient assignments (ineffassign)
8. ✅ Spelling (misspell)
9. ✅ **Comprehensive linting (golangci-lint)** ← NEW
10. ✅ Build
11. ✅ Tests with coverage
12. ✅ Coverage report

## Example Error Caught

The missing telemetry package documentation (ST1000) was caught by golangci-lint:

```
internal/telemetry/telemetry.go:1:1: at least one file in a package should have a package comment (ST1000)
```

This error was not caught by `staticcheck -checks=all ./...` running alone, but was caught by golangci-lint's aggregated checks.

## Usage

### Automatic Installation

The script now automatically installs golangci-lint if not present:

```bash
./run-ci-local.sh all
```

### Manual Installation

If you prefer to install manually:

```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
```

### Running Checks

```bash
# Run all checks for all modules
./run-ci-local.sh all

# Run checks for specific module
./run-ci-local.sh order-food

# Run golangci-lint manually
cd order-food
golangci-lint run --timeout=5m --out-format=colored-line-number
```

## Configuration

golangci-lint is configured via `.golangci.yml` in the project root:

```yaml
linters:
  enable:
    - unused          # Replaces: deadcode, structcheck, varcheck
    - copyloopvar     # Replaces: exportloopref
    - staticcheck
    - gosimple
    - govet
    - ineffassign
    - misspell
    - revive
    - stylecheck

output:
  formats:
    - format: colored-line-number
      path: stdout

run:
  timeout: 5m
  skip-dirs:
    - bin
    - vendor
```

## Benefits

### Before (Incomplete Coverage)
```bash
./run-ci-local.sh all
# ✅ Passes locally with staticcheck
# ❌ Fails in GitHub Actions with ST1000 error
```

### After (Complete Coverage)
```bash
./run-ci-local.sh all
# ✅ Catches all issues locally
# ✅ Matches GitHub Actions exactly
# ✅ No surprises in CI
```

## Troubleshooting

### golangci-lint not found after installation

```bash
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc  # or ~/.zshrc
```

### golangci-lint timeout

```bash
# Increase timeout
golangci-lint run --timeout=10m
```

### golangci-lint cache issues

```bash
# Clear cache
golangci-lint cache clean
```

## Summary

The local CI script now provides **complete parity** with GitHub Actions CI pipeline by including golangci-lint checks. This ensures:

- ✅ No surprises in CI
- ✅ Faster feedback loop
- ✅ Consistent development experience
- ✅ Catches all linting issues locally before pushing

**Bottom line**: `./run-ci-local.sh all` now runs **exactly** the same checks as GitHub Actions.
