# Kurs: Wdrożenie prostego programu Go z Helm — krok po kroku
To chyba nie jest dokladnie ta konwersacja, ale podobne:
https://chatgpt.com/c/69038bcb-f20c-8326-b77d-5d3f159093da

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

```makefile
make build-image
```

Uwaga na rózne wersje go w Dockerfile i go.mod!

## Uruchomienie rejestru, klastra k3d, import image do rejestru
```
k3d cluster create go-hello-cluster \
  --registry-create go-hello-registry:0.0.0.0:5000
```

```makefile
# 0
create-k3d-cluster-with-registry
	k3d cluster create $(CLUSTER_NAME) -p "30083:30083@server:0" --api-port 127.0.0.1:6445 --registry-create $(REGISTRY_NAME):$(REGISTRY_PORT); \
  ...
```

lub osobno (ale z tym mialem problemy testujac):
```
k3d registry create go-hello-registry --port 5000
k3d cluster create go-hello-cluster --registry-use k3d-go-hello-registry:5000
```

Tagowanie image
```
docker build -t kkolcz/go-hello:0.1.0 .
docker tag kkolcz/go-hello:0.1.0 k3d-go-hello-registry:5000/kkolcz/go-hello:0.1.0
```
```makefile
# 1
build-image:
	docker build -t $(TAG) .

# 2
tag-image:
	docker tag $(TAG) localhost:$(REGISTRY_PORT)/$(TAG)
	docker tag $(TAG) $(REGISTRY_NAME):$(REGISTRY_PORT)/$(TAG)
```


k3d-$(REGISTRY_NAME):$(REGISTRY_PORT)/$(TAG) 
to jest „tylko” tag, ale ten tag ma bardzo konkretne znaczenie – wskazuje do jakiego rejestru Docker ma wysłać obraz.


Pushowanie image. Registry dostepne z hosta pod adresem localhost:5000
Wyjaśnienie dlaczego pushuje na localhost:5000 a nie na k3d-go-hello-registry:5000
https://chatgpt.com/c/69666347-4bc0-8327-96f3-2ce4ba80eca4

Chodzi o to, ze k3d-go-hello-registry:5000 jest dostepne wewnatrz sieci kubernetesa, a nie z hosta.
```
docker push localhost:5000/kkolcz/go-hello:0.1.0
```
```makefile
# 3
k3d-import-image: build-image
	@echo "Importing docker image into k3d"
	docker push localhost:$(REGISTRY_PORT)/$(TAG)
```
Musze otagowac image jako localhost:$(REGISTRY_PORT)/$(TAG)

### Import image bez rejestru

```bash
docker build -t kkolcz/go-hello:0.1.0 .
k3d image import --cluster go-hello-cluster kkolcz/go-hello:0.1.0
```

### Push obrazu do rejestru (np. Docker Hub)

Zaloguj się i wypchnij:

```bash
docker login
docker push kkolcz/go-hello:0.1.0
```

## List komend
### Lista klastrow
```
k3d cluster list

```

### Rejestr

pobranie obrazow z lokalnego registry (dziala dla rejstru lokalnego na porcie 5000)
```
curl http://localhost:5000/v2/_catalog
```

lista tagow:
```
curl http://localhost:5000/v2/kkolcz/go-hello/tags/list
```

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

### Warto usuwać zbędne rzeczy.

https://chatgpt.com/c/69791d1b-3608-832b-b24f-97d6499c5a8c

Usuniemy zbędne rzeczy i zostawimy minimalny chart.
Chart ma opisywać to, co faktycznie deployujesz – nie wszystko, co Helm potrafi.
Czyli:
❌ nie zostawiamy „bo może się przyda”
✅ zostawiamy tylko to, co realnie używane
📉 im mniejszy chart → mniej błędów → łatwiejsze utrzymanie

### Co robi helm create w praktyce?
helm create generuje demo-chart:
pod wszystkie możliwe use-case’y
edukacyjny
nie produkcyjny
Dlatego masz tam m.in.:
ingress.yaml
hpa.yaml
serviceaccount.yaml
tests/
masę opcji w values.yaml

Większość projektów używa 20–30% tego, co tam jest.

Na razie usuwam
./go-hello/templates/hpa.yaml
./go-hello/templates/NOTES.txt

### Co zdecydowanie warto usuwać, jeśli nie używasz
❌ templates/
Usuń bez żalu, jeśli nie korzystasz:
templates/hpa.yaml
→ jeśli nie masz HPA
templates/ingress.yaml
→ jeśli ingress masz poza chartem albo innym tool’em
templates/serviceaccount.yaml
→ jeśli używasz default SA
templates/tests/*
→ jeśli nie robisz helm testów
templates/NOTES.txt
→ jeśli nikt tego nie czyta (99% zespołów)

❌ values.yaml

To jest największy anty-pattern:

autoscaling:
  enabled: false


Jeśli nie planujesz autoscalingu → usuń całą sekcję.
Dlaczego?
ktoś kiedyś ustawi enabled: true
HPA się stworzy
prod zacznie żyć własnym życiem 😬

Co zostawić zawsze (core)

Minimum sensownego chartu:

Chart.yaml
values.yaml
templates/
  deployment.yaml
  service.yaml
  _helpers.tpl


To jest solidny fundament 90% aplikacji backendowych.

Kiedy warto coś zostawić „na przyszłość”?

Tylko jeśli:

masz konkretny plan (np. HPA w Q2)

i komentarz dlaczego jest wyłączone

Przykład OK:

# Planned: enable HPA after load tests (Q2 2026)
autoscaling:
  enabled: false


Przykład ❌:

autoscaling:
  enabled: false

---

## 6. Dostosowanie szablonów

### `values.yaml` — przykład

```yaml
replicaCount: 1

image:
  # repository: kkolcz/go-hello
  repository: go-hello-registry:5000/kkolcz/go-hello
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
To świadomo odchudzona wersja deployment.yaml.
Usuwam (z pomocą chata) niepotrzebne części:

#### Co dokładnie wyciąłeś (świadomie)?
Element	Dlaczego out
autoscaling	nie potrzebujesz na start
serviceAccount	default wystarcza
securityContext	Go app + non-root później
liveness/readiness	dodaje się, gdy endpointy gotowe
resources	brak limitów = mniej throttlingu na dev
volumes	config przez ENV
affinity/tolerations	overkill
imagePullSecrets	local / public registry

Najlepsza strategia (i dokładnie to, co zrobiłeś):
Start: minimalny Deployment
A potem dokładać elementy świadomie, gdy:
wiesz po co

Przykład:
„pody giną” → livenessProbe
„OOMKilled” → resources
„multi-AZ” → affinity
„sekrety” → volumes + secrets

#### deployment
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
## Wyjaśnienie go template engine wykorzystywanego w helm chart
### {{ include "..." }}
```
name: {{ include "go-hello.fullname" . }}
```

include = funkcja Helm, która wywołuje definicję helpera z _helpers.tpl
"go-hello.fullname" = nazwa helpera (w _helpers.tpl np. funkcja tworząca pełną nazwę release)
. = kontekst całego release (wszystkie wartości .Values, .Release, itd.)

### {{ .Values... }}
```
type: {{ .Values.service.type }}
```

.Values → zawartość pliku values.yaml
.Values.service.type → bierze wartość z:
```
service:
  type: ClusterIP
```
Wynik w YAML:
```
type: ClusterIP
```
### {{ .Release.Name }}
Selector: include "go-hello.name" . i .Release.Name
```
selector:
  app.kubernetes.io/name: {{ include "go-hello.name" . }}
  app.kubernetes.io/instance: {{ .Release.Name }}
```
include "go-hello.name" . → helper w _helpers.tpl który zwraca np. samą nazwę chartu (go-hello)
.Release.Name → nazwa release podana przy helm install demo ... → np. demo

Jeśli instalujesz z:
```
helm install myapp ./go-hello
```
to .Release.Name = myapp.

### Funkcje Golangowe – np. toYaml, indent, upper
```
{{ toYaml .Values.app.env | indent 12 }}
```

toYaml konwertuje strukturę z values.yaml do YAML.
| indent 12 dodaje 12 spacji wcięcia, żeby było poprawnie w YAML.

### {{ .Chart.Name }}
.Chart – informacje o samym chartcie z Chart.yaml:
```
name: {{ .Chart.Name }}  # go-hello
version: {{ .Chart.Version }}  # 0.1.0
```

## 7. Wartości i nadpisywanie

Przykład instalacji:

```bash
helm install demo ./go-hello -f ./go-hello/values.yaml
```

```makefile
# 3
k3d-install-helm:
	helm install $(RELEASE) $(CHART_DIR) -f $(CHART_DIR)/values.yaml
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

### Podglad
```
helm template demo ./go-hello -f ./go-hello/values.yaml
---
# Source: go-hello/templates/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: postgres-custom-config
data:
  postgresql.conf: |
    log_lock_waits = on
    log_min_duration_statement = 0
    log_statement = 'all'
---
```
```
# Source: go-hello/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: demo-go-hello
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 8080
  selector:
    app.kubernetes.io/name: go-hello
    app.kubernetes.io/instance: demo
```
Co to robi:
Service w Kubernetes działa jak wewnętrzny load balancer / DNS dla podów.
type: ClusterIP → serwis dostępny tylko w klastrze.
ports:
port: 80 → port, pod którym service jest dostępny dla innych podów.
targetPort: 8080 → port w kontenerze (deployment) do którego ruch jest przekierowywany.
selector:
Service łączy się z podami, które mają te etykiety (app.kubernetes.io/name: go-hello, app.kubernetes.io/instance: demo).
💡 Efekt: inne pod-y w klastrze mogą zrobić curl http://demo-go-hello i trafią na pod go-hello.

```
---
# Source: go-hello/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-go-hello
  labels:
    app.kubernetes.io/name: go-hello
    app.kubernetes.io/instance: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: go-hello
      app.kubernetes.io/instance: demo
  template:
    metadata:
      labels:
        app.kubernetes.io/name: go-hello
        app.kubernetes.io/instance: demo
    spec:
      containers:
        - name: go-hello
          image: "kkolcz/go-hello:0.1.0"
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
          env:
            - name: APP_NAME
              value: go-hello-from-helm
```
Deployment zarządza replikami podów – zapewnia, że w klastrze zawsze działa określona liczba podów.
replicas: 1 → chcemy 1 pod.
selector → mówi Deploymentowi, które pody zarządzać (etykiety).
template → szablon poda, który Deployment tworzy.
containers → kontener go-hello:
image → obraz Dockera (kkolcz/go-hello:0.1.0)
imagePullPolicy: IfNotPresent → pobiera obraz tylko jeśli go nie ma lokalnie
ports → port 8080 w kontenerze
env → zmienna środowiskowa APP_NAME ustawiona na go-hello-from-helm
Efekt:
Deployment tworzy pod z Twoim kontenerem
Service demo-go-hello łączy się z tym podem przez port 80 → 8080

```
---
# Source: go-hello/templates/tests/test-connection.yaml
apiVersion: v1
kind: Pod
metadata:
  name: "demo-go-hello-test-connection"
  labels:
    helm.sh/chart: go-hello-0.1.0
    app.kubernetes.io/name: go-hello
    app.kubernetes.io/instance: demo
    app.kubernetes.io/version: "1.16.0"
    app.kubernetes.io/managed-by: Helm
  annotations:
    "helm.sh/hook": test
spec:
  containers:
    - name: wget
      image: busybox
      command: ['wget']
      args: ['demo-go-hello:80']
  restartPolicy: Never
```
To jest Helm test hook ("helm.sh/hook": test) → uruchamiany po wdrożeniu przez helm test.
Tworzy pod tymczasowy:
Kontener busybox z poleceniem wget demo-go-hello:80
Celem jest sprawdzenie, czy Service działa i odpowiada.
restartPolicy: Never → pod nie jest restartowany automatycznie
💡 Po teście pod zostanie usunięty przez Helm (nie zostaje w klastrze).

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

Jeżeli używasz `minikube` lub `kind` (na k3d tez dziala), otwórz port:

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


## Widocznosc na zewnatrz
### port-forward
```
kubectl port-forward svc/demo-go-hello 8080:80
```

```
localhost:8080
    ↓
Service demo-go-hello :80
    ↓
Pod :8080
```
### NodePort
```
service:
  type: NodePort
  port: 80
  nodePort: 30080
```
wtedy:
```
kubectl get nodes -o wide
```

i otwierasz:
```
http://localhost:30080
```

### Ingress
Jeśli masz ingress-nginx w k3d
```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-hello
spec:
  rules:
    - host: go-hello.localhost
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: demo-go-hello
                port:
                  number: 80
```
dodajesz do /etc/hosts:
```
127.0.0.1 go-hello.localhost
```
```
http://go-hello.localhost
```

## MAKEFILE
### k3d kubeconfig merge
```
k3d kubeconfig merge  $(CLUSTER_NAME) 
```

dodaje kubeconfig klastra k3d do Twojego lokalnego ~/.kube/config
i (opcjonalnie) przełącza aktualny context.
Każdy klaster k3d ma własny kubeconfig, który k3d trzyma wewnętrznie (nie zawsze zapisany w ~/.kube/config).
ta komenda robi:
pobiera kubeconfig tylko dla tego klastra
scala (merge) go z:
~/.kube/config

Co dokładnie trafia do kubeconfig?
Dodawane są
```
clusters:
- name: k3d-go-hello-cluster
  cluster:
    server: https://127.0.0.1:6443
    certificate-authority-data: ...

users:
- name: admin@k3d-go-hello-cluster
  user:
    client-certificate-data: ...
    client-key-data: ...

contexts:
- name: k3d-go-hello-cluster
  context:
    cluster: k3d-go-hello-cluster
    user: admin@k3d-go-hello-cluster

```


### Co robi --kubeconfig-switch-context?
To automatycznie ustawia nowy klaster jako aktywny context.

Aktualny kontekst zwroci:
```
kubectl config current-context
```


# Postgresql działający tutorial
https://chatgpt.com/c/69749b70-6264-832a-add3-52d079e4dc6b

Poniżej masz praktyczny tutorial krok-po-kroku, oparty o Helm + chart Bitnami, który jest najczęściej używany w produkcji i devie.

Zakładam, że:
masz działający klaster Kubernetes
masz zainstalowane: kubectl i helm
pracujesz lokalnie (np. minikube / k3s / DOKS / EKS)
Tutorial: PostgreSQL w Kubernetes + baza danych + użytkownik

## 1. Sprawdź wersje
kubectl version --client
helm version

## 2. Utwórz namespace (zalecane)
kubectl create namespace database

## 3. Dodaj repozytorium Helm Bitnami
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update

## 4. Przygotuj plik values.yaml

Utwórz plik postgresql/values.yaml:
(architektonicznie, lepiej nie łączyć plików values.yaml i postgresql-values.yaml)
values.yaml - values dla chartu aplikacji

POPRAWNE PODEJŚCIE #1 (najlepsze): osobne charty + osobne values
Struktura repo

```
helm/
├── go-hello/
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│
└── postgresql/
    └── values.yaml
```

postgresql/values.yaml:
```
auth:
  username: appuser
  password: strongpassword
  database: appdb
  postgresPassword: superadminpassword

primary:
  persistence:
    enabled: true
    size: 10Gi

  resources:
    requests:
      memory: 256Mi
      cpu: 250m
    limits:
      memory: 512Mi
      cpu: 500m
```

Co tu się dzieje?
✔ tworzony jest użytkownik appuser
✔ tworzona jest baza appdb
✔ hasło dla roota (postgres)
✔ PersistentVolume (dane nie znikną po restarcie)

## 5. Zainstaluj PostgreSQL
helm install postgresql bitnami/postgresql \
  -n database \
  -f postgresql/values.yaml

Sprawdź status:
kubectl get pods -n database

Powinno być:
postgresql-0   1/1   Running

## 6. Sprawdź, czy baza i użytkownik istnieją
### 6.1 Pobierz hasło (jeśli nie użyłeś własnego)

Jeśli nie podałeś haseł w values.yaml, możesz je pobrać:

kubectl get secret postgresql -n database \
  -o jsonpath="{.data.postgres-password}" | base64 -d

### 6.2 Wejdź do kontenera PostgreSQL
kubectl exec -it postgresql-0 -n database -- psql \
  -U appuser \
  -d appdb

### 6.3 Sprawdź bazę i użytkownika
\du
\l

Powinieneś zobaczyć:
użytkownika appuser
bazę appdb

## 7. Połączenie z PostgreSQL z aplikacji
### 7.1 Service w Kubernetes

Bitnami tworzy Service:

kubectl get svc -n database


Zwykle:

postgresql   ClusterIP   10.x.x.x   5432/TCP

### 7.2 Connection string (z wnętrza klastra)
postgresql://appuser:strongpassword@postgresql.database.svc.cluster.local:5432/appdb

## 8. Połączenie lokalne (port-forward)
kubectl port-forward svc/postgresql 5434:5432 -n database


Połącz się lokalnie:

psql -h 127.0.0.1 -p 5434 -U appuser -d appdb

## 9. Dodanie dodatkowej bazy / użytkownika (opcjonalnie)

Jeśli chcesz więcej baz:

kubectl exec -it postgresql-0 -n database -- psql -U postgres

CREATE DATABASE otherdb;
CREATE USER otheruser WITH PASSWORD 'pass123';
GRANT ALL PRIVILEGES ON DATABASE otherdb TO otheruser;

## 10. Usunięcie PostgreSQL (uwaga na dane!)
helm uninstall postgresql -n database

⚠️ PV zostaje – dane nadal są na dysku
Aby usunąć wszystko:

kubectl delete pvc -n database --all

## 11. Struktura „produkcyjna” (rekomendacja)

hasła → Kubernetes Secrets

values.yaml → bez haseł

backup → pg_dump + CronJob

produkcja → managed PostgreSQL (DO, AWS, GCP)


# Sealed Secret
W klastrze działa controller Sealed Secrets
Ma on klucz prywatny, którego nie ma nikt poza klastrem
Ty lokalnie:
bierzesz Secret
szyfrujesz go narzędziem kubeseal
Do Gita trafia SealedSecret (zaszyfrowany YAML)
Controller w klastrze:
odszyfrowuje go
tworzy normalny Secret
📌 Nawet jeśli ktoś ukradnie repo — nie odczyta sekretów

## Create secret

```
kubectl create secret generic db-secret \
  --from-literal=DB_PASSWORD=supersecret \
  --dry-run=client -o yaml > secret.yaml
```
## seal secret
```
kubeseal \
  --controller-namespace kube-system \
  --format yaml < secret.yaml > sealed-secret.yaml
```

# K9s
## Ubuntu
curl -sS https://webinstall.dev/k9s | bash
