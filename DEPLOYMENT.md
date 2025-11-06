# Deployment Guide

This guide explains how to build and deploy all three modules (database-migration, database-load, and order-food) to a Minikube environment.

## Prerequisites

The deployment script will check for these dependencies and automatically install what's missing:

- **Docker** - Container runtime (required - must be installed manually)
- **kubectl** - Kubernetes CLI (required - must be installed manually)
- **Helm** - Kubernetes package manager (will be installed automatically if missing)
- **minikube** - Local Kubernetes cluster (will be installed automatically if missing)

### Installing Prerequisites (if needed)

#### macOS
```bash
# Install Docker Desktop from https://www.docker.com/products/docker-desktop

# Install kubectl
brew install kubectl

# Install Helm
brew install helm
```

#### Linux
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## Quick Start

### Deploy Everything

Simply run the deployment script:

```bash
./deploy.sh
```

This script will:
1. Check if all required tools are installed
2. Install minikube if not present
3. Start minikube if not running
4. Build Docker images for all three modules
5. Deploy all modules using Helm
6. Display verification and access information

### Access the order-food Service

After deployment, access the service:

```bash
kubectl port-forward -n default svc/order-food 8080:80
```

Then open your browser to: http://localhost:8080

### Cleanup

To remove all deployments:

```bash
./cleanup.sh
```

To stop minikube:

```bash
minikube stop
```

To completely remove minikube:

```bash
minikube delete
```

## Modules Overview

### 1. database-migration
- **Type**: Init Container #1 (runs inside order-food pod)
- **Purpose**: Runs database schema migrations
- **Deployment**: Runs first as init container, must complete successfully before next init container

### 2. database-load
- **Type**: Init Container #2 (runs inside order-food pod)
- **Purpose**: Loads initial data into the database
- **Deployment**: Runs second as init container, must complete successfully before main container starts

### 3. order-food
- **Type**: Kubernetes Deployment
- **Purpose**: Main application service
- **Deployment**: Long-running service with health checks
- **Init Containers**:
  1. database-migration (runs first)
  2. database-load (runs second)

## Configuration

### Minikube Settings

Edit these variables in `deploy.sh` to customize minikube:

```bash
MINIKUBE_DRIVER="docker"    # Options: docker, virtualbox, hyperkit
MINIKUBE_MEMORY="4096"      # Memory in MB
MINIKUBE_CPUS="2"           # Number of CPUs
```

### Application Settings

Customize Helm values by creating a custom values file:

```bash
# Create custom values
cat > custom-values.yaml <<EOF
replicaCount: 2
resources:
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi
EOF

# Deploy with custom values
helm upgrade --install order-food ./order-food/helm -f custom-values.yaml
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n default
```

### View Logs

```bash
# Database Migration logs (init container #1)
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-migration

# Database Load logs (init container #2)
kubectl logs -n default -l app.kubernetes.io/name=order-food -c database-load

# Order Food logs (main container)
kubectl logs -n default -l app.kubernetes.io/name=order-food -c order-food

# Follow logs in real-time
kubectl logs -n default -l app.kubernetes.io/name=order-food -c order-food -f
```

### Describe a Pod

```bash
kubectl describe pod <pod-name> -n default
```

### Check Helm Releases

```bash
helm list -n default
```

### Access Minikube Dashboard

```bash
minikube dashboard
```

### Common Issues

#### Docker daemon not running
```bash
# macOS - Start Docker Desktop from Applications
# Linux
sudo systemctl start docker
```

#### Minikube won't start
```bash
# Delete and recreate
minikube delete
minikube start
```

#### Images not found
The script sets `imagePullPolicy: Never` to use locally built images. If you see ImagePullBackOff:
```bash
# Rebuild images
eval $(minikube docker-env)
docker build -t database-migration:latest ./database-migration
docker build -t database-load:latest ./database-load
docker build -t order-food:latest ./order-food
```

## Manual Deployment Steps

If you prefer to deploy manually:

### 1. Start Minikube
```bash
minikube start --driver=docker --memory=4096 --cpus=2
```

### 2. Build Images
```bash
eval $(minikube docker-env)
docker build -t database-migration:latest ./database-migration
docker build -t database-load:latest ./database-load
docker build -t order-food:latest ./order-food
```

### 3. Deploy with Helm
```bash
# Deploy order-food (database-migration and database-load will run as init containers)
helm upgrade --install order-food ./order-food/helm \
  --set image.pullPolicy=Never \
  --set initContainers.databaseMigration.image.pullPolicy=Never \
  --set initContainers.databaseLoad.image.pullPolicy=Never
```

### 4. Verify
```bash
kubectl get all
```

## Directory Structure

```
.
├── deploy.sh                    # Main deployment script
├── cleanup.sh                   # Cleanup script
├── DEPLOYMENT.md               # This file
├── database-migration/
│   ├── Dockerfile
│   ├── helm/
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   └── cmd/main.go
├── database-load/
│   ├── Dockerfile
│   ├── helm/
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   └── cmd/main.go
└── order-food/
    ├── Dockerfile
    ├── helm/
    │   ├── Chart.yaml
    │   ├── values.yaml
    │   └── templates/
    └── cmd/main.go
```

## Next Steps

1. Customize the Helm values files for your environment
2. Add environment-specific configuration (dev, staging, prod)
3. Set up CI/CD pipelines for automated deployments
4. Configure monitoring and logging
5. Add database persistence with PersistentVolumes
6. Configure Ingress for external access

## Support

For issues or questions:
1. Check the logs using the commands above
2. Review the troubleshooting section
3. Verify all prerequisites are installed correctly
