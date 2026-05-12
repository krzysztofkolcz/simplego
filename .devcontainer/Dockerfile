FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm

# ----------------------------------------------------
# Versions
# ----------------------------------------------------

ENV GO_VERSION=1.25.0

ENV KUBECTL_VERSION=v1.31.0
ENV HELM_VERSION=v3.17.3
ENV K9S_VERSION=v0.50.9

ENV SQLC_VERSION=v1.29.0
ENV MIGRATE_VERSION=v4.18.3

# ----------------------------------------------------
# Remove problematic yarn repo
# ----------------------------------------------------

RUN rm -f /etc/apt/sources.list.d/yarn.list

# ----------------------------------------------------
# Basic packages
# ----------------------------------------------------

RUN apt-get update && apt-get install -y \
    curl \
    wget \
    unzip \
    jq \
    vim \
    git \
    make \
    ca-certificates \
    gnupg \
    lsb-release \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

# ----------------------------------------------------
# Install Go 1.25
# ----------------------------------------------------

RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \
    rm go${GO_VERSION}.linux-amd64.tar.gz

# ----------------------------------------------------
# kubectl
# ----------------------------------------------------

RUN curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \
    rm kubectl

# ----------------------------------------------------
# Helm
# ----------------------------------------------------

RUN curl -LO https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    tar -zxvf helm-${HELM_VERSION}-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/helm && \
    rm -rf linux-amd64 helm-${HELM_VERSION}-linux-amd64.tar.gz

# ----------------------------------------------------
# k9s
# ----------------------------------------------------

RUN wget https://github.com/derailed/k9s/releases/download/${K9S_VERSION}/k9s_linux_amd64.deb && \
    apt install -y ./k9s_linux_amd64.deb && \
    rm k9s_linux_amd64.deb

# ----------------------------------------------------
# sqlc
# ----------------------------------------------------

RUN wget https://github.com/sqlc-dev/sqlc/releases/download/${SQLC_VERSION}/sqlc_1.29.0_linux_amd64.tar.gz && \
    tar -xzf sqlc_1.29.0_linux_amd64.tar.gz && \
    mv sqlc /usr/local/bin/sqlc && \
    chmod +x /usr/local/bin/sqlc && \
    rm sqlc_1.29.0_linux_amd64.tar.gz

# ----------------------------------------------------
# golang-migrate
# ----------------------------------------------------

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/

# ----------------------------------------------------
# Environment
# ----------------------------------------------------

ENV GOTOOLCHAIN=auto

ENV PATH="/usr/local/go/bin:/go/bin:${PATH}"

# ----------------------------------------------------
# Permissions
# ----------------------------------------------------

RUN chown -R vscode:vscode /go