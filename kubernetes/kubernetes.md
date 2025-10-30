
docker build -t registry-service:0.0.1 .

k3d image import registry-service:0.0.1 -c cmkcluster


kubectl get svc -n cmk
kubectl port-forward svc/cmk-postgresql 5433:5432 -n cmk 