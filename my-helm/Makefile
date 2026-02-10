NAMESPACE=my-helm-ns
RELEASE=my-release
CHART_DIR=./_base/go-hello 
CHART_DIR_01=./01-basic-deploy
CHART_DIR_02=./02-config-secret
CHART_DIR_03=./03-sealed-secret
VALUES=values.yaml

create-ns:
	kubectl create namespace $(NAMESPACE)

install-helm-01:
	helm install $(RELEASE) $(CHART_DIR) -f $(CHART_DIR_01)/$(VALUES) -n $(NAMESPACE)

install-helm-02:
	helm upgrade --install $(RELEASE) $(CHART_DIR_02) -f $(CHART_DIR_02)/$(VALUES) -n $(NAMESPACE)

sealed-secret-03:
	kubeseal --format yaml --namespace $(NAMESPACE) < ./03-sealed-secret/secret-raw.yaml > ./03-sealed-secret/sealedsecret-org.yaml

install-helm-03:
	helm upgrade --install $(RELEASE) $(CHART_DIR_03) -f $(CHART_DIR_03)/$(VALUES) -n $(NAMESPACE)

template-helm-03:
	helm template $(RELEASE) $(CHART_DIR_03) -f $(CHART_DIR_03)/$(VALUES) -n $(NAMESPACE)

get-sealed-03:
	kubectl get sealedsecret -n my-helm-ns

get-secret-03:
	kubectl get secret -n my-helm-ns

forward:
	kubectl port-forward svc/go-hello-my-release 8080:8080