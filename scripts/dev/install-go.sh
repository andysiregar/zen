#!/bin/bash

# Install Go 1.21 on Ubuntu/Debian
set -e

GO_VERSION="1.21.5"
GO_OS="linux"
GO_ARCH="amd64"

echo "🚀 Installing Go $GO_VERSION..."

# Remove old Go installation
sudo rm -rf /usr/local/go

# Download Go
wget -q https://golang.org/dl/go${GO_VERSION}.${GO_OS}-${GO_ARCH}.tar.gz -O /tmp/go${GO_VERSION}.tar.gz

# Extract to /usr/local
sudo tar -C /usr/local -xzf /tmp/go${GO_VERSION}.tar.gz

# Cleanup
rm /tmp/go${GO_VERSION}.tar.gz

# Add to PATH if not already added
if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo 'export GOPATH=$HOME/go' >> ~/.bashrc
    echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
fi

# Source the profile
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go

echo "✅ Go installed successfully!"
echo "🔄 Please run 'source ~/.bashrc' or restart your terminal"
echo "📝 Verify installation with: go version"

/usr/local/go/bin/go version