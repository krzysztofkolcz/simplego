#!/bin/bash
#!/bin/bash

# 1. Zainstaluj NGINX Ingress Controller
kubectl create namespace ingress-nginx

helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --set controller.publishService.enabled=true

# 2. Zainstaluj cert-manager
kubectl apply --validate=false -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml

# Poczekaj na gotowość
echo "Czekam na cert-manager..."
kubectl wait --for=condition=Available --timeout=180s deployment/cert-manager -n cert-manager

# 3. Stwórz ClusterIssuer z Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    email: kontakt@technicarium.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
EOF

# 4. Stwórz Ingress dla aplikacji
cat <<EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: saas-ingress
  namespace default
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
    - hosts:
        - saas.technicarium.com
      secretName: saas-tls
  rules:
    - host: saas.technicarium.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-service
                port:
                  number: 80
EOF

echo "✅ Gotowe! Ingress + HTTPS skonfigurowany dla https://saas.technicarium.com"



Właśnie Krisu!, Wszystkiego najlepszego z okazji urodzin!
Zdrowia, szczęścia, pociechy z żony i syna, cierpliowści do życia i fajnych górskich tras z Rosomakami ;)