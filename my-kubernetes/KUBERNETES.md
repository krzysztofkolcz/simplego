# Lista wszystkich kontekstów
kubectl config get-contexts

# Aktualnie używany kontekst
kubectl config current-context

# Jak zmieniać klaster (context) — klasycznie
kubectl config use-context k3d-go-hello-cluster

bardzo łatwo strzelić deploy na zły klaster, dlatego seniorzy używają kubectx.
sudo apt install kubectx

## Zmiana klastra (1 komenda)
```
kubectx
```

## kubectx + fuzzy finder (TODO)
https://www.youtube.com/watch?v=oTNRvnQLLLs

# Namespace’y — kubens (para do kubectx)
## Lista namespace
```
kubens
```

## Zmiana namespace
kubens myrfns

Od teraz:
kubectl get pods
działa w myrfns, bez -n.

# Najczęstszy problem: różne terminale = różne klastry 😬
Sprawdź:
echo $KUBECONFIG
Jeśli:
w jednym terminalu jest ustawione
w drugim nie
➡️ używacie różnych kubeconfigów
🔧 Sprawdź skąd kubectl bierze config
kubectl config view --minify

🧠 Best practice (senior-level)
export KUBECONFIG=~/.kube/config

Albo wiele plików:

export KUBECONFIG=~/.kube/config:~/.kube/kubeconfig-garden

# Workflow seniora (polecam 💡)
kubectx            # wybieram klaster
kubens myrfns      # wybieram namespace
kubectl get pods

Zawsze:
najpierw sprawdzam klaster
potem namespace
dopiero potem deploy / delete / upgrade

# Bonus: aliasy, które przyspieszają x3
alias k=kubectl
alias kx=kubectx
alias kn=kubens

I nagle:
kx
kn myrfns
k get pods

# Co mi da: export KUBECONFIG=~/.kube/config:~/.kube/kubeconfig-garden
## 1️⃣ Co robi export KUBECONFIG=...:...?
export KUBECONFIG=~/.kube/config:~/.kube/kubeconfig-garden
➡️ Mówisz kubectl:
„Nie czytaj jednego pliku, tylko połącz (merge) te kubeconfigi w jeden logiczny widok”.
🔧 Co dokładnie się dzieje?
kubectl czyta pliki od lewej do prawej
scala sekcje:
clusters users contexts
konteksty z obu plików są widoczne naraz

Efekt:
kubectl config get-contexts
pokazuje wszystkie klastry:
lokalne (k3d, kind, minikube)
firmowe / cloudowe (Garden, GKE, EKS, AKS)

Bez kopiowania czegokolwiek.

## 2️⃣ Co jest w ~/.kube/config?
To jest domyślny kubeconfig.
Zawiera zwykle:
lokalne klastry:
minikube
kind
k3d
czasem:
ręcznie dodane klastry
stare testowe środowiska
Przykład:
```
clusters:
- name: k3d-go-hello
  cluster:
    server: https://0.0.0.0:6443

users:
- name: admin
  user:
    client-certificate-data: ...

contexts:
- name: k3d-go-hello-cluster
  context:
    cluster: k3d-go-hello
    user: admin
    namespace: default


👉 Twoje lokalne playground / dev

### 3️⃣ Co jest w ~/.kube/kubeconfig-garden?

To jest zewnętrzny kubeconfig, zwykle:
klaster firmowy
klaster produkcyjny lub staging
zarządzany przez:
Garden
GKE / EKS / AKS
VPN / SSO / certyfikaty
Często zawiera:
tokeny OIDC
certyfikaty
dynamiczne auth (exec plugin)
Przykład:
```
clusters:
- name: kms2s-prod
  cluster:
    server: https://api.prod.k8s.company.com

users:
- name: garden-user
  user:
    exec:
      command: gardenlogin
      args: ["get-token"]

contexts:
- name: garden-kms2s--aweu-cmk-p-02
  context:
    cluster: kms2s-prod
    user: garden-user
```

👉 Prawdziwe środowiska (staging / prod)

### 4️⃣ Co by było BEZ tego exporta?
❌ Bez KUBECONFIG
kubectl config get-contexts
➡️ tylko:
k3d-go-hello-cluster
Nie widzisz:
klastra garden
prod
staging

Bo kubectl domyślnie czyta tylko ~/.kube/config.

### 5️⃣ Co by było, gdybyś ustawił TYLKO garden?
export KUBECONFIG=~/.kube/kubeconfig-garden
➡️ Nagle:
znikają lokalne klastry
kubectx pokazuje tylko prod/staging
To jest częsta pułapka.

### 6️⃣ Dlaczego to jest dobre rozwiązanie?
✅ Plusy
✔ nie kopiujesz kubeconfigów
✔ masz wszystkie klastry w jednym widoku
✔ kubectx działa idealnie
✔ łatwo rozdzielić local vs cloud
✔ zero konfliktów

⚠️ Jedyny minus

Jeśli oba pliki mają context o tej samej nazwie
➡️ wygrywa ten późniejszy (z prawej)

### 7️⃣ Jak sprawdzić, co faktycznie jest załadowane?
kubectl config view
Tylko aktywny kontekst:
kubectl config view --minify
Skąd pochodzi kontekst:
kubectl config get-contexts

### 8️⃣ Best practice (które polecam 🔒)
~/.zshrc albo ~/.bashrc
export KUBECONFIG="$HOME/.kube/config:$HOME/.kube/kubeconfig-garden"

I nigdy więcej o tym nie myślisz.

KUBECONFIG=plik1:plik2 → merge kubeconfigów
~/.kube/config → lokalne klastry
~/.kube/kubeconfig-garden → firmowe / prod
razem → pełna lista klastrów
kubectx działa jak marzenie 💙