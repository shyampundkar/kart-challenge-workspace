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

# Main cleanup function
cleanup_deployments() {
    print_info "===== Starting Cleanup ====="
    echo ""

    print_info "Uninstalling Helm releases..."

    # Uninstall order-food
    if helm list -n "${NAMESPACE}" | grep -q "order-food"; then
        print_info "Uninstalling order-food..."
        helm uninstall order-food -n "${NAMESPACE}"
        print_success "order-food uninstalled"
    else
        print_warning "order-food not found"
    fi

    # Uninstall database-load CronJob
    if helm list -n "${NAMESPACE}" | grep -q "database-load"; then
        print_info "Uninstalling database-load CronJob..."
        helm uninstall database-load -n "${NAMESPACE}"
        print_success "database-load CronJob uninstalled"
    else
        print_warning "database-load not found"
    fi

    # Note: database-migration and database-load init container are removed with order-food
    print_info "Note: database-migration and database-load init containers are removed with order-food"

    echo ""
    print_success "All Helm releases uninstalled"

    echo ""
    print_info "Remaining resources in namespace ${NAMESPACE}:"
    kubectl get all -n "${NAMESPACE}" | grep -E "database-migration|database-load|order-food" || print_info "No resources found"

    echo ""
    print_info "Cleanup options:"
    echo "  To stop minikube:   minikube stop"
    echo "  To delete minikube: minikube delete"
}

# Run cleanup
cleanup_deployments
