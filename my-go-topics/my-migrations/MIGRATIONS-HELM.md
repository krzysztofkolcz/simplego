## Pojedyńczy dockerfile, z argumentami do uruchamiania migracji
Idę po tutorialu z ./helm/go-hello


## Tworzę image z migracjami embed w golang

## Job idempotentny - TODO - wyjaśnić
ważna rzecz (często pomijana)
👉 Job musi być idempotentny
czyli:
m.Up() // + ErrNoChange

## Dockerfile.migrate
```
# Etap 1: Budowanie workera
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate

# Etap 2: Minimalny obraz uruchomieniowy
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/migrate /app/migrate

WORKDIR /app

CMD ["./migrate"]
```

## Zbudować image

## Push images (repository - dockerhub lub lokalnie)

## Kubernetes job
```
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migrate
spec:
  template:
    spec:
      containers:
        - name: migrate
          image: your-app:latest
          command: ["./app", "migrate"]
          env:
            - name: DB_URL
              valueFrom:
                secretKeyRef:
                  name: db-secret
                  key: url
      restartPolicy: Never
```
## Lokalnie - migrate-job.yaml
kubectl apply -f migrate-job.yaml

## Lokalnie - helm - test w ./helm/go-hello
hooks:
  - pre-install
  - pre-upgrade

### Kolejność
```
helm upgrade:
  1. 🔥 migracja DB (job)
  2. deploy aplikacji
```

### Drzewo plików helm
```
helm/
└── myapp/
    ├── Chart.yaml
    ├── values.yaml
    └── templates/
        ├── deployment.yaml
        ├── service.yaml
        └── migrate-job.yaml   👈 kluczowe
```
### migrate-job.yaml
```
{{- if .Values.migrate.enabled }}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "myapp.fullname" . }}-migrate
  annotations:
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": hook-succeeded
spec:
  backoffLimit: 3
  template:
    spec:
      restartPolicy: Never

      containers:
        - name: migrate
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          command: ["./app", "migrate"]   # 👈 ważne

          env:
            - name: DB_URL
              value: "{{ .Values.db.url }}"
{{- end }}
```

### values.yaml
```
image:
  repository: myapp
  tag: latest

db:
  url: postgres://user:pass@postgres:5432/db?sslmode=disable

migrate:
  enabled: true
```

## Github actions - TODO
```
- name: Run DB migration
  run: |
    kubectl apply -f k8s/migrate-job.yaml
    kubectl wait --for=condition=complete job/db-migrate --timeout=60s
```

## Migracja pada - nie robić deploy aplikacji


## migrate enable w values.yaml
```
migrate:
  enable: true
```
## Uruchomienie migracji:
jest:
```
annotations:
  "helm.sh/hook": pre-install,pre-upgrade
```

odpala się automatycznie przy:
```
helm upgrade --install
```