#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="default"
MINIKUBE_DRIVER="docker"
MINIKUBE_MEMORY="4096"
MINIKUBE_CPUS="2"

# Function to print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Helm on macOS
install_helm_mac() {
    print_info "Installing Helm on macOS..."
    if command_exists brew; then
        brew install helm
    else
        print_info "Homebrew not found. Downloading Helm binary..."
        curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    fi
}

# Function to install Helm on Linux
install_helm_linux() {
    print_info "Installing Helm on Linux..."
    curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
}

# Function to install minikube on macOS
install_minikube_mac() {
    print_info "Installing minikube on macOS..."
    if command_exists brew; then
        brew install minikube
    else
        print_info "Homebrew not found. Downloading minikube binary..."
        curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64
        sudo install minikube-darwin-amd64 /usr/local/bin/minikube
        rm minikube-darwin-amd64
    fi
}

# Function to install minikube on Linux
install_minikube_linux() {
    print_info "Installing minikube on Linux..."
    curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
    sudo install minikube-linux-amd64 /usr/local/bin/minikube
    rm minikube-linux-amd64
}

# Check and install dependencies
check_dependencies() {
    print_info "Checking dependencies..."

    # Check Docker
    if ! command_exists docker; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    print_success "Docker found"

    # Check if Docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker daemon is not running. Please start Docker."
        exit 1
    fi
    print_success "Docker daemon is running"

    # Check kubectl
    if ! command_exists kubectl; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    print_success "kubectl found"

    # Check Helm
    if ! command_exists helm; then
        print_warning "Helm is not installed. Installing..."
        OS="$(uname -s)"
        case "${OS}" in
            Linux*)
                install_helm_linux
                ;;
            Darwin*)
                install_helm_mac
                ;;
            *)
                print_error "Unsupported operating system: ${OS}"
                exit 1
                ;;
        esac
        print_success "Helm installed successfully"
    else
        print_success "Helm found"
    fi

    # Check minikube
    if ! command_exists minikube; then
        print_warning "minikube is not installed. Installing..."
        OS="$(uname -s)"
        case "${OS}" in
            Linux*)
                install_minikube_linux
                ;;
            Darwin*)
                install_minikube_mac
                ;;
            *)
                print_error "Unsupported operating system: ${OS}"
                exit 1
                ;;
        esac
        print_success "minikube installed successfully"
    else
        print_success "minikube found"
    fi
}

# Start minikube if not running
start_minikube() {
    print_info "Checking minikube status..."

    if minikube status >/dev/null 2>&1; then
        print_success "minikube is already running"
    else
        print_info "Starting minikube with driver: ${MINIKUBE_DRIVER}, memory: ${MINIKUBE_MEMORY}MB, cpus: ${MINIKUBE_CPUS}"
        minikube start --driver="${MINIKUBE_DRIVER}" --memory="${MINIKUBE_MEMORY}" --cpus="${MINIKUBE_CPUS}"
        print_success "minikube started successfully"
    fi

    # Wait for minikube to be ready
    print_info "Waiting for minikube to be ready..."
    kubectl wait --for=condition=Ready nodes --all --timeout=300s
    print_success "minikube is ready"
}

# Build Docker images
build_images() {
    print_info "Building Docker images..."

    # Set Docker environment to use minikube's Docker daemon
    eval $(minikube docker-env)

    # Fix Docker credential helper issue by temporarily disabling credsStore
    # This is needed when Docker Desktop credential helper is not available in minikube's Docker daemon
    export DOCKER_CONFIG=$(mktemp -d)
    cat > "${DOCKER_CONFIG}/config.json" <<EOF
{
  "auths": {}
}
EOF
    print_info "Using temporary Docker config to avoid credential helper issues"

    # Build database-migration
    print_info "Building database-migration image..."
    docker build -t database-migration:latest ./database-migration
    print_success "database-migration image built"

    # Build database-load
    print_info "Building database-load image..."
    docker build -t database-load:latest ./database-load
    print_success "database-load image built"

    # Build order-food
    print_info "Building order-food image..."
    docker build -t order-food:latest ./order-food
    print_success "order-food image built"

    # Clean up temporary Docker config
    rm -rf "${DOCKER_CONFIG}"
    unset DOCKER_CONFIG

    print_success "All images built successfully"
}

# Deploy using Helm
deploy_with_helm() {
    print_info "Deploying applications with Helm..."

    # Deploy PostgreSQL database
    print_info "Deploying PostgreSQL database..."
    helm upgrade --install postgres ./postgres/helm \
        --namespace "${NAMESPACE}" \
        --wait \
        --timeout 5m
    print_success "PostgreSQL database deployed"

    # Wait for PostgreSQL to be ready
    print_info "Waiting for PostgreSQL to be ready..."
    kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=postgres -n "${NAMESPACE}" --timeout=300s
    print_success "PostgreSQL is ready"

    # Deploy database-load CronJob for periodic data refresh
    print_info "Deploying database-load CronJob..."
    helm upgrade --install database-load ./database-load/helm \
        --namespace "${NAMESPACE}" \
        --set image.pullPolicy=Never \
        --set job.enabled=false \
        --set cronjob.enabled=true \
        --wait \
        --timeout 5m
    print_success "database-load CronJob deployed (runs periodically, failures won't impact order-food)"

    # Deploy order-food (with database-migration and database-load as init containers)
    print_info "Deploying order-food..."
    print_info "Init containers will run in sequence: database-migration -> database-load"
    helm upgrade --install order-food ./order-food/helm \
        --namespace "${NAMESPACE}" \
        --set image.pullPolicy=Never \
        --set initContainers.databaseMigration.image.pullPolicy=Never \
        --set initContainers.databaseLoad.image.pullPolicy=Never \
        --wait \
        --timeout 10m
    print_success "order-food deployed (database-migration and database-load ran as init containers)"

    print_success "All applications deployed successfully"
}

# Verify deployments
verify_deployments() {
    print_info "Verifying deployments..."

    echo ""
    print_info "=== Checking Deployments ==="
    kubectl get deployments -n "${NAMESPACE}"

    echo ""
    print_info "=== Checking CronJobs ==="
    kubectl get cronjobs -n "${NAMESPACE}" | grep database-load || true

    echo ""
    print_info "=== Checking Pods ==="
    kubectl get pods -n "${NAMESPACE}"

    echo ""
    print_info "=== Checking Services ==="
    kubectl get services -n "${NAMESPACE}"

    echo ""
    print_info "=== Init Container Status ==="
    kubectl get pods -n "${NAMESPACE}" -l app.kubernetes.io/name=order-food -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{range .status.initContainerStatuses[*]}  {"- "}{.name}{": "}{.state}{"\n"}{end}{end}' || echo "No init containers found"
}

# Display access information
display_access_info() {
    echo ""
    print_success "===== Deployment Complete ====="
    echo ""
    print_info "To access the order-food service, run:"
    echo "  kubectl port-forward -n ${NAMESPACE} svc/order-food 8080:80"
    echo ""
    print_info "Then access the service at: http://localhost:8080"
    echo ""
    print_info "To access PostgreSQL database, run:"
    echo "  kubectl port-forward -n ${NAMESPACE} svc/postgres 5432:5432"
    echo "  psql -h localhost -U postgres -d orderfood"
    echo ""
    print_info "To view logs:"
    echo "  PostgreSQL:                          kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=postgres"
    echo "  Database Migration (init container): kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=order-food -c database-migration"
    echo "  Database Load (init container):      kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=order-food -c database-load"
    echo "  Order Food (main container):         kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=order-food -c order-food"
    echo "  Database Load (CronJob):             kubectl logs -n ${NAMESPACE} -l app.kubernetes.io/name=database-load"
    echo ""
    print_info "To check CronJob:"
    echo "  View CronJob:          kubectl get cronjobs -n ${NAMESPACE}"
    echo "  View CronJob history:  kubectl get jobs -n ${NAMESPACE} -l app.kubernetes.io/name=database-load"
    echo "  Trigger manual run:    kubectl create job --from=cronjob/database-load manual-load-\$(date +%s) -n ${NAMESPACE}"
    echo ""
    print_info "To check database schema:"
    echo "  kubectl exec -it -n ${NAMESPACE} deployment/postgres -- psql -U postgres -d orderfood -c '\\dt'"
    echo ""
    print_info "To uninstall:"
    echo "  helm uninstall postgres database-load order-food -n ${NAMESPACE}"
    echo ""
    print_info "To stop minikube:"
    echo "  minikube stop"
    echo ""
    print_info "To delete minikube cluster:"
    echo "  minikube delete"
    echo ""
}

# Cleanup function
cleanup() {
    if [ "$?" -ne 0 ]; then
        print_error "Deployment failed. Check the logs above for details."
        echo ""
        print_info "To debug, you can run:"
        echo "  kubectl get pods -n ${NAMESPACE}"
        echo "  kubectl describe pod <pod-name> -n ${NAMESPACE}"
        echo "  kubectl logs <pod-name> -n ${NAMESPACE}"
    fi
}

trap cleanup EXIT

# Main execution
main() {
    print_info "===== Starting Deployment to Minikube ====="
    echo ""

    check_dependencies
    echo ""

    start_minikube
    echo ""

    build_images
    echo ""

    deploy_with_helm
    echo ""

    verify_deployments
    echo ""

    display_access_info
}

# Run main function
main
