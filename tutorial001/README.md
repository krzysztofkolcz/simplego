https://www.youtube.com/watch?v=1Lu1F94exhU

# Alias
alias k='kubectl'

# Namespace
k create -f namespace-prod.yaml
k get namespace
k describe namespace prod

k get pods
> No resources found....
 
k get pods -n prod

k config set-context --current --namespace=prod


# Create HelloWorld deployment
k create deployment hello-node --image=k8s.gcr.io/echoserver:1.4
k create deployment hello-node --image=k8s.gcr.io/echoserver:1.4 -n dev
k get pods -n dev
k get pods --all-namespaces

# Events
k get events -n dev

# Services
To access pods we need services 
k expose deployment hello-node --type=LoadBalancer --port=8080 -n dev
k get services
k get services -n dev
