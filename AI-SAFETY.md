Ubuntu Host
│
├── VM / Dev Container
│     ├── Claude/Codex
│     ├── Go project
│     ├── ephemeral AWS credentials
│     └── brak dostępu do host secrets
│
├── Vault
│     ├── generuje short-lived secrets
│     └── audytuje dostęp
│
├── AWS IAM Role
│     └── minimalne permissions
│
└── CloudTrail + Vault Audit Logs

# User
```
sudo useradd -m aiagent
sudo passwd aiagent
```

```
su - aiagent
```
# Instalacja docker
```
sudo apt remove docker docker-engine docker.io containerd runc
```

```
sudo apt update

sudo apt install \
  ca-certificates \
  curl \
  gnupg \
  lsb-release
```

```
sudo mkdir -p /etc/apt/keyrings
```

# Dev container
mkdir -p .devcontainer