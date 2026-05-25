# Dodaj helm repo
```
helm repo add doppler https://helm.doppler.com
helm repo update
```

# Zainstaluj operator
```
helm upgrade -install doppler-operator doppler/doppler-kubernetes-operator \
  --namespace doppler-operator-system \
  --create-namespace
```
Może być błąd

# Zmieniam deployment.yaml i values.yaml
values.yaml:
Usuwam pojedyńcze zmienne (pozostawiam tylko APP_NAME) i dodaje secretName: my-migrations-secrets
```
app:
  name: go-migrations
  secretName: my-migrations-secrets  
  env:
    - name: APP_NAME
      value: "go-migrations-app"
```
secrets będą szły z my-migrations-secrets

deployment.yaml:
```
{{ toYaml .Values.app.env | indent 12 }}
          {{- if .Values.app.secretName }}
          envFrom:
            - secretRef:
                name: {{ .Values.app.secretName }}
          {{- end }}
```

# Service token
```
doppler configs tokens create local-k8s-my-migrations \
  --project my-migrations \
  --config local \
  --plain
```
REMOVED_DOPPLER_TOKEN

→ skopiuj token: dp.st.local.xxx...

```
doppler secrets set DB_NAME "appdb" --project my-migrations --config local
doppler secrets set DB_USER "appuser" --project my-migrations --config local
doppler secrets set DB_PASS "strongpassword" --project my-migrations --config local
doppler secrets set DB_PORT "5432" --project my-migrations --config local
doppler secrets set DB_HOST "postgresql.database.svc.cluster.local" --project my-migrations --config local
doppler secrets set POSTGRES_PASSWORD "superadminpassword" --project my-migrations --config local
```
# Zapisanie tokenu do ns aplikacji
```
kubectl create secret generic doppler-token \
  --namespace go-migrations-ns \
  --from-literal=serviceToken="REMOVED_DOPPLER_TOKEN"
```

# Doppler secret
k8s/local/my-migrations/doppler-secret.yaml
zapisałem w:
./my-migrations/charts/doppler/doppler-secret.yaml
```
kubectl apply -f charts/doppler/doppler-secret-yaml
```
```
apiVersion: secrets.doppler.com/v1alpha1
kind: DopplerSecret
metadata:
  name: doppler-sync
  namespace: go-migrations-ns
spec:
  tokenSecret:
    name: doppler-token
    namespace: go-migrations-ns
  managedSecret:
    name: my-migrations-secrets    # ← musi zgadzać się z secretName w values.yaml
    namespace: go-migrations-ns
```
```
kubectl apply -f k8s/local/my-migrations/doppler-secret.yaml
```

# Weryfikacjia
## Sprawdź czy secret został utworzony
```
kubectl get secret my-migrations-secrets -n go-migrations-ns
```

## Sprawdź czy sekrety mają właściwe wartości
```
kubectl get secret my-migrations-secrets -n go-migrations-ns \
  -o jsonpath='{.data}' | jq 'to_entries[] | {key: .key, value: (.value | @base64d)}'
```