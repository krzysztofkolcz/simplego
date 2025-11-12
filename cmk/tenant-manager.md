
make docker-compose-dependencies-up-and-log


 helm-install-registry: create-registry-db
        helm upgrade --install registry oci://ghcr.io/openkcm/charts/registry \
                --namespace $(NAMESPACE) \
-               --set image.tag=v1.2.0 \
+        --set image.tag=v1.2.0 \
+        --set image.pullPolicy=Always \


values-dev.yaml
33:
  secretRef:
    type: insecure



593:
    amqp:
      url: amqp://guest:guest@rabbitmq.cmk.svc.cluster.local:5672
      target: cmk.global.tenants
      source: cmk.emea.tenants
    secretRef:
      type: insecure


internal/db/dsn/postgres.go
-       return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
+       return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",


internal/testutils/db.go
-       Port: "5433",
+       Port: "5432",


# Gotestsum
go install gotest.tools/gotestsum@latest
export PATH=$(go env GOPATH)/bin:$PATH
echo 'export PATH=$(go env GOPATH)/bin:$PATH' >> ~/.bashrc
