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
