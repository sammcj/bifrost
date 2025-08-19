#!/usr/bin/env bash
set -euo pipefail

# ----------- CONFIG -----------

REGISTRY="docker.io"
ACCOUNT="maximhq"
IMAGE_NAME="bifrost"
IMAGE="${REGISTRY}/${ACCOUNT}/${IMAGE_NAME}"
DOCKERFILE="transports/Dockerfile"
CONTEXT_DIR="."
CACHE_DIR=".buildx-cache"
BUILDER_NAME="multiarch-builder-${GITHUB_RUN_ID:-$$}"
PLATFORMS="linux/amd64,linux/arm64"

# ----------- AUTH -----------

DOCKER_USERNAME="${DOCKER_USERNAME:-}"
DOCKER_PASSWORD="${DOCKER_PASSWORD:-}"

if [[ -z "$DOCKER_USERNAME" ]]; then
  if [[ -n "${CI:-}" ]]; then
    echo "❌ DOCKER_USERNAME is required in CI. Set it as a secret/env var."
    exit 1
  fi
  read -rp "🔑 Docker Hub username: " DOCKER_USERNAME
fi
if [[ -z "$DOCKER_PASSWORD" ]]; then
  if [[ -n "${CI:-}" ]]; then
    echo "❌ DOCKER_PASSWORD is required in CI. Set it as a secret/env var."
    exit 1
  fi
  read -rsp "🔐 Docker Hub password: " DOCKER_PASSWORD
  echo
fi

echo "🔐 Logging into Docker Hub..."
echo "$DOCKER_PASSWORD" | docker login --username "$DOCKER_USERNAME" --password-stdin

# ----------- BUILDX SETUP -----------

echo "🔧 Ensuring Buildx is ready..."

# Note: QEMU/binfmt should be configured via GitHub Actions using:
# - docker/setup-qemu-action@v3 (platforms: all)
# - docker/setup-buildx-action@v3

if ! docker buildx version >/dev/null 2>&1; then
  echo "❌ Docker Buildx is not available. Please upgrade Docker."
  exit 1
fi

if ! docker buildx inspect "$BUILDER_NAME" >/dev/null 2>&1; then
  docker buildx create --use --name "$BUILDER_NAME"
else
  docker buildx use "$BUILDER_NAME"
fi

docker buildx inspect --bootstrap

# ----------- VERSION -----------

if [[ -n "${1:-}" ]]; then
  RAW_VERSION="$1"
else
  RAW_VERSION=$(git describe --tags --abbrev=0 --match "transports/v*" 2>/dev/null || true)
  RAW_VERSION="${RAW_VERSION:-transports/v0.0.0}"
fi

VERSION_ONLY="${RAW_VERSION#transports/}"
VERSION_ONLY="${VERSION_ONLY#v}"
VERSION="v${VERSION_ONLY}"

# Check if version contains prerelease identifiers
if [[ "$VERSION_ONLY" =~ (alpha|beta|rc|pre|dev|snapshot) ]]; then
  echo "🔍 Detected prerelease version: ${VERSION} - skipping 'latest' tag"
  TAGS=(
    "${IMAGE}:${VERSION}"
  )
else
  echo "🔍 Detected stable version: ${VERSION} - including 'latest' tag"
  TAGS=(
    "${IMAGE}:${VERSION}"
    "${IMAGE}:latest"
  )
fi

LABELS=(
  "org.opencontainers.image.title=Bifrost LLM Gateway (HTTP)"
  "org.opencontainers.image.description=The fastest LLM gateway written in Go. Learn more here: https://github.com/maximhq/bifrost"
  "org.opencontainers.image.source=https://github.com/maximhq/bifrost"
  "org.opencontainers.image.version=${VERSION}"
  "org.opencontainers.image.created=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
  "org.opencontainers.image.revision=$(git rev-parse HEAD)"
)

# ----------- BUILD -----------

mkdir -p "$CACHE_DIR"

echo "🚀 Building and pushing Docker image: ${IMAGE}:${VERSION}"

BUILD_ARGS=()
CACHE_ARGS=()

for tag in "${TAGS[@]}"; do
  BUILD_ARGS+=(--tag "$tag")
done

for label in "${LABELS[@]}"; do
  BUILD_ARGS+=(--label "$label")
done

# Cache strategy
if [[ -n "${GITHUB_ACTIONS:-}" ]]; then
  CACHE_ARGS=(--cache-from=type=gha --cache-to=type=gha,mode=max)
else
  CACHE_ARGS=(--cache-from="type=local,src=${CACHE_DIR}" --cache-to="type=local,dest=${CACHE_DIR},mode=max")
fi

docker buildx build \
  --platform "$PLATFORMS" \
  --file "$DOCKERFILE" \
  --push \
  --pull \
  --provenance=true \
  --sbom=true \
  "${CACHE_ARGS[@]}" \
  "${BUILD_ARGS[@]}" \
  "$CONTEXT_DIR"


# ----------- CLEANUP -----------

echo "🧼 Cleanup: Pruning Buildx cache (non-destructive)..."
docker buildx prune --force

echo "👋 Logging out of Docker Hub..."
docker logout "$REGISTRY"

echo "✅ Done."
