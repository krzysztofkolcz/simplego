
kubectl get svc -n cmk
kubectl port-forward svc/cmk-postgresql 5433:5432 -n cmk 

# Dodanie imaga do klastra
## Zbudowanie
docker build -t registry-service:0.0.1 .

## Dodanie - opcja A
k3d image import registry-service:0.0.1 -c cmkcluster

## Dodanie image - opcja B - Lokalne registry
k3d registry create local-registry --port 5000

output
```
INFO[0000] Creating node 'k3d-local-registry'           
INFO[0002] Pulling image 'docker.io/library/registry:2' 
INFO[0006] Successfully created registry 'k3d-local-registry' 
INFO[0006] Starting node 'k3d-local-registry'           
INFO[0006] Successfully created registry 'k3d-local-registry' 
# You can now use the registry like this (example):
# 1. create a new cluster that uses this registry
k3d cluster create --registry-use k3d-local-registry:5000

# 2. tag an existing local image to be pushed to the registry
docker tag nginx:latest k3d-local-registry:5000/mynginx:v0.1

# 3. push that image to the registry
docker push k3d-local-registry:5000/mynginx:v0.1

# 4. run a pod that uses this image
kubectl run mynginx --image k3d-local-registry:5000/mynginx:v0.1
```
docker tag registry-service:0.0.1 localhost:5000/registry-service:0.0.1
docker push localhost:5000/registry-service:0.0.1

docker tag registry-service:0.0.1 k3d-local-registry:5000/registry-service:0.0.1
docker push k3d-local-registry:5000/registry-service:0.0.1

Helm:
--set image.repository=k3d-local-registry:5000/registry-service \
--set image.tag=0.0.1


# Postgresql

kubectl port-forward svc/cmk-postgresql 5432:5432 -n cmk

## Tymaczasowy klient psql na podzie (nie zadziałał)
kubectl run psql-client --rm -it --image=postgres:16 -n cmk -- bash


# oci
helm pull oci://ghcr.io/openkcm/charts/registry --version 1.1.0
pobierze registry-1.1.0.tgz

tar -xzf registry-1.1.0.tgz

Edycja values.yaml

helm upgrade --install registry ./registry -n cmk

# pobranie configmap.yaml
kubectl get configmap registry-config -n cmk -o yaml

# usuniecie klastra
k3d cluster delete cmkcluster

# Rabbit
kubectl port-forward -n cmk pod/rabbitmq-0 15672:15672
