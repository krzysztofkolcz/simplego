```
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/latest/download/controller.yaml
```
```
kubectl get pods -n kube-system | grep sealed
```
```
sealed-secrets-controller-6d78c4ddb-kj2nh   1/1     Running   0             7s
```

# Jak powstaje SealedSecret 
## 3.1 Tworzysz Zwykły Secret (lokalnie)
./03-sealed-secret
secret-raw.yaml (NIE COMMITUJ!)

```
kubeseal \
  --format yaml \
  --namespace my-helm-ns \
  < secret-raw.yaml \
  > sealedsecret.yaml
```

Efekt:
zaszyfrowany secret, bezpieczny do repo, działa tylko w tym klastrze + namespace

# Wyjaśnienie
Sealedsecret zdeployowany na klaster zostaje odszyfrowany przez controller
Następnie jest już wykorzystywany jako secret i moge pobierać dane
Czyli jak mam:
```
kind: SealedSecret
metadata:
  creationTimestamp: null
  name: {{ include "go-hello.fullname" . }}-secret
  namespace: my-helm-ns
spec:
  encryptedData:
    DB_NAME: Ag...
    DB_PASS: AgC...
    DB_PASSWORD: AgC...
```

To zostanie z tego zrobiony secret z taką samą nazwą jak { .metadata.name }, czyli go-hello-my-release-secret

A w tym momencie w deploymencie mam:
```
  template:
    metadata:
      labels:
        app: {{ include "go-hello.name" . }}
    spec:
      containers:
        - name: app
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - containerPort: {{ .Values.service.port }}
          envFrom:
          - configMapRef:
              name: {{ include "go-hello.fullname" . }}-config
          - secretRef:
              name: {{ include "go-hello.fullname" . }}-secret
```
czyli zmienne środowiskowe będą zaczytywane z tego pliku secret
```
name: {{ include "go-hello.fullname" . }}-secret
```
i dostępne np. 
```
dbname := os.Getenv("DB_NAME")
```