#!/usr/bin/env bash
set -euo pipefail

# Check if core version has been incremented and needs release
# Usage: ./check-core-version-increment.sh

CURRENT_VERSION=$(cat core/version)
TAG_NAME="core/v${CURRENT_VERSION}"

echo "📋 Current core version: $CURRENT_VERSION"
echo "🏷️ Expected tag: $TAG_NAME"

# Check if tag already exists
if git rev-parse --verify "$TAG_NAME" >/dev/null 2>&1; then
  echo "⚠️ Tag $TAG_NAME already exists"
  {
      echo "should-release=false"
      echo "new-version=$CURRENT_VERSION"
      echo "tag-exists=true"
  } >> "$GITHUB_OUTPUT"
  exit 0
fi

# Get previous version from git tags
LATEST_CORE_TAG=$(git tag -l "core/v*" | sort -V | tail -1)

if [ -z "$LATEST_CORE_TAG" ]; then
  echo "📦 No existing core tags found, this will be the first release"
  {
      echo "should-release=true"
      echo "new-version=$CURRENT_VERSION"
      echo "tag-exists=false"
  } >> "$GITHUB_OUTPUT"
  exit 0
fi

PREVIOUS_VERSION=${LATEST_CORE_TAG#core/v}
echo "📋 Previous core version: $PREVIOUS_VERSION"

# Compare versions using sort -V (version sort)
if [ "$(printf '%s\n' "$PREVIOUS_VERSION" "$CURRENT_VERSION" | sort -V | tail -1)" = "$CURRENT_VERSION" ] && [ "$PREVIOUS_VERSION" != "$CURRENT_VERSION" ]; then
  echo "✅ Version incremented from $PREVIOUS_VERSION to $CURRENT_VERSION"
  echo "🚀 Core release needed"
  {
      echo "should-release=true"
      echo "new-version=$CURRENT_VERSION"
      echo "tag-exists=false"
  } >> "$GITHUB_OUTPUT"
else
  echo "⏭️ No version increment detected (current: $CURRENT_VERSION, latest: $PREVIOUS_VERSION)"
  {
      echo "should-release=false"
      echo "new-version=$CURRENT_VERSION"
      echo "tag-exists=false"
  } >> "$GITHUB_OUTPUT"
fi
