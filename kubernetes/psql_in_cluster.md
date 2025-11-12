kubectl get pods --all-namespaces | grep cmk-postgresql

## Zalogowanie do kontenera
kubectl exec -it cmk-postgresql-0 -n cmk -- bash
kubectl exec -it cmk-postgresql-0 -n cmk -- sh

psql -U postgres -d postgres

## Połączenie z port-forward
kubectl port-forward -n cmk cmk-postgresql-0 5432:5432
psql -h localhost -p 5432 -U postgres -d cmk


\l       -- lista baz danych
\dn      -- lista schematów
\dt x.*  -- tabele ze schematu x
