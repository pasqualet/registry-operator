# Kubernetes

The Registry Operator can be deployed in any Kubernetes distribution.

## Kubectl

```bash
kubectl apply -f https://...
```

## Helm

Add the Helm Repository:

```bash
helm repo add astrokube ...
```

Install the Chart:

```bash
helm install registry-operator astrokube/registry-operator
```
