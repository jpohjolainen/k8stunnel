# k8stunnel

K8stunnel creates a tunnel through Kubernetes. It can be used to test
connections from Kubernetes cluster.

## Details

SoCat image is deployed to Kubernetes cluster and then port forward is created
to it. This is basicly same as "kubectl run" and "kubectl port-forward" but
in a nicer package that can be build for different OS.

## Usage

Download binary to your correct OS from the [releases](https://github.com/jpohjolainen/k8stunnel/releases).

```
NAME:
   k8stunnel - Create tunnel through K8s

USAGE:
   k8stunnel [options] <host> <port> [localport]

VERSION:
   1.0.0

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --kubeconfig value, -k value  Path to kubeconfig (default: "/home/messis/.kube/config") [$KUBECONFIG]
   --namespace value, -n value   Namespace in K8s to deploy Socat (default: "default")
   --help, -h                    show help (default: false)
   --version, -v                 print the version (default: false)
```

Connecting to www.google.com though Kubernetes cluster:

```bash

$ k8stunnel www.google.com 80 1080
Deploying 'k8stunnel-vdrkz' with tunnel to 'www.google.com:80'...done.

Ready to receive traffic to localhost:1080
Press CTRL-C to quit..

## On another terminal
$ curl localhost:1080
<!DOCTYPE html>
<html lang=en>
...

## Pressing ctrl-c
^CDeleting 'k8stunnel-vdrkz'...done.
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
