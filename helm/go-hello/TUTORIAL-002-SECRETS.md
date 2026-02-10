# Sealed Secrets

ARCHITEKTURA (docelowa)
Secret (SealedSecret)
        ↓
Kubernetes Secret
        ↓
ENV w Deployment
        ↓
Go: os.Getenv(...)

PostgreSQL: bierze hasła z Secret
Aplikacja: bierze te same dane z ENV

# KROK 0 – Wymagania (lokalnie)
kubeseal, k3d, helm, kubectl
```
https://github.com/bitnami-labs/sealed-secrets/releases
```
# KROK 1 – Zainstaluj Sealed Secrets Controller
kubectl create namespace sealed-secrets

helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets
helm repo update

helm install sealed-secrets sealed-secrets/sealed-secrets \
  -n sealed-secrets


Sprawdź:

kubectl get pods -n sealed-secrets

Tworzy controller 'sealed-secrets' o takiej nazwie. Jeżeli potrzebuje zmienić nazwę kontrolera musiałbym dodać fullnameOverride:
```
helm install sealed-secrets sealed-secrets/sealed-secrets \
  -n sealed-secrets \
  --set fullnameOverride=sealed-secrets-controller
```

# KROK 2 – Utwórz plain Secret (TYLKO lokalnie)
Ten plik NIE trafia do repo.

postgres-secret.yaml:
```
apiVersion: v1
kind: Secret
metadata:
  name: postgresql-secret
  namespace: database
type: Opaque
stringData:
  DB_USER: appuser
  DB_PASS: strongpassword
  DB_NAME: appdb
  POSTGRES_PASSWORD: superadminpassword
```

# KROK 3 – Zaszyfruj go do SealedSecret
kubectl create namespace database

kubeseal \
  --controller-name sealed-secrets \
  --controller-namespace sealed-secrets \
  --format yaml \
  < postgres-secret.yaml \
  > go-hello/templates/sealedsecret-postgres.yaml


💡 Od teraz tylko TEN plik trafia do repo

# KROK 4 – SealedSecret (w Helm)

go-hello/templates/sealedsecret-postgres.yaml:
```
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  creationTimestamp: null
  name: postgresql-secret
  namespace: database
spec:
  encryptedData:
    DB_USER: AgB...
    DB_PASS: AgC...
    DB_NAME: AgD...
    POSTGRES_PASSWORD: AgE...
  template:
    metadata:
      creationTimestamp: null
      name: postgresql-secret
      namespace: database
    type: Opaque
```


✔ bezpieczne
✔ git-friendly
✔ 12-factor

# KROK 5 – PostgreSQL Helm values (POPRAWIONE)

❌ USUWAMY hasła z values.yaml

postgres/values.yaml:
```
auth:
  existingSecret: postgresql-secret
  username: appuser
  database: appdb
  secretKeys:
    adminPasswordKey: POSTGRES_PASSWORD
    userPasswordKey: DB_PASS

primary:
  persistence:
    enabled: true
    size: 10Gi
```

Postgres:
bierze hasła z Secret
zero sekretów w values

# KROK 6 – ENV w go-hello/values.yaml

go-hello/values.yaml:
```
app:
  name: go-hello
  env:
    - name: APP_NAME
      value: "go-hello-from-helm"

    - name: DB_NAME
      valueFrom:
        secretKeyRef:
          name: postgresql-secret
          key: DB_NAME

    - name: DB_PASS
      valueFrom:
        secretKeyRef:
          name: postgresql-secret
          key: DB_PASS
```
✔ ENV
✔ Secret
✔ 12-factor

# KROK 7 – Deployment (JUŻ MASZ OK)
```
env:
{{ toYaml .Values.app.env | indent 12 }}
```

Helm sam wstrzyknie secretKeyRef.

# KROK 8 – Go (100% 12-factor)

Wypisanie w main.go:
```
dbname := os.Getenv("DB_NAME")
dbpass := os.Getenv("DB_PASS")
```

# KROK 9 – Deploy kolejność
## 1. sealed secret (helm)
helm install go-hello ./go-hello -n services

## 2. postgres
helm install postgresql bitnami/postgresql \
  -n database \
  -f postgres/values.yaml

kubectl get secret -n database postgresql-secret

# DLACZEGO TO JEST „PRAWIDŁOWE” 12-FACTOR
Secrets w repo	❌
SealedSecrets	✅
ENV w Podzie	✅
os.Getenv	✅
Jeden image	✅
Różne env	✅
Co dalej (polecam)