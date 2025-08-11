#!/usr/bin/env bash
set -euo pipefail

# Extract NPX version from tag
# Usage: ./extract-npx-version.sh

# Extract tag name from ref (prefer GITHUB_REF_NAME, fallback to GITHUB_REF)
# Use an intermediate to avoid set -u errors when both are unset in local runs
RAW_REF="${GITHUB_REF_NAME:-${GITHUB_REF:-}}"
TAG_NAME="${RAW_REF#refs/tags/}"
if [[ -z "${TAG_NAME}" ]]; then
  echo "❌ TAG_NAME is empty. Ensure this runs on a tag ref or set GITHUB_REF_NAME."
  exit 1
fi

echo "📋 Processing tag: ${TAG_NAME}"

# Validate tag format (npx/vX.Y.Z or prerelease like npx/vX.Y.Z-rc.1)
if [[ ! "${TAG_NAME}" =~ ^npx/v[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]]; then  
  echo "❌ Invalid tag format '${TAG_NAME}'. Expected format: npx/vMAJOR.MINOR.PATCH"
  exit 1
fi

# Extract version (remove 'npx/v' prefix to get just the version number)
VERSION="${TAG_NAME#npx/v}"
echo "📦 Extracted NPX version: ${VERSION}"
echo "🏷️ Full tag: ${TAG_NAME}"
# Set outputs (only when running in GitHub Actions)
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "version=${VERSION}"
    echo "full-tag=${TAG_NAME}"
  } >> "$GITHUB_OUTPUT"
else
  echo "::notice::GITHUB_OUTPUT not set; skipping outputs (local run?)"
fi