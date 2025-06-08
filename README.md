# PodLogMonitor Controller

## Overview

The `PodLogMonitor` controller is a Kubernetes operator that automates the process of monitoring pod logs for specific error messages and restarting pods when those messages are found. This can be useful for automatically handling errors or failures in pods by ensuring they are restarted based on predefined log patterns.

## Features

- **Monitor logs** of pods in a specific namespace.
- **Trigger restarts** for pods when logs match a specified message.
- **Custom Resource Definition (CRD)** to define `PodLogMonitor` resources in the cluster.
- **Leader Election** to ensure a single active controller instance.

## Prerequisites

- A Kubernetes cluster.
- `kubectl` and `go` installed on your local machine.
- The ability to create Kubernetes resources in your cluster.

## Installation

### 1. Install the Custom Resource Definition (CRD)

First, apply the CRD to your Kubernetes cluster. This will enable the `PodLogMonitor` custom resource.

```bash
kubectl apply -f crd.yml
```

### 2. Create the Custom Resource

Create the `PodLogMonitor` custom resource which is a cluster level object where the namespace and error message to look for(by the controller) is defined

```bash
kubectl apply -f cr.yml
```

### 3. Apply RBAC (Service account , Cluster role and Cluster rolebinding)

Apply the RBAC to your cluster so the controller pods have the needed privelages on the custom resource and pods
NOTE: You may want to change of the namespace where the service account is created if you want to deploy your controller in a different namespace.

```bash
kubectl apply -f rbac.yml
```

### 4. Deploy the controller

Create an image using the Dockerfile to deploy the controller
NOTE: You may want to change of the namespace if you want to deploy your controller in a different namespace.
```bash
docker build -t pod-log-monitor-controller:latest .
kubectl create deployment pod-log-monitor-controller --image pod-log-monitor-controller:latest -n default
```



