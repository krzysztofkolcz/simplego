
  1. PostgreSQL — jednorazowa instalacja

  # Namespace dla bazy
  kubectl create namespace database

  # Dodaj repo bitnami (jeśli nie masz)
  make psql2-add-bitnami

  # Zainstaluj sealed-secrets (do zarządzania hasłami w k8s)
  kubectl create namespace sealed-secrets
  make sealed-add-repo
  make sealed-install

  # Wygeneruj zaszyfrowany secret z hasłami (charts/postgres/postgres-secret.yaml → sealedsecret)
  make sealed-generate

  # Zainstaluj PostgreSQL
  make psql2-install-psql

  Hasła są w charts/postgres/postgres-secret.yaml — baza łączy się w klastrze pod postgresql.database.svc.cluster.local.