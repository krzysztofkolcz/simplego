# Dodaj helm repo
```
helm repo add doppler https://helm.doppler.com
```

```
"doppler" has been added to your repositories
```

```
helm repo update
```
```
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "doppler" chart repository
...Successfully got an update from the "sealed-secrets" chart repository
...Successfully got an update from the "ingress-nginx" chart repository
...Successfully got an update from the "bitnami" chart repository
Update Complete. ⎈Happy Helming!⎈
```

# Zainstaluj operator
```
helm install doppler-operator doppler/doppler-kubernetes-operator \
  --namespace doppler-operator-system \
  --create-namespace
```

Uwaga: otrzymałem błąd
```
Error: INSTALLATION FAILED: 1 error occurred:
	* namespaces "doppler-operator-system" already exists
```


# Service token dla klastra
## Utwórz service token dla lokalnego dev
Każdy klaster potrzebuje osobnego tokenu z dostępem tylko do jednego środowiska:
```
doppler configs tokens create local-cluster \
  --project myapp \
  --config dev \
  --plain
```
output
```
# → dp.st.dev.xxxxxxxxxxxx
```

Czyli rozumiem, że jeden klaster mogę połączyć z jednym środowiskiem należącym do jednego projektu?

## Utwórz Kubernetes Secret z tokenem (jednorazowo):
```
bashkubectl create secret generic doppler-token-local \
  --namespace doppler-operator-system \
  --from-literal=serviceToken="dp.st.dev.xxxxxxxxxxxx"
```

## DopplerSecret manifest
```
# k8s/doppler-secret.yaml
apiVersion: secrets.doppler.com/v1alpha1
kind: DopplerSecret
metadata:
  name: myapp-doppler
  namespace: default
spec:
  tokenSecret:
    name: doppler-token-local      # Kubernetes Secret z tokenem
  managedSecret:
    name: myapp-secrets            # nazwa tworzonego Kubernetes Secret
    namespace: default
```
```
kubectl apply -f k8s/doppler-secret.yaml
```

## Sprawdź czy sekrety zostały zsynchronizowane
```
kubectl get secret myapp-secrets -o jsonpath='{.data}' | jq
```

## Deployment aplikacji
```
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: myapp
          image: myapp:latest
          envFrom:
            - secretRef:
                name: myapp-secrets  # wszystkie sekrety z Dopplera naraz
```