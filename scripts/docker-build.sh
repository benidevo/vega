#!/bin/bash

# Docker build script for Vega AI production image
set -e

# Default values
IMAGE_NAME="vega-ai"
TAG="latest"
REGISTRY="ghcr.io"
REPO_NAME="${GITHUB_REPOSITORY:-benidevo/vega-ai}"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -t, --tag TAG       Docker image tag (default: latest)"
            echo "  -r, --registry REG  Docker registry (default: ghcr.io)"
            echo "  --push              Push image to registry after build"
            echo "  -h, --help          Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

FULL_IMAGE_NAME="${REGISTRY}/${REPO_NAME}:${TAG}"

echo "Building Docker image: ${FULL_IMAGE_NAME}"
echo "========================================="

# Build the Docker image
docker build -f docker/prod/Dockerfile -t "${FULL_IMAGE_NAME}" .

echo "Build completed successfully!"

# Push if requested
if [[ "${PUSH}" == "true" ]]; then
    echo "Pushing image to registry..."
    docker push "${FULL_IMAGE_NAME}"
    echo "Push completed successfully!"
fi

echo "Image: ${FULL_IMAGE_NAME}"
