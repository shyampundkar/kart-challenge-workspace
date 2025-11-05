# GitHub Actions Workflows

This directory contains all CI/CD workflows for the kart-challenge-workspace project.

## Workflows Overview

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| CI Pipeline | `ci.yml` | Push, PR | Build, test, and validate code |
| Docker Build | `docker.yml` | Push to main, Tags | Build and push container images |
| Helm Validation | `helm.yml` | Helm changes | Validate and test Helm charts |
| Pull Request | `pr.yml` | PR events | PR-specific validations |
| Release | `release.yml` | Version tags | Create releases with assets |
| Security Scan | `codeql.yml` | Push, PR, Schedule | Security vulnerability scanning |
| Deploy | `deploy.yml` | Manual | Deploy to environments |

## Quick Reference

### Status Badges

Add these to your README:

```markdown
![CI](https://github.com/<username>/<repo>/actions/workflows/ci.yml/badge.svg)
![Docker](https://github.com/<username>/<repo>/actions/workflows/docker.yml/badge.svg)
![Helm](https://github.com/<username>/<repo>/actions/workflows/helm.yml/badge.svg)
![Security](https://github.com/<username>/<repo>/actions/workflows/codeql.yml/badge.svg)
```

### Running Workflows Locally

#### Test with Act

```bash
# Install act
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run CI workflow
act push

# Run specific job
act -j build-order-food

# List workflows
act -l
```

#### Test Docker Builds

```bash
# Test database-migration
docker build -t test-db-migration ./database-migration

# Test database-load
docker build -t test-db-load ./database-load

# Test order-food
docker build -t test-order-food ./order-food
```

#### Test Helm Charts

```bash
# Lint charts
helm lint database-migration/helm
helm lint database-load/helm
helm lint order-food/helm

# Template charts
helm template test database-migration/helm
helm template test database-load/helm
helm template test order-food/helm

# Test in Kind cluster
kind create cluster --name test
kind load docker-image test-order-food --name test
helm install test-release ./order-food/helm --set image.pullPolicy=Never
```

## Workflow Details

### CI Pipeline (`ci.yml`)

**Key Features:**
- Path-based change detection
- Parallel module builds
- Code coverage reporting
- Security scanning
- Linting with golangci-lint

**Runs on:**
- Every push to main, develop, feature branches
- All pull requests

**Duration:** ~3-5 minutes per module

### Docker Build (`docker.yml`)

**Key Features:**
- Multi-arch builds (amd64, arm64)
- Automatic semantic versioning
- Layer caching for faster builds
- Trivy security scanning
- Image testing

**Registry:** GitHub Container Registry (ghcr.io)

**Tags Generated:**
- `latest` (main branch)
- `v1.0.0` (version tags)
- `v1.0`, `v1` (semantic versions)
- `main-abc1234` (branch + SHA)
- `pr-123` (pull requests)

**Duration:** ~5-8 minutes per image

### Helm Validation (`helm.yml`)

**Key Features:**
- Chart linting with helm lint
- Chart testing with ct
- Manifest validation with kubeval
- Integration testing in Kind
- Chart packaging

**Runs on:** Changes to helm directories

**Duration:** ~8-12 minutes

### Pull Request (`pr.yml`)

**Key Features:**
- Semantic PR title validation
- Merge conflict detection
- Automated code review
- Size labeling
- Summary comments

**Runs on:** PR events

**Duration:** ~1-2 minutes

### Release (`release.yml`)

**Key Features:**
- Automated changelog generation
- Cross-platform binary builds
- Docker image releases
- Helm chart packaging
- GitHub release creation

**Trigger:** Push tag `v*.*.*`

**Example:**
```bash
git tag v1.0.0
git push origin v1.0.0
```

**Duration:** ~15-20 minutes

**Artifacts:**
- Binaries: Linux, macOS, Windows (amd64, arm64)
- Docker images: All modules, all platforms
- Helm charts: Packaged and indexed

### Security Scan (`codeql.yml`)

**Key Features:**
- CodeQL analysis
- Gosec security scanner
- Dependency review
- SARIF report upload

**Runs on:**
- Push to main/develop
- Pull requests
- Weekly schedule (Mondays)

**Duration:** ~10-15 minutes

### Deploy (`deploy.yml`)

**Key Features:**
- Manual workflow dispatch
- Environment selection
- Version/tag selection
- Automated smoke tests
- Rollback on failure

**Trigger:** Manual (workflow_dispatch)

**Environments:**
- development
- staging
- production

**Usage:**
```bash
# Via GitHub CLI
gh workflow run deploy.yml \
  -f environment=staging \
  -f version=v1.0.0

# Via GitHub UI
Actions → Deploy to Kubernetes → Run workflow
```

**Duration:** ~5-8 minutes

## Environment Setup

### Required Secrets

Configure in: Settings → Secrets and variables → Actions

**Automatically Available:**
- `GITHUB_TOKEN` - GitHub Actions token (auto-generated)

**Optional (for cloud deployments):**
- `KUBE_CONFIG` - Kubernetes configuration
- `DOCKER_USERNAME` - DockerHub username
- `DOCKER_PASSWORD` - DockerHub password
- Cloud-specific credentials (AWS, GCP, Azure)

### Environment Configuration

Configure in: Settings → Environments

Create environments:
- `development`
- `staging`
- `production`

For each environment:
1. Add environment-specific secrets
2. Configure protection rules
3. Set deployment branches
4. Add required reviewers (for production)

## Monitoring

### View Workflow Status

**GitHub UI:**
- Repository → Actions tab
- Click workflow name
- View runs and logs

**GitHub CLI:**
```bash
# List workflows
gh workflow list

# View runs
gh run list --workflow=ci.yml

# Watch live
gh run watch

# View logs
gh run view <run-id> --log
```

### Notifications

Configure in: Settings → Notifications

Options:
- Email notifications
- Slack/Discord webhooks
- Status checks in PRs

## Optimization Tips

### Speed Up Builds

1. **Use caching:**
   - Go module cache
   - Docker layer cache
   - Build artifact cache

2. **Parallelize jobs:**
   - Run independent modules concurrently
   - Use matrix strategy for multiple versions

3. **Optimize Docker:**
   - Multi-stage builds
   - Layer ordering
   - .dockerignore file

### Reduce Costs

1. **Path filters:**
   - Only build changed modules
   - Skip unnecessary workflows

2. **Conditional jobs:**
   - Use `if` conditions
   - Skip jobs on draft PRs

3. **Artifact retention:**
   - Reduce retention days
   - Clean up old artifacts

## Troubleshooting

### Common Issues

**Workflow not running:**
- Check branch name matches trigger
- Verify workflow file syntax
- Check repository permissions

**Build fails:**
- Review error logs
- Test locally
- Check dependency versions
- Verify environment variables

**Docker push fails:**
- Check GITHUB_TOKEN permissions
- Enable package write permission
- Verify registry authentication

**Helm test fails:**
- Check chart syntax
- Verify template output
- Test in local cluster

### Debug Commands

```bash
# Validate workflow syntax
gh workflow view ci.yml

# Download artifacts
gh run download <run-id>

# Re-run failed jobs
gh run rerun <run-id> --failed

# Cancel running workflow
gh run cancel <run-id>
```

## Best Practices

1. **Keep workflows DRY:**
   - Use reusable workflows
   - Share common steps
   - Use composite actions

2. **Security:**
   - Use minimal permissions
   - Pin action versions
   - Scan for vulnerabilities
   - Never commit secrets

3. **Testing:**
   - Test workflows locally
   - Use draft PRs for testing
   - Validate before merging

4. **Documentation:**
   - Comment complex steps
   - Update this README
   - Document secret requirements

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Action Marketplace](https://github.com/marketplace?type=actions)
- [CI/CD Guide](../CI_CD_GUIDE.md)

## Contributing

When adding or modifying workflows:

1. Test locally with `act`
2. Use descriptive job/step names
3. Add comments for complex logic
4. Update this README
5. Test in draft PR first
6. Get review before merging

## Support

For issues with workflows:
1. Check workflow logs
2. Review this documentation
3. Search GitHub Actions issues
4. Ask in team chat
5. Create repository issue
