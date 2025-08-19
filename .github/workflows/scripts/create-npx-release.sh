#!/usr/bin/env bash
set -euo pipefail

# Create GitHub release for NPX package
# Usage: ./create-npx-release.sh <version> <full-tag>

VERSION="$1"
FULL_TAG="$2"

if [[ -z "$VERSION" || -z "$FULL_TAG" ]]; then
  echo "❌ Usage: $0 <version> <full-tag>"
  exit 1
fi
# Mark prereleases when version contains a hyphen
PRERELEASE_FLAG=""
if [[ "$VERSION" == *-* ]]; then
  PRERELEASE_FLAG="--prerelease"
fi
TITLE="NPX Package v$VERSION"

# Create release body
BODY="## NPX Package Release

### 📦 NPX Package v$VERSION

The Bifrost CLI is now available on npm!

### Installation

\`\`\`bash
# Install globally
npm install -g @maximhq/bifrost

# Or use with npx (no installation needed)
npx @maximhq/bifrost --help
\`\`\`

### Usage

\`\`\`bash
# Start Bifrost HTTP server
bifrost

# Use specific transport version
bifrost --transport-version v1.2.3

# Get help
bifrost --help
\`\`\`

### Links

- 📦 [View on npm](https://www.npmjs.com/package/@maximhq/bifrost)
- 📚 [Documentation](https://github.com/maximhq/bifrost)
- 🐛 [Report Issues](https://github.com/maximhq/bifrost/issues)

### What's New

This NPX package provides a convenient way to run Bifrost without manual binary downloads. The CLI automatically:

- Detects your platform and architecture
- Downloads the appropriate binary
- Supports version pinning with \`--transport-version\`
- Provides progress indicators for downloads

---
_This release was automatically created from tag \`$FULL_TAG\`_"

# Create release
echo "🎉 Creating GitHub release for $TITLE..."
if gh release view "$FULL_TAG" >/dev/null 2>&1; then
  echo "ℹ️ Release $FULL_TAG already exists. Skipping creation."
  exit 0
fi
gh release create "$FULL_TAG" \
  --title "$TITLE" \
  --notes "$BODY" \
  --latest=false \
  --verify-tag \
  ${PRERELEASE_FLAG}
