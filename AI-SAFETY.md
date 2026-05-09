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