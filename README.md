# k8stunnel

K8stunnel creates a tunnel through Kubernetes. It can be used to test
connections from Kubernetes cluster.

## Details

SoCat image is deployed to Kubernetes cluster and then port forward is created
to it. This is basicly same as "kubectl run" and "kubectl port-forward" but
in a nicer package that can be build for different OS.

## Usage

Download binary to your correct OS from the releases.

Connecting to www.google.com though Kubernetes cluster:

```bash

$ k8stunnel www.google.com 80 1080
...

# On another terminal
$ curl localhost:1080
...
```

## Build

### Requirements

You need to have Go installed.

### Building

* Clone this repo.
* Execute:

```bash

$ go build -o k8stunnel
```
