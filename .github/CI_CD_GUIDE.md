# CI/CD Pipeline Guide

This document describes the complete CI/CD setup for the kart-challenge-workspace project.

## Overview

The project uses GitHub Actions for continuous integration and deployment with the following workflows:

1. **CI Pipeline** - Build, test, and validate code
2. **Docker Build** - Build and push container images
3. **Helm Validation** - Validate and test Helm charts
4. **Pull Request** - PR-specific validations
5. **Release** - Automated releases with binaries and charts
6. **CodeQL** - Security scanning
7. **Deploy** - Manual deployment to environments

## Workflows

### 1. CI Pipeline (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main`, `develop`, `feature/**` branches
- Pull requests to `main`, `develop`

**Jobs:**
- **changes** - Detects which modules changed using path filters
- **build-database-migration** - Builds and tests database-migration
- **build-database-load** - Builds and tests database-load
- **build-order-food** - Builds and tests order-food
- **security-scan** - Runs Trivy vulnerability scanner
- **lint** - Runs golangci-lint for all modules

**Features:**
- Smart module detection (only builds changed modules)
- Go 1.23.2
- Code coverage with Codecov integration
- Format checking with `gofmt`
- Vet checks
- Dependency verification
- Race condition detection
- Artifact upload for binaries

**Example:**
```bash
# This workflow runs automatically on push/PR
# To see results: GitHub Actions tab in your repository
```

### 2. Docker Build and Push (`.github/workflows/docker.yml`)

**Triggers:**
- Push to `main` branch
- Tags matching `v*.*.*`
- Pull requests to `main`

**Jobs:**
- **build-database-migration** - Builds and pushes database-migration image
- **build-database-load** - Builds and pushes database-load image
- **build-order-food** - Builds and pushes order-food image
- **test-order-food** - Tests the order-food image

**Features:**
- Multi-platform builds (linux/amd64, linux/arm64)
- Automatic tagging:
  - `latest` for main branch
  - `v1.0.0`, `v1.0`, `v1` for semantic version tags
  - Branch names for feature branches
  - PR numbers for pull requests
- GitHub Container Registry (ghcr.io)
- Docker layer caching
- Security scanning with Trivy
- Automated image testing

**Container Registry:**
Images are pushed to: `ghcr.io/<username>/<module>:<tag>`

**Example:**
```bash
# Pull an image
docker pull ghcr.io/<username>/order-food:latest

# Use in Kubernetes
kubectl set image deployment/order-food \
  order-food=ghcr.io/<username>/order-food:v1.0.0
```

### 3. Helm Chart Validation (`.github/workflows/helm.yml`)

**Triggers:**
- Changes to `**/helm/**` directories
- Changes to workflow file

**Jobs:**
- **lint-and-validate** - Lints and validates charts
- **test-install** - Tests installation in Kind cluster
- **package-charts** - Packages charts for distribution

**Features:**
- Helm linting
- Chart testing (ct)
- Manifest validation with kubeval
- Automated testing in Kind cluster
- End-to-end deployment tests
- Chart packaging and indexing

**Example:**
```bash
# Local testing
helm lint database-migration/helm
helm template database-migration database-migration/helm
```

### 4. Pull Request Validation (`.github/workflows/pr.yml`)

**Triggers:**
- Pull request opened, synchronized, or reopened

**Jobs:**
- **pr-info** - Displays PR information
- **validate-pr-title** - Validates semantic PR titles
- **check-conflicts** - Checks for merge conflicts
- **code-review** - Automated code review checks
- **size-label** - Labels PR by size
- **test-summary** - Posts validation summary

**PR Title Format:**
```
<type>: <description>

Types: feat, fix, docs, style, refactor, perf, test, build, ci, chore, revert

Examples:
feat: add user authentication
fix: resolve order processing bug
docs: update API documentation
```

### 5. Release Workflow (`.github/workflows/release.yml`)

**Triggers:**
- Tags matching `v*.*.*` (e.g., v1.0.0)

**Jobs:**
- **create-release** - Creates GitHub release with changelog
- **build-and-push-images** - Builds and pushes Docker images
- **build-binaries** - Builds cross-platform binaries
- **package-helm-charts** - Packages and uploads Helm charts

**Release Assets:**
- Docker images for all platforms
- Binaries for:
  - Linux (amd64, arm64)
  - macOS (amd64, arm64)
  - Windows (amd64, arm64)
- Helm chart packages (.tgz)
- Helm repository index

**Creating a Release:**
```bash
# Tag and push
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions will:
# 1. Create release
# 2. Build and push images
# 3. Build binaries for all platforms
# 4. Package Helm charts
# 5. Upload all assets
```

### 6. Security Scanning (`.github/workflows/codeql.yml`)

**Triggers:**
- Push to `main`, `develop`
- Pull requests
- Weekly schedule (Monday midnight)

**Jobs:**
- **analyze** - CodeQL analysis
- **gosec** - Go security scanner
- **dependency-review** - Dependency vulnerability check

**Features:**
- Automated security vulnerability detection
- SARIF report upload to GitHub Security
- Go-specific security checks
- Dependency vulnerability scanning

### 7. Deployment Workflow (`.github/workflows/deploy.yml`)

**Triggers:**
- Manual workflow dispatch

**Inputs:**
- `environment` - development, staging, or production
- `version` - Tag/version to deploy (optional)

**Jobs:**
- **deploy** - Deploys all modules to Kubernetes
  - Updates Helm releases
  - Verifies deployment
  - Runs smoke tests
  - Automatic rollback on failure

**Manual Deployment:**
```bash
# Via GitHub UI:
# Actions → Deploy to Kubernetes → Run workflow
# Select environment and version

# Via GitHub CLI:
gh workflow run deploy.yml \
  -f environment=staging \
  -f version=v1.0.0
```

## Configuration Files

### `.golangci.yml`

Configures golangci-lint with multiple linters:
- Code quality checks
- Performance analysis
- Style enforcement
- Security checks
- Complexity analysis

**Local Usage:**
```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run
golangci-lint run ./...
```

### `.github/dependabot.yml`

Configures automated dependency updates:
- Go modules (weekly)
- GitHub Actions (weekly)
- Docker base images (weekly)

**Features:**
- Automatic PR creation
- Semantic commit messages
- Proper labeling
- Grouped updates

### `.github/PULL_REQUEST_TEMPLATE.md`

Standardized PR template ensuring:
- Clear description
- Type of change
- Testing verification
- Checklist compliance

## Secrets and Variables

### Required Secrets

**For Docker Registry:**
- `GITHUB_TOKEN` - Automatically provided (GitHub Container Registry)

**For Kubernetes Deployment (if using cloud providers):**
- `KUBE_CONFIG` - Kubernetes configuration
- `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` (for EKS)
- `GCP_SA_KEY` (for GKE)
- `AZURE_CREDENTIALS` (for AKS)

### Environment Variables

Configure in GitHub repository settings:

**Environments:**
- `development`
- `staging`
- `production`

**Variables per environment:**
- `KUBE_NAMESPACE`
- `CLUSTER_NAME`
- Any app-specific config

## Best Practices

### 1. Branching Strategy

```
main (production)
  ↑
develop (staging)
  ↑
feature/xyz (development)
```

### 2. Commit Messages

Follow Conventional Commits:
```bash
feat(order-food): add payment processing
fix(database): resolve connection timeout
docs: update deployment guide
ci: add security scanning workflow
```

### 3. Versioning

Use Semantic Versioning (SemVer):
- `v1.0.0` - Major.Minor.Patch
- `v1.1.0` - New features (backward compatible)
- `v1.1.1` - Bug fixes
- `v2.0.0` - Breaking changes

### 4. Testing

**Before Pushing:**
```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run ./...

# Run tests
go test -v -race ./...

# Build
go build ./...
```

### 5. Docker Images

**Best Practices:**
- Use specific tags, avoid `latest` in production
- Multi-stage builds for smaller images
- Security scanning before deployment
- Regular base image updates

## Monitoring and Debugging

### View Workflow Runs

```bash
# List workflows
gh workflow list

# View runs
gh run list

# Watch a run
gh run watch

# View logs
gh run view <run-id> --log
```

### Check Build Status

- GitHub repository badges
- Actions tab in GitHub
- Email notifications
- Slack/Discord webhooks (configurable)

### Debug Failed Builds

1. Check workflow logs in GitHub Actions
2. Review error messages
3. Test locally:
   ```bash
   # Reproduce build
   docker build -t test ./order-food

   # Run tests
   cd order-food && go test -v ./...
   ```

## CI/CD Metrics

The workflows provide:
- Build duration
- Test coverage
- Code quality scores
- Security vulnerabilities
- Deployment success rate

## Extending the Pipeline

### Add a New Workflow

1. Create `.github/workflows/my-workflow.yml`
2. Define triggers and jobs
3. Commit and push
4. Verify in Actions tab

### Add a New Module

1. Update path filters in `ci.yml`
2. Add build job for the module
3. Update Docker workflow
4. Create Helm chart
5. Update deployment workflow

### Add Environment

1. Go to Settings → Environments
2. Create new environment
3. Configure protection rules
4. Add environment secrets
5. Update deployment workflow

## Troubleshooting

### Workflow Not Triggering

- Check branch name matches trigger pattern
- Verify file paths in path filters
- Check if workflow file has syntax errors

### Docker Push Fails

- Verify GITHUB_TOKEN permissions
- Check container registry settings
- Ensure package write permission enabled

### Deployment Fails

- Verify Kubernetes credentials
- Check namespace exists
- Review pod logs
- Verify image pull secrets

### Tests Failing

- Check Go version compatibility
- Verify dependencies are up to date
- Review test logs
- Run tests locally

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Helm Chart Testing](https://github.com/helm/chart-testing)
- [golangci-lint](https://golangci-lint.run/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

## Support

For issues with CI/CD:
1. Check workflow logs
2. Review this documentation
3. Test locally
4. Create an issue with workflow run link
