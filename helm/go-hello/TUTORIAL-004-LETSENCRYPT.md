Czym jest ten yaml?:
```
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    email: krzysztof.kolcz@gmail.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```
 mogę zobaczyć ten zasób w k9s?


To jest manifest Kubernetes, który tworzy zasób cert-manager odpowiedzialny za automatyczne wystawianie certyfikatów TLS z Let’s Encrypt.

Rozbijmy go na części 👇

To jest manifest Kubernetes, który tworzy zasób cert-manager odpowiedzialny za automatyczne wystawianie certyfikatów TLS z Let’s Encrypt.

Rozbijmy go na części 👇

# 🔐 Co to jest ten YAML?
Tworzy on zasób:
👉 ClusterIssuer
czyli globalny (dla całego klastra) issuer certyfikatów dla cert-managera.
Cert-manager to kontroler w Kubernetes, który:
automatycznie generuje certyfikaty TLS
odnawia je przed wygaśnięciem
integruje się np. z Let’s Encrypt

## 📄 Szczegółowe wyjaśnienie pliku
### 🧩 Typ zasobu
```
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
```
ClusterIssuer → issuer dostępny w całym klastrze
alternatywa: Issuer → działa tylko w jednym namespace

### 🏷️ Metadata
```
metadata:
  name: letsencrypt-prod
```

To jest nazwa zasobu.
Później możesz jej używać np. w:

Certificate
Ingress annotations

### ⚙️ Spec – konfiguracja ACME
```
spec:
  acme:
```

ACME to protokół używany przez Let’s Encrypt do wystawiania certyfikatów.

### 📧 Email
```
email: krzysztof.kolcz@gmail.com
```

Let’s Encrypt używa go do:
powiadomień o wygasaniu
problemów z certyfikatem

### 🌐 Serwer Let’s Encrypt
```
server: https://acme-v02.api.letsencrypt.org/directory
```

To jest produkcyjny endpoint Let’s Encrypt.
👉 Istnieje też staging (do testów):
```
https://acme-staging-v02.api.letsencrypt.org/directory
```

### 🔑 Secret z kluczem ACME
```
privateKeySecretRef:
  name: letsencrypt-prod
```

cert-manager zapisze tam klucz konta ACME.
To NIE jest certyfikat — tylko klucz konta w Let’s Encrypt.

### 🧪 Solver – jak udowodnić własność domeny
```
solvers:
- http01:
    ingress:
      class: nginx
```

Oznacza:
👉 użyj HTTP-01 challenge
Let’s Encrypt:
wywołuje:
http://twojadomena/.well-known/acme-challenge/...

cert-manager:
tworzy tymczasowy ingress
zwraca token

🧠 Dlaczego ClusterIssuer?
Masz jeden issuer dla:
wielu namespace
wielu aplikacji
wielu ingressów
To jest standardowa praktyka.

## ✅ Czy można zobaczyć ten zasób w k9s?
Tak 👍

### 🔎 Jak znaleźć ClusterIssuer w k9s (NIE DZIAŁA)
W k9s wpisz:
:clusterissuers
lub krócej:
:ci

Jeśli alias nie działa
Możesz użyć:

### (Działa)
:crds


i znaleźć:

clusterissuers.cert-manager.io
| certificaterequests cert-manager.io CertificateRequest v1 Namespaced 260d │
│ certificates        cert-manager.io Certificate        v1 Namespaced 260d |

### Alternatywnie przez kubectl

```
kubectl get clusterissuer
```
albo:
```
kubectl describe clusterissuer letsencrypt-prod
```

###  🧪 Jak sprawdzić czy działa poprawnie
Po deployu:

```
kubectl get clusterissuer letsencrypt-prod
```

Powinieneś zobaczyć:

STATUS   Ready

📌 Jak używa się tego dalej

Np. w Ingress:

annotations:
  cert-manager.io/cluster-issuer: letsencrypt-prod

⚠️ Ważna rzecz (częsty błąd)

ClusterIssuer działa tylko jeśli:

👉 cert-manager jest zainstalowany
👉 ingress nginx istnieje
👉 domena wskazuje na ingress

Jeśli chcesz, mogę Ci pokazać:

✅ pełny flow
Ingress → Certificate → Secret
✅ jak cert-manager generuje certyfikat
✅ jak debugować gdy Let’s Encrypt nie działa
✅ różnicę ClusterIssuer vs Issuer
✅ staging vs production best practices

Chcesz iść w którąś z tych stron?