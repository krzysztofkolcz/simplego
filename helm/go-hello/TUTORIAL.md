# Kurs: Wdrożenie prostego programu Go z Helm — krok po kroku

Ten kurs prowadzi Cię od *najprostszego* programu w Go do działającego wdrożenia na klastrze Kubernetes przy użyciu Helm. Zakładam, że masz lokalny klaster (minikube / kind / k3d / klaster w chmurze) oraz `kubectl`, `helm`, `docker`/`podman` i `git`.

> Czas wykonania: ~60–120 minut w zależności od środowiska.

---

## Spis treści

1. Wymagania wstępne
2. Utworzenie prostego programu w Go (HTTP server)
3. Dockerfile i budowanie obrazu
4. Wypchnięcie obrazu do rejestru (Docker Hub / GitHub Container Registry)
5. Utworzenie podstawowego Helm Charta
6. Dostosowanie tpl: Deployment, Service, ConfigMap
7. Wartości `values.yaml` i nadpisywanie
8. Instalacja Chartu i sprawdzenie
9. Aktualizacje (upgrade) i rollback
10. Dodanie Ingress / TLS (opcjonalnie)
11. Najczęstsze problemy i debugowanie
12. Rozszerzenia (CI/CD, liveness/readiness, secrets)

---

## 1. Wymagania wstępne

* `go` (>= 1.18)
* `docker` lub `podman` (lub `kaniko`/`buildx` do budowy obrazów)
* `kubectl` skonfigurowane do Twojego klastra
* `helm` (3.x)
* konto w Docker Hub lub GHCR (jeśli pushujesz obraz zdalnie)

Sprawdź:

```bash
go version
docker --version
kubectl version --client
helm version
```

---

## 2. Prosty program w Go — `main.go`

Stwórz katalog projektu `go-hello` i plik `main.go`:

```go
package main

import (
    "fmt"
    "net/http"
    "os"
)

func handler(w http.ResponseWriter, r *http.Request) {
    name := os.Getenv("APP_NAME")
    if name == "" {
        name = "go-hello"
    }
    fmt.Fprintf(w, "Hello from %s!\n", name)
}

func main() {
    http.HandleFunc("/", handler)
    port := ":8080"
    http.ListenAndServe(port, nil)
}
```

Szybki build i test lokalny:

```bash
go mod init example.com/go-hello
go mod tidy
go run main.go
# potem: curl http://localhost:8080
```

---

## 3. Dockerfile

Prosty, niewielki obraz:

```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.20-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./

FROM alpine:3.18
COPY --from=build /app /app
EXPOSE 8080
ENV APP_NAME=go-hello
ENTRYPOINT ["/app"]
```

Zbuduj lokalnie:

```bash
docker build -t kkolcz/go-hello:0.1.0 .
```

Uwaga na rózne wersje go w Dockerfile i go.mod!

---

## 4. Push obrazu do rejestru (np. Docker Hub)

Zaloguj się i wypchnij:

```bash
docker login
docker push kkolcz/go-hello:0.1.0
```

Jeżeli używasz GHCR, użyj odpowiednich tagów i uwierzytelnień.

> Alternatywa: `kind`/`minikube` umożliwiają użycie obrazów lokalnych bez push.

---

## 5. Tworzymy Helm Chart

Użyj `helm create`:

```bash
helm create go-hello
```

To stworzy strukturę:

```
go-hello/
  Chart.yaml
  values.yaml
  templates/
    deployment.yaml
    service.yaml
    ingress.yaml
    _helpers.tpl
```

Usuniemy zbędne rzeczy i zostawimy minimalny chart.

---

## 6. Dostosowanie szablonów

### `values.yaml` — przykład

```yaml
replicaCount: 1

image:
  repository: kkolcz/go-hello
  pullPolicy: IfNotPresent
  tag: "0.1.0"

service:
  type: ClusterIP
  port: 80

app:
  name: go-hello
  env:
    - name: APP_NAME
      value: "go-hello-from-helm"
```

### `templates/deployment.yaml` — minimalny deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "go-hello.fullname" . }}
  labels:
    app.kubernetes.io/name: {{ include "go-hello.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ include "go-hello.name" . }}
      app.kubernetes.io/instance: {{ .Release.Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ include "go-hello.name" . }}
        app.kubernetes.io/instance: {{ .Release.Name }}
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - containerPort: 8080
          env:
{{ toYaml .Values.app.env | indent 12 }}
```

### `templates/service.yaml` — prosty service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ include "go-hello.fullname" . }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: 8080
  selector:
    app.kubernetes.io/name: {{ include "go-hello.name" . }}
    app.kubernetes.io/instance: {{ .Release.Name }}
```

---

## 7. Wartości i nadpisywanie

Przykład instalacji:

```bash
helm install demo ./go-hello -f ./go-hello/values.yaml
```

Nadpisanie pojedynczej wartości:

```bash
helm install demo ./go-hello --set image.tag=0.1.1
```

Podgląd renderowanych manifestów:

```bash
helm template demo ./go-hello -f ./go-hello/values.yaml
# lub
helm install --dry-run --debug demo ./go-hello -f ./go-hello/values.yaml
```

---

## 8. Instalacja i sprawdzenie

Zainstaluj:

```bash
helm upgrade --install demo ./go-hello -f ./go-hello/values.yaml
```

Sprawdź zasoby:

```bash
kubectl get deployments
kubectl get pods
kubectl get svc
kubectl logs -l app.kubernetes.io/name=go-hello
```

Jeżeli używasz `minikube` lub `kind`, otwórz port:

```bash
kubectl port-forward svc/demo-go-hello 8080:80
# potem: curl http://localhost:8080
```

---

## 9. Aktualizacje i rollback

Aktualizacja obrazu i tagu w `values.yaml`, potem:

```bash
helm upgrade demo ./go-hello -f values.yaml
```

Rollback do poprzedniej wersji:

```bash
helm rollback demo 1
```

Sprawdź historię:

```bash
helm history demo
```

---

## 10. Ingress i TLS (opcjonalnie)

Jeżeli masz Ingress Controller (nginx/traefik), w `values.yaml` ustaw `ingress.enabled: true` i dodaj host. Do TLS użyj cert-manager lub wbudowanej obsługi (Let's Encrypt).

---

## 11. Debugowanie — najczęstsze problemy

* Pod w CrashLoop: `kubectl logs pod -c <container>` i `kubectl describe pod`.
* Obraz nie ściąga się: sprawdź `imagePullSecrets` i `imagePullPolicy`.
* Helm renderuje złe wartości: użyj `helm template` i sprawdź `values.yaml` i `--set`.

---

## 12. Rozszerzenia i dobre praktyki

* Liveness & readiness probes w `deployment.yaml`.
* Resource requests/limits.
* ConfigMap & Secret zamiast env dla wrażliwych danych.
* CI: GitHub Actions / GitLab CI — budowa obrazu, push, `helm upgrade --install`.
* Użyj `helm test` do testów smoke po wdrożeniu.

---

## 13. Przykładowy minimalny workflow CI (schemat)

1. Push do `main`
2. CI: buduj obraz, taguj, push
3. CI: `helm upgrade --install` na klasterze staging
4. Smoke tests
5. (opcjonalnie) promuj do produkcji

---

## 14. Co możesz zrobić dalej (zadania)

* Dodaj readiness/liveness.
* Dodaj autoscaling (HorizontalPodAutoscaler).
* Dodaj sekret z DB connection string i połącz z aplikacją.
* Zintegruj z GitHub Actions — automatyczne deploye.

---

### Gotowe repo przykładowe

Jeśli chcesz, mogę:

* wygenerować strukturę repo z pełnym przykładem (Go + Dockerfile + Helm) i przygotować instrukcję `README`,
* lub przygotować plik `values` dla produkcji (resources, probes, PVC, ingress),
* lub napisać pipeline CI (GitHub Actions).

Powiedz, które z tych chcesz teraz — zrobię to natychmiast.
