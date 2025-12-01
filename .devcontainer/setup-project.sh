#!/bin/bash
# Setup script for Michishirube development environment

echo "🏗️ Setting up Michishirube development environment..."

# Download Go dependencies
echo "📦 Downloading Go dependencies..."
go mod download

# Generate code if needed
if [ -f "Makefile" ]; then
    echo "🔧 Running make generate..."
    make generate || echo "⚠️ make generate failed, but continuing..."

    echo "📚 Generating API documentation..."
    make docs || echo "⚠️ make docs failed, but continuing..."
fi

# Create data directory for SQLite database
echo "🗄️ Setting up database directory..."
mkdir -p /workspace/data
chmod 755 /workspace/data

# Set up git config for Claude Code commits (with user's signature preference)
echo "📝 Setting up git configuration..."
git config --global commit.gpgsign false
git config --global user.name "${GIT_AUTHOR_NAME:-Claude Code Dev}"
git config --global user.email "${GIT_AUTHOR_EMAIL:-dev@example.com}"

# Ensure proper permissions for the workspace
echo "🔐 Setting up permissions..."
sudo chown -R developer:developer /workspace || echo "⚠️ Permission setup failed, but continuing..."

# Show helpful information
echo ""
echo "✅ Development environment ready!"
echo ""
echo "🚀 Quick start commands:"
echo "  make build          # Build the application"
echo "  make run            # Run in development mode"
echo "  make test           # Run tests"
echo "  make docs           # Generate API documentation"
echo ""
echo "🌐 Application will be available at:"
echo "  http://localhost:8080        # Main application"
echo "  http://localhost:8080/docs   # API documentation"
echo ""
echo "🧪 Development environment ready!"
echo "  Note: Claude Code extension is available in Cursor/VS Code"
echo "  All Go tools are installed and ready to use"
echo ""