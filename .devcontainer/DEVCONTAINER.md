Profesjonalny setup Dev Container dla Go + Kubernetes + AI agents
Zbudowałeś profesjonalne środowisko developerskie:


izolowane od hosta,


reproducible,


kompatybilne z Go 1.25,


gotowe pod:


Go backend,


React,


Kubernetes,


Claude Code,


OpenAI Codex,


sqlc,


DevOps tooling.





Co było głównym problemem?
Autocomplete w Go nie działał poprawnie:


brak receiver methods po r.,


losowe podpowiedzi,


problemy gopls.


Prawdziwe przyczyny:


go.mod wymagał Go 1.25


Devcontainer miał Go 1.24


gopls nie mógł poprawnie załadować workspace


tools były instalowane jako root


VS Code próbował instalować własne latest


broken syntax w workspace rozwalał AST parser



Najważniejsze decyzje architektoniczne
1. Izolacja przez Dev Containers
Zamiast:


instalować wszystko na hoście,


używasz:


izolowanego kontenera developerskiego.


Dzięki temu:


AI agent nie ma pełnego dostępu do hosta,


środowisko jest reproducible,


każdy projekt może mieć inne wersje toolingu.



2. Pinned versions zamiast latest
Zamiast:
@latest
używasz:
@vX.Y.Z
Dlaczego:


Go tooling często łamie kompatybilność,


latest powodował błędy:


gopls


air


golangci-lint




To standard enterprise-grade.

3. Go tools instalowane jako user vscode
Nie instalujesz:


gopls


dlv


goimports


w Dockerfile jako root.
Tylko:


w post-create.sh


jako user vscode.


Dlaczego:


unikasz permission denied,


VS Code poprawnie widzi tools,


działa lepiej z Go extension.



4. Go 1.25 doinstalowany ręcznie
Image:
mcr.microsoft.com/devcontainers/go:1-1.25-bookworm
jeszcze nie istnieje.
Dlatego:


bazujesz na 1.24,


ręcznie instalujesz Go 1.25.



5. VS Code nie aktualizuje tools automatycznie
Dodane:
"go.toolsManagement.autoUpdate": false
Dlaczego:


VS Code nie instaluje losowych latest,


zachowujesz reproducibility.



6. Rozdzielenie odpowiedzialności
Dockerfile
Instaluje:


system packages,


kubectl,


helm,


k9s,


sqlc,


migrate,


Go runtime.


post-create.sh
Instaluje:


IDE tools,


AI CLI tools,


Go developer tools.


To profesjonalny pattern.

Finalny Dockerfile
FROM mcr.microsoft.com/devcontainers/go:1-1.24-bookworm# ----------------------------------------------------# Versions# ----------------------------------------------------ENV GO_VERSION=1.25.0ENV KUBECTL_VERSION=v1.31.0ENV HELM_VERSION=v3.17.3ENV K9S_VERSION=v0.50.9ENV SQLC_VERSION=v1.29.0ENV MIGRATE_VERSION=v4.18.3# ----------------------------------------------------# Remove problematic yarn repo# ----------------------------------------------------RUN rm -f /etc/apt/sources.list.d/yarn.list# ----------------------------------------------------# Basic packages# ----------------------------------------------------RUN apt-get update && apt-get install -y \    curl \    wget \    unzip \    jq \    vim \    git \    make \    ca-certificates \    gnupg \    lsb-release \    postgresql-client \    && rm -rf /var/lib/apt/lists/*# ----------------------------------------------------# Install Go 1.25# ----------------------------------------------------RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \    rm -rf /usr/local/go && \    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz && \    rm go${GO_VERSION}.linux-amd64.tar.gz# ----------------------------------------------------# kubectl# ----------------------------------------------------RUN curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \    install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl && \    rm kubectl# ----------------------------------------------------# Helm# ----------------------------------------------------RUN curl -LO https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz && \    tar -zxvf helm-${HELM_VERSION}-linux-amd64.tar.gz && \    mv linux-amd64/helm /usr/local/bin/helm && \    rm -rf linux-amd64 helm-${HELM_VERSION}-linux-amd64.tar.gz# ----------------------------------------------------# k9s# ----------------------------------------------------RUN wget https://github.com/derailed/k9s/releases/download/${K9S_VERSION}/k9s_linux_amd64.deb && \    apt install -y ./k9s_linux_amd64.deb && \    rm k9s_linux_amd64.deb# ----------------------------------------------------# sqlc# ----------------------------------------------------RUN wget https://github.com/sqlc-dev/sqlc/releases/download/${SQLC_VERSION}/sqlc_1.29.0_linux_amd64.tar.gz && \    tar -xzf sqlc_1.29.0_linux_amd64.tar.gz && \    mv sqlc /usr/local/bin/sqlc && \    chmod +x /usr/local/bin/sqlc && \    rm sqlc_1.29.0_linux_amd64.tar.gz# ----------------------------------------------------# golang-migrate# ----------------------------------------------------RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz \    | tar -xz && \    mv migrate /usr/local/bin/# ----------------------------------------------------# Environment# ----------------------------------------------------ENV GOTOOLCHAIN=autoENV PATH="/usr/local/go/bin:/go/bin:${PATH}"# ----------------------------------------------------# Permissions# ----------------------------------------------------RUN chown -R vscode:vscode /go

Finalny post-create.sh
#!/bin/bashset -eexport PATH="/usr/local/go/bin:/go/bin:$PATH"echo "Installing pnpm..."npm install -g pnpmecho "Installing Claude Code..."npm install -g @anthropic-ai/claude-codeecho "Installing OpenAI Codex..."npm install -g @openai/codexecho "Installing Go tools..."go install golang.org/x/tools/gopls@v0.20.0go install github.com/go-delve/delve/cmd/dlv@v1.24.2go install honnef.co/go/tools/cmd/staticcheck@2025.1go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8go install github.com/air-verse/air@v1.61.7go install golang.org/x/tools/cmd/goimports@v0.38.0go install github.com/cweill/gotests/gotests@v1.6.0go install github.com/josharian/impl@v1.4.0echo "Done."

Finalny devcontainer.json
{  "name": "go-react-k8s-dev",  "build": {    "dockerfile": "Dockerfile"  },  "workspaceFolder": "/workspace",  "mounts": [    "source=${localWorkspaceFolder},target=/workspace,type=bind",    "source=go-history,target=/commandhistory,type=volume"  ],  "remoteEnv": {    "HISTFILE": "/commandhistory/.bash_history"  },  "remoteUser": "vscode",  "features": {    "ghcr.io/devcontainers/features/node:1": {      "version": "22"    },    "ghcr.io/devcontainers/features/docker-in-docker:2": {},    "ghcr.io/devcontainers/features/github-cli:1": {}  },  "postCreateCommand": "bash .devcontainer/post-create.sh",  "customizations": {    "vscode": {      "extensions": [        "golang.Go",        "ms-azuretools.vscode-docker",        "ms-kubernetes-tools.vscode-kubernetes-tools",        "redhat.vscode-yaml",        "dbaeumer.vscode-eslint",        "esbenp.prettier-vscode",        "bradlc.vscode-tailwindcss",        "eamodio.gitlens",        "ms-vscode.makefile-tools",        "GitHub.copilot"      ],      "settings": {        "editor.formatOnSave": true,        "editor.quickSuggestions": {          "other": true,          "comments": false,          "strings": false        },        "editor.suggestOnTriggerCharacters": true,        "editor.acceptSuggestionOnCommitCharacter": true,        "go.useLanguageServer": true,        "go.toolsManagement.autoUpdate": false,        "go.formatTool": "goimports",        "go.lintTool": "golangci-lint",        "gopls": {          "ui.semanticTokens": true,          "staticcheck": true,          "gofumpt": true        },        "terminal.integrated.defaultProfile.linux": "bash"      }    }  }}

Jak teraz pracować?
Pierwszy raz
Dev Containers: Rebuild Without Cache

Codziennie
Po prostu:
Reopen in Container

Gdy zmienisz:


Dockerfile,


devcontainer.json,


tooling,


robisz:
Rebuild Container

Co osiągnąłeś?
Masz setup bardzo zbliżony do:


profesjonalnych Go backend teams,


cloud-native teams,


platform engineering,


AI-assisted development environments.


Masz:


izolację,


reproducibility,


pinned tooling,


poprawny Go language server,


Kubernetes tooling,


AI agents w sandboxie,


profesjonalny DX.