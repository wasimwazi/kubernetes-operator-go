# Sample Nginx Operator

The goal of this project is developing an operator that deploys a sample application and operates it on top of Kubernetes using a custom resource. 

# Pre-requisites

- Docker Desktop - https://www.docker.com/products/docker-desktop
- Kind (Kubernetes in Docker) - https://kind.sigs.K8s.io/docs/user/quick-start/
- Kubectl - https://kubernetes.io/docs/tasks/tools/

# Installation

<b>Step 1: Create a cluster</b>

Create a cluster using kind. Kind is a tool for running local Kubernetes clusters using Docker container.

```
kind create cluster --name=<cluster_name> --config <cluster_config_file>
```
A sample cluster config can be found [here](https://raw.githubusercontent.com/wasimwazi/kubernetes-operator-go/workflow/deployment_config/cluster.yaml). Note that this cluster yaml file contains configurations to support ingress. This is achieved using kubeadmConfigPatches in kind.

<b>Step 2: Install nginx ingress controller</b>

Install the nginx ingress controller for the operator using manifests and wait for it to be in ready state. Please note that the nginx ingress controller will create a namespace called ingress-nginx by default using the manifest.
```
$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
$ kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=90s
```

<b>Step 3: Install cert-manager </b>

```
$ kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.2/cert-manager.yaml
```
<b> Step 4: Pull the docker image for the operator</b>
```
$ docker pull ghcr.io/wasimwazi/kubernetes-operator-go:latest
```

<b> Step 5: Load the docker image to kind cluster</b>
```
$ kind load docker-image ghcr.io/wasimwazi/kubernetes-operator-go:latest --name <cluster-name>
```

<b> Step 7: Install the CRD</b>

```
$ kubectl apply -f https://raw.githubusercontent.com/wasimwazi/kubernetes-operator-go/master/config/crd/bases/app.cisco.com_nginxoperators.yaml
```


<b> Step 7: Create service account and role for the custom controller</b>
```
$ kubectl apply -f https://raw.githubusercontent.com/wasimwazi/kubernetes-operator-go/workflow/deployment_config/accounts.yaml
```

Above manifest will create a service account, clusterrole and clusterrole binding for the operator to authenticate to the Kubernetes APIs inside ```operator-deployment-namespace``` namespace.

<b> Step 8: Deploy the custom controller</b>
```
$ kubectl apply -f https://raw.githubusercontent.com/wasimwazi/kubernetes-operator-go/workflow/deployment_config/deployment.yaml
```
Applying this manifest will deploy the custom controller inside ```operator-deployment-namespace``` namespace.

<b>Step 9: Apply the custom operator yaml</b>

Create a yaml file which corresponds to the CRD for this project and apply the configuration using ```kubectl apply``` command. 
A sample CRD configuration file present in this repository can be applied.
```
$ kubectl apply -f https://raw.githubusercontent.com/wasimwazi/kubernetes-operator-go/workflow/deployment_config/sample-oper.yaml
```

