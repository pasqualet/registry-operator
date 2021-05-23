# Introdution

The Registry Operator provides [Kubernetes](https://kubernetes.io) integration with Container Registries. The purpose of this project is to simplify and automate the configuration of external Container Registries for Kubernetes clusters.

The Registry Operator include the following features:
* Integrate external Container Registries with declarative configuration
* Automatic injection of ImagePullSecrets in the Pods

## Compatibility

The Registry Operator is compatible with Kubernetes and OpenShift clusters.

| Distribution | Min version | Max version |
| --- | --- | --- |
| Kubernetes | 1.16 | 1.20 |
| OpenShift | 4.1 | 4.7 |
