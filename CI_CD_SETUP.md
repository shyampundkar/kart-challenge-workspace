# CI/CD Setup Summary

Complete GitHub Actions CI/CD pipeline implementation for kart-challenge-workspace.

## What's Been Created

### ✅ GitHub Actions Workflows (7 workflows)

| Workflow | Purpose | Status |
|----------|---------|--------|
| 🔧 **CI Pipeline** | Build, test, lint all modules | ✅ Ready |
| 🐳 **Docker Build** | Build & push container images | ✅ Ready |
| ⎈ **Helm Validation** | Validate & test Helm charts | ✅ Ready |
| 🔍 **Pull Request** | PR validation & checks | ✅ Ready |
| 🚀 **Release** | Automated releases | ✅ Ready |
| 🔒 **Security Scan** | CodeQL & vulnerability scanning | ✅ Ready |
| 📦 **Deploy** | Manual deployment to environments | ✅ Ready |

### ✅ Configuration Files

| File | Purpose |
|------|---------|
| `.golangci.yml` | Go linter configuration |
| `.github/dependabot.yml` | Dependency update automation |
| `.github/PULL_REQUEST_TEMPLATE.md` | Standardized PR template |

### ✅ Documentation

| Document | Description |
|----------|-------------|
| `.github/CI_CD_GUIDE.md` | Complete CI/CD guide |
| `.github/workflows/README.md` | Workflows reference |
| `CI_CD_SETUP.md` | This file - setup summary |

## Quick Start

### 1. Enable GitHub Actions

GitHub Actions is enabled by default. Verify in:
- Repository Settings → Actions → General

### 2. Configure Container Registry

Enable GitHub Container Registry:
- Repository Settings → Actions → General
- Workflow permissions: "Read and write permissions"

### 3. Set Up Secrets (Optional)

For cloud deployments, add secrets:
- Repository Settings → Secrets and variables → Actions
- Add environment-specific secrets

### 4. Create Environments (Optional)

For deployment workflow:
- Repository Settings → Environments
- Create: `development`, `staging`, `production`
- Configure protection rules

## How It Works

### On Every Push/PR

```
┌─────────────┐
│  Git Push   │
└──────┬──────┘
       │
       ├─→ CI Pipeline
       │   ├─ Detect changes
       │   ├─ Build modules
       │   ├─ Run tests
       │   ├─ Code coverage
       │   ├─ Security scan
       │   └─ Lint code
       │
       ├─→ Pull Request Validation (if PR)
       │   ├─ Validate title
       │   ├─ Check conflicts
       │   ├─ Code review
       │   └─ Size labeling
       │
       └─→ Security Scanning
           ├─ CodeQL analysis
           ├─ Gosec scan
           └─ Dependency review
```

### On Push to Main

```
┌──────────────┐
│ Push to main │
└──────┬───────┘
       │
       ├─→ CI Pipeline (above)
       │
       ├─→ Docker Build & Push
       │   ├─ Build images
       │   ├─ Tag: latest, main-sha
       │   ├─ Push to ghcr.io
       │   ├─ Security scan
       │   └─ Test images
       │
       └─→ Helm Validation
           ├─ Lint charts
           ├─ Test in Kind
           └─ Package charts
```

### On Version Tag (v*.*.*)

```
┌──────────────┐
│  Tag v1.0.0  │
└──────┬───────┘
       │
       └─→ Release Workflow
           ├─ Create GitHub release
           ├─ Generate changelog
           ├─ Build Docker images
           │  └─ Tag: v1.0.0, v1.0, v1, latest
           ├─ Build binaries
           │  ├─ Linux (amd64, arm64)
           │  ├─ macOS (amd64, arm64)
           │  └─ Windows (amd64, arm64)
           ├─ Package Helm charts
           └─ Upload all assets
```

### Manual Deployment

```
┌──────────────┐
│ Run Deploy   │
└──────┬───────┘
       │
       └─→ Deploy Workflow
           ├─ Select environment
           ├─ Select version
           ├─ Deploy with Helm
           ├─ Verify deployment
           ├─ Run smoke tests
           └─ Rollback if failed
```

## Usage Examples

### Create a Feature Branch

```bash
# Create branch
git checkout -b feature/new-api-endpoint

# Make changes
# ... edit files ...

# Commit with conventional commit
git add .
git commit -m "feat(order-food): add payment endpoint"

# Push
git push origin feature/new-api-endpoint
```

**What happens:**
- ✅ CI pipeline runs
- ✅ Tests execute
- ✅ Code is linted
- ✅ Security scan runs

### Create a Pull Request

```bash
# Via GitHub CLI
gh pr create --title "feat: add payment endpoint" \
  --body "Adds payment processing endpoint"

# Or via GitHub web UI
```

**What happens:**
- ✅ PR validation runs
- ✅ Title is validated
- ✅ Conflicts checked
- ✅ Size labeled
- ✅ CI runs again
- ✅ Summary posted

### Merge to Main

```bash
# Merge PR via GitHub UI
# Or via CLI
gh pr merge <pr-number> --squash
```

**What happens:**
- ✅ All workflows run
- ✅ Docker images built
- ✅ Images pushed to ghcr.io
- ✅ Tagged with: latest, main-abc1234
- ✅ Helm charts validated

### Create a Release

```bash
# Create and push tag
git checkout main
git pull
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

**What happens:**
- ✅ Release workflow triggers
- ✅ GitHub release created
- ✅ Changelog generated
- ✅ Docker images built & pushed
- ✅ Binaries built (all platforms)
- ✅ Helm charts packaged
- ✅ All assets uploaded

**Release includes:**
```
Release v1.0.0
├── Docker Images
│   ├── ghcr.io/<user>/database-migration:v1.0.0
│   ├── ghcr.io/<user>/database-load:v1.0.0
│   └── ghcr.io/<user>/order-food:v1.0.0
├── Binaries
│   ├── database-migration-linux-amd64
│   ├── database-migration-linux-arm64
│   ├── database-migration-darwin-amd64
│   ├── database-migration-darwin-arm64
│   ├── database-migration-windows-amd64.exe
│   ├── database-migration-windows-arm64.exe
│   └── ... (same for other modules)
└── Helm Charts
    ├── database-migration-0.1.0.tgz
    ├── database-load-0.1.0.tgz
    ├── order-food-0.1.0.tgz
    └── index.yaml
```

### Deploy to Environment

```bash
# Via GitHub CLI
gh workflow run deploy.yml \
  -f environment=staging \
  -f version=v1.0.0

# Via GitHub UI
# Actions → Deploy to Kubernetes → Run workflow
# Select environment and version
```

**What happens:**
- ✅ Connects to Kubernetes
- ✅ Deploys with Helm
- ✅ Waits for ready
- ✅ Runs smoke tests
- ✅ Notifies on completion
- ❌ Rollback if failed

## Container Images

### Image Naming Convention

```
ghcr.io/<username>/<module>:<tag>

Examples:
ghcr.io/shyampundkar/order-food:latest
ghcr.io/shyampundkar/order-food:v1.0.0
ghcr.io/shyampundkar/order-food:main-abc1234
```

### Using Images

```bash
# Pull image
docker pull ghcr.io/<username>/order-food:latest

# Run locally
docker run -p 8080:8080 ghcr.io/<username>/order-food:latest

# Use in Kubernetes
kubectl set image deployment/order-food \
  order-food=ghcr.io/<username>/order-food:v1.0.0

# Use in Helm
helm upgrade order-food ./order-food/helm \
  --set image.repository=ghcr.io/<username>/order-food \
  --set image.tag=v1.0.0
```

### Image Tags Explained

| Tag | Created On | Example | Use Case |
|-----|-----------|---------|----------|
| `latest` | Push to main | `latest` | Development |
| `v1.0.0` | Version tag | `v1.0.0` | Production |
| `v1.0` | Version tag | `v1.0` | Minor version pin |
| `v1` | Version tag | `v1` | Major version pin |
| `main-abc123` | Push to main | `main-abc1234` | Specific commit |
| `pr-123` | Pull request | `pr-123` | PR testing |

## Security Features

### Automated Security Scanning

1. **Trivy** - Container vulnerability scanning
2. **CodeQL** - Code analysis for security issues
3. **Gosec** - Go-specific security scanner
4. **Dependabot** - Dependency updates

### Security Reports

View in: Security → Code scanning alerts

### Dependency Updates

Dependabot automatically creates PRs for:
- Go module updates (weekly)
- GitHub Actions updates (weekly)
- Docker base image updates (weekly)

## Monitoring

### Status Badges

Add to README.md:

```markdown
![CI](https://github.com/<user>/<repo>/actions/workflows/ci.yml/badge.svg)
![Docker](https://github.com/<user>/<repo>/actions/workflows/docker.yml/badge.svg)
![Helm](https://github.com/<user>/<repo>/actions/workflows/helm.yml/badge.svg)
![Security](https://github.com/<user>/<repo>/actions/workflows/codeql.yml/badge.svg)
```

### View Workflow Status

```bash
# Install GitHub CLI
brew install gh

# List workflows
gh workflow list

# View recent runs
gh run list

# Watch live
gh run watch

# View logs
gh run view <run-id> --log

# Download artifacts
gh run download <run-id>
```

## Cost Optimization

### GitHub Actions Minutes

**Free tier:**
- Public repos: Unlimited
- Private repos: 2,000 minutes/month

**Optimization tips:**
1. Use path filters (only build changed modules)
2. Cancel redundant runs
3. Use caching (Go modules, Docker layers)
4. Run expensive jobs conditionally

### Storage

**Free tier:**
- 500 MB package storage
- 1 GB artifacts storage

**Optimization tips:**
1. Set artifact retention (7-30 days)
2. Clean up old packages
3. Use external registries for large images

## Troubleshooting

### Workflow Not Running

**Check:**
- Branch name matches trigger pattern
- Workflow file has no syntax errors
- Actions are enabled in repository settings

**Fix:**
```bash
# Validate workflow
gh workflow view ci.yml

# Check syntax
yamllint .github/workflows/ci.yml
```

### Build Fails

**Debug:**
1. View error logs in Actions tab
2. Test locally:
   ```bash
   cd order-food
   go test ./...
   go build ./cmd/main.go
   ```
3. Check dependencies:
   ```bash
   go mod verify
   go mod tidy
   ```

### Docker Push Fails

**Check:**
- GITHUB_TOKEN has package write permission
- Container registry is accessible
- Image name is correct

**Fix:**
```bash
# Test locally
docker build -t test ./order-food
docker tag test ghcr.io/<user>/order-food:test
docker push ghcr.io/<user>/order-food:test
```

### Deployment Fails

**Check:**
- Kubernetes credentials are valid
- Namespace exists
- Image is accessible
- Helm chart is valid

**Debug:**
```bash
# Test Helm chart
helm lint ./order-food/helm
helm template test ./order-food/helm

# Test deployment locally
helm install test ./order-food/helm --dry-run
```

## Next Steps

### 1. Customize Workflows

Edit workflows in `.github/workflows/` to:
- Add project-specific steps
- Configure notifications
- Add custom tests
- Integrate with external services

### 2. Set Up Environments

Configure deployment environments:
- Add environment-specific secrets
- Set up protection rules
- Configure required reviewers

### 3. Add Monitoring

Integrate monitoring tools:
- Prometheus metrics
- Grafana dashboards
- Slack notifications
- PagerDuty alerts

### 4. Enhance Testing

Add more comprehensive tests:
- Integration tests
- E2E tests
- Performance tests
- Load tests

### 5. Documentation

Keep documentation updated:
- Update README with badges
- Document environment setup
- Add runbooks for common issues

## Resources

- [CI/CD Guide](.github/CI_CD_GUIDE.md) - Detailed guide
- [Workflows README](.github/workflows/README.md) - Workflow reference
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Container Registry Docs](https://docs.github.com/en/packages)

## Support

For CI/CD issues:
1. Check workflow logs
2. Review documentation
3. Test locally
4. Create issue with logs
