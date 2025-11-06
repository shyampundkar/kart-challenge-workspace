# Fix: IDE Go Version Mismatch

## Issue

The IDE is reporting:
```
go: go.work requires go >= 1.25 (running go 1.23.2)
```

But the terminal shows Go 1.25.0 is installed.

## Root Cause

Your IDE (VSCode/Goland) is using a cached or different Go installation than your terminal.

## Solution

### For VSCode

1. **Reload Window**
   ```
   Command Palette (Cmd+Shift+P) → "Developer: Reload Window"
   ```

2. **Update Go Extension Settings**

   Open VSCode settings (Cmd+,) and verify:
   ```json
   {
     "go.goroot": "",  // Leave empty to auto-detect
     "go.toolsGopath": "",  // Leave empty to use GOPATH
     "go.useLanguageServer": true
   }
   ```

3. **Restart Go Language Server**
   ```
   Command Palette (Cmd+Shift+P) → "Go: Restart Language Server"
   ```

4. **Clear Go Module Cache** (if needed)
   ```bash
   go clean -modcache
   ```

### Manual Verification

Check which Go version different tools see:

```bash
# Terminal Go version
go version

# Which go binary
which go

# Go environment
go env GOROOT
go env GOPATH

# VSCode Go extension version
# Command Palette → "Go: Locate Configured Go Tools"
```

### Expected Output

```bash
$ go version
go version go1.25.0 darwin/arm64

$ go env GOROOT
/usr/local/go
```

### If Issue Persists

1. **Close VSCode completely**
   ```bash
   # Kill all VSCode processes
   killall "Visual Studio Code" 2>/dev/null
   ```

2. **Clear VSCode cache**
   ```bash
   rm -rf ~/Library/Application\ Support/Code/Cache/*
   rm -rf ~/Library/Application\ Support/Code/CachedData/*
   ```

3. **Restart VSCode**
   ```bash
   code /Users/shyamvijayraopundkar/Documents/GitHub/kart-challenge-workspace
   ```

4. **Reinstall Go Tools**
   ```
   Command Palette (Cmd+Shift+P) → "Go: Install/Update Tools"
   Select all tools and install
   ```

### Alternative: Explicit Go Path

If auto-detection fails, set explicit Go path in VSCode settings:

1. Find your Go installation:
   ```bash
   which go
   # Output: /usr/local/go/bin/go
   ```

2. Add to VSCode settings.json:
   ```json
   {
     "go.goroot": "/usr/local/go",
     "go.alternateTools": {
       "go": "/usr/local/go/bin/go"
     }
   }
   ```

### Verify Fix

After restarting VSCode, the error should be gone. Verify:

```bash
# In VSCode terminal
go version

# Should show: go version go1.25.0 darwin/arm64
```

## Quick Fix (Try This First)

```bash
# 1. Reload VSCode window
# Command Palette → "Developer: Reload Window"

# 2. If that doesn't work, restart Go Language Server
# Command Palette → "Go: Restart Language Server"

# 3. If still not working, close and reopen VSCode
```

## Why This Happens

VSCode caches the Go version when it first starts. After upgrading Go, VSCode may still use the old cached version until:
- The window is reloaded
- The language server is restarted
- VSCode is fully restarted

## Verification

Once fixed, you should see no errors in:
- `database-migration/go.mod`
- `database-load/go.mod`
- `order-food/go.mod`
- `go.work`

All should show Go 1.25 without errors.
