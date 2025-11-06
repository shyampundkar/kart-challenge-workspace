# Auto-Installation Feature

## Summary

Updated `deploy.sh` to automatically install **Helm** if it's not found on the system, similar to how it already auto-installs **minikube**.

## What Changed

### Modified File
- **[deploy.sh](deploy.sh:40-55)** - Added Helm auto-installation functions

### New Functions Added

#### 1. `install_helm_mac()`
Installs Helm on macOS:
- Uses Homebrew if available
- Falls back to official Helm install script if Homebrew is not found

#### 2. `install_helm_linux()`
Installs Helm on Linux:
- Uses official Helm install script

### Updated Function

#### `check_dependencies()`
Now automatically installs Helm if not found:
- Detects operating system (macOS or Linux)
- Calls appropriate installation function
- Shows success message after installation

## Installation Methods

### macOS
```bash
# With Homebrew (preferred)
brew install helm

# Without Homebrew (fallback)
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### Linux
```bash
# Official Helm install script
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## Dependencies Status

| Dependency | Required | Auto-Install | Notes |
|------------|----------|--------------|-------|
| **Docker** | ✅ Yes | ❌ No | Must be installed manually |
| **kubectl** | ✅ Yes | ❌ No | Must be installed manually |
| **Helm** | ✅ Yes | ✅ **NEW** | Auto-installs if missing |
| **minikube** | ✅ Yes | ✅ Yes | Auto-installs if missing |

## User Experience

### Before
```bash
$ ./deploy.sh
[INFO] Checking dependencies...
[ERROR] Helm is not installed. Please install Helm first.
# User must manually install Helm and run script again
```

### After
```bash
$ ./deploy.sh
[INFO] Checking dependencies...
[WARNING] Helm is not installed. Installing...
[INFO] Installing Helm on macOS...
[SUCCESS] Helm installed successfully
# Script continues automatically
```

## How It Works

```
1. Check if Helm exists
   ├─ YES → Continue
   └─ NO → Install Helm
       ├─ Detect OS (macOS or Linux)
       ├─ macOS:
       │   ├─ Check for Homebrew
       │   ├─ YES → brew install helm
       │   └─ NO → Use Helm install script
       └─ Linux:
           └─ Use Helm install script
```

## Testing

### Test Auto-Installation

```bash
# Uninstall Helm (if installed)
# macOS:
brew uninstall helm

# Linux:
sudo rm /usr/local/bin/helm

# Run deployment script
./deploy.sh

# Helm will be automatically installed
```

### Verify Installation

```bash
# Check Helm version
helm version

# Should show: version.BuildInfo{Version:"v3.x.x", ...}
```

## Benefits

✅ **Improved User Experience** - One less manual step
✅ **Faster Setup** - No need to leave script to install Helm
✅ **Consistency** - Same auto-install behavior as minikube
✅ **Cross-Platform** - Works on both macOS and Linux
✅ **Fallback Options** - Multiple installation methods

## Supported Platforms

- ✅ **macOS** (Darwin) - With or without Homebrew
- ✅ **Linux** - All distributions
- ❌ **Windows** - Not supported (script uses bash)

## Documentation Updated

- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Updated prerequisites section to reflect auto-installation

## Example Run

```bash
$ ./deploy.sh

[INFO] ===== Starting Deployment to Minikube =====

[INFO] Checking dependencies...
[SUCCESS] Docker found
[SUCCESS] Docker daemon is running
[SUCCESS] kubectl found
[WARNING] Helm is not installed. Installing...
[INFO] Installing Helm on macOS...
Downloading https://get.helm.sh/helm-v3.13.0-darwin-amd64.tar.gz
Verifying checksum... Done.
Preparing to install helm into /usr/local/bin
helm installed into /usr/local/bin/helm
[SUCCESS] Helm installed successfully
[SUCCESS] minikube found

[INFO] Checking minikube status...
[SUCCESS] minikube is already running
[SUCCESS] minikube is ready

[INFO] Building Docker images...
# ... continues with deployment
```

## Troubleshooting

### Issue: Helm Installation Fails

**Symptoms:**
```
Error: Failed to download helm
```

**Solutions:**
```bash
# Check internet connection
ping -c 3 get.helm.sh

# Install manually
# macOS:
brew install helm

# Linux:
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### Issue: Permission Denied

**Symptoms:**
```
Permission denied when installing helm
```

**Solution:**
```bash
# The script uses sudo for installations
# Ensure you have sudo privileges
sudo -v

# Or install manually with sudo
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | sudo bash
```

## Manual Installation (If Needed)

If auto-installation fails, install manually:

### macOS
```bash
# Option 1: Homebrew (recommended)
brew install helm

# Option 2: Direct download
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### Linux
```bash
# Option 1: Package manager
# Debian/Ubuntu
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm

# Option 2: Install script
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## Summary

The deploy.sh script now provides a **fully automated setup experience** for Helm and minikube, requiring users to only install Docker and kubectl manually. This streamlines the deployment process and reduces setup friction.

**Result**: Users can run `./deploy.sh` and have Helm and minikube automatically installed if missing, making the setup process faster and more user-friendly.
