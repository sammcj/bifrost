#!/usr/bin/env bash
set -euo pipefail

# Install cross-compilation toolchains for Go + CGO
# Usage: ./install-cross-compilers.sh

echo "📦 Installing cross-compilation toolchains for Go + CGO..."

sudo apt-get update
sudo apt-get install -y \
  gcc-x86-64-linux-gnu \
  gcc-aarch64-linux-gnu \
  gcc-mingw-w64-x86-64 \  
  musl-tools \
  clang

# Create symbolic links for musl compilers
sudo ln -sf /usr/bin/x86_64-linux-gnu-gcc /usr/local/bin/x86_64-linux-musl-gcc
sudo ln -sf /usr/bin/x86_64-linux-gnu-g++ /usr/local/bin/x86_64-linux-musl-g++
sudo ln -sf /usr/bin/aarch64-linux-gnu-gcc /usr/local/bin/aarch64-linux-musl-gcc
sudo ln -sf /usr/bin/aarch64-linux-gnu-g++ /usr/local/bin/aarch64-linux-musl-g++

# For Darwin, we can use the system clang with appropriate flags
# No need for osxcross for simple CGO cross-compilation
echo "🍎 Setting up Darwin cross-compilation..."

# Create wrapper scripts for Darwin cross-compilation
sudo tee /usr/local/bin/o64-clang > /dev/null << 'EOF'
#!/bin/bash
exec clang -target x86_64-apple-darwin "$@"
EOF

sudo tee /usr/local/bin/o64-clang++ > /dev/null << 'EOF'
#!/bin/bash
exec clang++ -target x86_64-apple-darwin "$@"
EOF

sudo tee /usr/local/bin/oa64-clang > /dev/null << 'EOF'
#!/bin/bash
exec clang -target arm64-apple-darwin "$@"
EOF

sudo tee /usr/local/bin/oa64-clang++ > /dev/null << 'EOF'
#!/bin/bash
exec clang++ -target arm64-apple-darwin "$@"
EOF

# Make wrapper scripts executable
sudo chmod +x /usr/local/bin/o64-clang
sudo chmod +x /usr/local/bin/o64-clang++
sudo chmod +x /usr/local/bin/oa64-clang
sudo chmod +x /usr/local/bin/oa64-clang++

echo "✅ Cross-compilation toolchains installed"
echo ""
echo "Available cross-compilers:"
echo "  Linux amd64:   x86_64-linux-musl-gcc, x86_64-linux-musl-g++"
echo "  Linux arm64:   aarch64-linux-musl-gcc, aarch64-linux-musl-g++"
echo "  Windows amd64: x86_64-w64-mingw32-gcc, x86_64-w64-mingw32-g++"
echo "  Windows arm64: aarch64-w64-mingw32-gcc, aarch64-w64-mingw32-g++"
echo "  Darwin amd64:  o64-clang, o64-clang++"
echo "  Darwin arm64:  oa64-clang, oa64-clang++"