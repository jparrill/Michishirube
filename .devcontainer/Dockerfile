# Complete Claude Code Development Container for Michishirube
# Full-featured development environment with all tools

FROM golang:1.24-bookworm

# Arguments
ARG TZ=America/Los_Angeles
ARG GIT_DELTA_VERSION=0.18.2

# Set timezone
ENV TZ=$TZ
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Install all system dependencies and development tools
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y --no-install-recommends \
        # Base development tools
        git ca-certificates tzdata make wget tar curl \
        # Build tools for your project (matching production Dockerfile)
        gcc libc6-dev libsqlite3-dev \
        gcc-aarch64-linux-gnu libc6-dev-arm64-cross \
        gcc-x86-64-linux-gnu libc6-dev-amd64-cross \
        # Development and productivity tools
        openssh-client gnupg2 \
        iproute2 procps lsof htop net-tools psmisc \
        rsync unzip zip nano vim less jq sudo zsh \
        # Network tools for devcontainer security
        iptables ipset dnsutils aggregate python3 \
        # SQLite for database management
        sqlite3 \
        # Build essentials
        build-essential \
    && apt-get autoremove -y && apt-get clean -y

# Install Node.js for Claude Code CLI (when available)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs

# Install git-delta for better diff output (multi-arch support)
RUN ARCH=$(dpkg --print-architecture) \
    && if [ "$ARCH" = "amd64" ]; then GOARCH="amd64"; elif [ "$ARCH" = "arm64" ]; then GOARCH="arm64"; else GOARCH="amd64"; fi \
    && wget -q -O /tmp/git-delta.deb \
    "https://github.com/dandavison/delta/releases/download/${GIT_DELTA_VERSION}/git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb" \
    && dpkg -i /tmp/git-delta.deb \
    && rm /tmp/git-delta.deb

# Install fzf for better command line experience
RUN git clone --depth 1 https://github.com/junegunn/fzf.git /opt/fzf \
    && /opt/fzf/install --all

# Create development user
RUN groupadd -g 1000 developer && \
    useradd -u 1000 -g developer -s /bin/zsh -m developer

# Setup sudo for developer user
RUN echo developer ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/developer \
    && chmod 0440 /etc/sudoers.d/developer

# Copy firewall initialization script
COPY .devcontainer/init-firewall.sh /usr/local/bin/init-firewall.sh
RUN chmod +x /usr/local/bin/init-firewall.sh

# Create Go workspace with proper permissions (as root before switching users)
RUN mkdir -p /go/bin /go/src /go/pkg \
    && chown -R developer:developer /go

# Switch to developer user and setup environment
USER developer
WORKDIR /home/developer

# Install ZSH configuration for the developer user
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended \
    && echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc \
    && echo 'export GOPATH="/go"' >> ~/.bashrc \
    && echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.bashrc \
    && echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.zshrc \
    && echo 'export GOPATH="/go"' >> ~/.zshrc \
    && echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.zshrc

# Install all Go tools for development (pinned versions for reproducibility)
RUN go install golang.org/x/tools/gopls@v0.16.2 \
    && go install github.com/go-delve/delve/cmd/dlv@v1.23.1 \
    && go install honnef.co/go/tools/cmd/staticcheck@2024.1.1 \
    && go install golang.org/x/tools/cmd/goimports@v0.26.0 \
    && go install go.uber.org/mock/mockgen@v0.4.0 \
    && go install github.com/swaggo/swag/cmd/swag@v1.16.4 \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

# Create Claude directory
RUN mkdir -p /home/developer/.claude

# Set working directory
WORKDIR /workspace

# Default command
CMD ["/bin/zsh"]