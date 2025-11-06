# golangci-lint Configuration Fix

## Problem

The CI was failing with deprecated linter errors:

```
level=error msg="[linters_context] deadcode: This linter is fully inactivated"
level=error msg="[linters_context] exportloopref: This linter is fully inactivated"
level=error msg="[linters_context] structcheck: This linter is fully inactivated"
level=error msg="[linters_context] varcheck: This linter is fully inactivated"
Error: golangci-lint exit with code 7
```

## Root Cause

The `.golangci.yml` configuration was using deprecated linters:
- `deadcode` - Deprecated since v1.49.0
- `structcheck` - Deprecated since v1.49.0
- `varcheck` - Deprecated since v1.49.0
- `exportloopref` - Deprecated since v1.60.2
- `exhaustivestruct` - Deprecated
- `gomnd` - Deprecated

## Solution

Updated [.golangci.yml](.golangci.yml) to use modern linter replacements:

### Deprecated Linters Removed

| Old Linter | Status | Replacement |
|------------|--------|-------------|
| `deadcode` | ❌ Removed | `unused` |
| `structcheck` | ❌ Removed | `unused` |
| `varcheck` | ❌ Removed | `unused` |
| `exportloopref` | ❌ Removed | `copyloopvar` |
| `exhaustivestruct` | ❌ Removed | N/A (not needed) |
| `gomnd` | ❌ Removed | `mnd` (disabled) |

### Modern Linters Enabled

✅ **Essential Linters:**
- `gofmt` - Code formatting
- `govet` - Go tool vet
- `errcheck` - Unchecked errors
- `staticcheck` - Advanced static analysis
- `unused` - **NEW** - Replaces deadcode/structcheck/varcheck
- `gosimple` - Simplification suggestions
- `ineffassign` - Ineffectual assignments
- `typecheck` - Type checking

✅ **Code Quality:**
- `goconst` - Repeated strings
- `gocyclo` - Cyclomatic complexity
- `goimports` - Import formatting
- `misspell` - Spelling
- `unparam` - Unused parameters
- `unconvert` - Unnecessary conversions
- `dupl` - Code duplication
- `gocritic` - Code critiques
- `gochecknoinits` - Init functions check
- `whitespace` - Whitespace issues
- `revive` - Fast linter with many rules

✅ **Security:**
- `gosec` - Security issues

✅ **Modern:**
- `copyloopvar` - **NEW** - Loop variable capture (Go 1.22+)

### Configuration Changes

#### Before (Deprecated)
```yaml
linters:
  enable:
    - deadcode         # ❌ Deprecated
    - structcheck      # ❌ Deprecated
    - varcheck         # ❌ Deprecated
    - exportloopref    # ❌ Deprecated
  disable:
    - exhaustivestruct # ❌ Deprecated
    - gomnd            # ❌ Deprecated

output:
  format: colored-line-number  # ⚠️ Deprecated config
```

#### After (Modern)
```yaml
run:
  go: '1.23'

linters:
  enable:
    - unused           # ✅ Replaces deadcode, structcheck, varcheck
    - copyloopvar      # ✅ Replaces exportloopref
  disable:
    - mnd              # ✅ Replacement for gomnd (disabled)

output:
  formats:            # ✅ Modern config
    - format: colored-line-number
```

## Testing

### Test Locally

```bash
# Install golangci-lint (if not installed)
brew install golangci-lint  # macOS
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run for all modules
golangci-lint run --config .golangci.yml ./database-migration/...
golangci-lint run --config .golangci.yml ./database-load/...
golangci-lint run --config .golangci.yml ./order-food/...

# Or from each module directory
cd database-migration && golangci-lint run
cd database-load && golangci-lint run
cd order-food && golangci-lint run
```

### Verify No Deprecation Warnings

```bash
# Should NOT show any of these:
# ❌ "The linter 'deadcode' is deprecated"
# ❌ "The linter 'structcheck' is deprecated"
# ❌ "The linter 'varcheck' is deprecated"
# ❌ "The linter 'exportloopref' is deprecated"
# ❌ "exhaustivestruct is deprecated"
# ❌ "gomnd is deprecated"

# Should show:
# ✅ Clean run or only actual code issues
```

## Benefits

### 1. No More Deprecation Warnings
- ✅ CI runs cleanly
- ✅ No "fully inactivated" errors
- ✅ Exit code 0 (success)

### 2. Better Coverage
- ✅ `unused` is more comprehensive than deadcode/structcheck/varcheck
- ✅ `copyloopvar` handles Go 1.22+ loop variable semantics
- ✅ Modern linters find more issues

### 3. Future-Proof
- ✅ Uses actively maintained linters
- ✅ Compatible with Go 1.23
- ✅ Ready for future Go versions

### 4. Consistent with Best Practices
- ✅ Follows golangci-lint recommendations
- ✅ Uses official replacement linters
- ✅ Modern output format configuration

## Linter Mappings

### `unused` Replaces Three Linters

The `unused` linter is a comprehensive replacement for:

```yaml
# Old (3 separate linters)
- deadcode      # Find unused code
- structcheck   # Find unused struct fields
- varcheck      # Find unused global variables and constants

# New (1 unified linter)
- unused        # Finds all unused: code, struct fields, variables, constants
```

### `copyloopvar` for Go 1.22+

```yaml
# Old (pre-Go 1.22)
- exportloopref  # Prevent loop variable capture bugs

# New (Go 1.22+ with loopvar)
- copyloopvar    # Handle loop variable semantics properly
```

Go 1.22 introduced automatic loop variable capture, making `exportloopref` unnecessary.

## Configuration Best Practices

### Modern Output Format

```yaml
# Old format (deprecated)
output:
  format: colored-line-number

# New format (recommended)
output:
  formats:
    - format: colored-line-number
  print-issued-lines: true
  print-linter-name: true
  sort-results: true
```

### Go Version Specification

```yaml
run:
  go: '1.23'  # Specify Go version for linter compatibility
```

### Test File Exclusions

```yaml
issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - dupl        # Allow duplication in tests
        - gosec       # Less strict security in tests
        - goconst     # Allow repeated strings in tests
        - gocyclo     # Allow complex tests
```

## CI Integration

The updated configuration works seamlessly with GitHub Actions:

```yaml
# .github/workflows/ci.yml
- name: Run golangci-lint for database-migration
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest
    working-directory: database-migration
    args: --timeout=5m --out-format=colored-line-number

# ✅ No more exit code 7 errors!
```

## Local Development

### Quick Lint Check

```bash
# From project root
golangci-lint run ./...

# From specific module
cd database-migration
golangci-lint run
```

### Auto-fix Issues

```bash
# Fix imports and formatting
goimports -w .
go fmt ./...

# Run linter
golangci-lint run --fix
```

### VS Code Integration

Add to `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast"
  ]
}
```

## Troubleshooting

### Issue: "unknown linter 'copyloopvar'"

**Solution:** Update golangci-lint to v1.60.2+

```bash
# macOS
brew upgrade golangci-lint

# Linux/Go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Issue: Too many issues reported

**Solution:** Enable linters gradually

```yaml
linters:
  disable:
    - gocritic  # Temporarily disable strict linters
    - revive
```

### Issue: Linter taking too long

**Solution:** Adjust timeout and concurrency

```yaml
run:
  timeout: 10m
  concurrency: 4  # Adjust based on CPU cores
```

## Summary

### Changes Made
- ✅ Removed 6 deprecated linters
- ✅ Added 2 modern replacement linters
- ✅ Updated output format configuration
- ✅ Added Go version specification
- ✅ Improved test file exclusions

### Result
- ✅ CI passes without deprecation warnings
- ✅ No "fully inactivated" errors
- ✅ Better code analysis with modern linters
- ✅ Future-proof configuration

### Files Updated
- [.golangci.yml](.golangci.yml) - Main linter configuration

## References

- [golangci-lint Linters](https://golangci-lint.run/usage/linters/)
- [Linter Deprecation Cycle](https://golangci-lint.run/product/roadmap/#linter-deprecation-cycle)
- [Go 1.22 Loop Variable Changes](https://go.dev/blog/loopvar-preview)
- [unused Linter](https://github.com/dominikh/go-tools/tree/master/unused)
- [copyloopvar Linter](https://github.com/timakin/copyloopvar)
