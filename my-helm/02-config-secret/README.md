# 02 – ConfigMap & Secret

## Cel
Przekazanie konfiguracji i sekretów do aplikacji

## Nowe elementy
- ConfigMap
- Secret
- envFrom w Deployment

## Kroki
1. Dodaj config do values.yaml
2. Wygeneruj ConfigMap
3. Wygeneruj Secret
4. Podłącz przez envFrom

## Sprawdzenie
kubectl exec pod -- env | grep APP_

## Problemy
- Secret musi być w tym samym namespace
- Nazwy muszą się zgadzać z deploymentem

## Co zapamiętać
- ConfigMap = jawne dane
- Secret ≠ bezpieczny w repo

## sprawdzenie
kubectl exec -it <pod> -- env
