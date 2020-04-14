package main

import (
  "os"
  "fmt"
  "sync"
  "time"
  "strings"
  "syscall"
  "net/url"
  "net/http"
  "os/signal"

  "golang.org/x/net/context"

  apiv1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

  "k8s.io/cli-runtime/pkg/genericclioptions"

  "k8s.io/client-go/rest"
  "k8s.io/client-go/informers"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/cache"
  "k8s.io/client-go/transport/spdy"
  "k8s.io/client-go/tools/portforward"

  // "k8s.io/apimachinery/pkg/util/wait"

)

type k8sTunnel struct {
  ctx context.Context
  destinationHost string
  destinationPort int32
  containerPort int32
  localPort int32
  podName string
  config *rest.Config
  client *kubernetes.Clientset
  stopCh chan struct{}
}

var (
  soCatPod = apiv1.Pod{
    ObjectMeta: metav1.ObjectMeta{
      Name: "",
      Namespace: "",
      Labels: map[string]string{
        "app": "",
      },
    },
    Spec: apiv1.PodSpec{
      Containers: []apiv1.Container{
        {
          Name:  "socat",
          Image: "alpine/socat:1.7.3.4-r0",
          Command: []string{
            "socat",
          },
          Args: []string{},
          Ports: []apiv1.ContainerPort{
            {
              Name:          "tunnel",
              Protocol:      apiv1.ProtocolTCP,
              ContainerPort: 0,
            },
          },
        },
      },
    },
  }
)

func (tunnel *k8sTunnel) start() {
  tunnel.deploy()
  defer tunnel.delete()

  tunnel.startPortForward()
}

func (tunnel *k8sTunnel) deploy() {
  suffix := randomChars(5)
  tunnel.podName = fmt.Sprintf("k8stunnel-%s", suffix)

  soCatPod.ObjectMeta.Name = tunnel.podName
  soCatPod.ObjectMeta.Namespace = namespace
  soCatPod.ObjectMeta.Labels["app"] = tunnel.podName
  soCatPod.Spec.Containers[0].Args = []string{
    fmt.Sprintf("TCP-LISTEN:%d,fork", tunnel.containerPort),
    fmt.Sprintf("TCP:%s:%d", tunnel.destinationHost, tunnel.destinationPort),
  }
  soCatPod.Spec.Containers[0].Ports[0].ContainerPort = tunnel.containerPort

  podClient := tunnel.client.CoreV1().Pods(namespace)

  // fmt.Printf("%#v\n", soCatPod)

  fmt.Printf("Deploying '%s' with tunnel to '%s:%d'...", tunnel.podName,
    tunnel.destinationHost, tunnel.destinationPort)

  _, err := podClient.Create(tunnel.ctx, &soCatPod, metav1.CreateOptions{})
  if err != nil {
    fmt.Printf("\nError: %v\n", err)
    os.Exit(4)
  }
  // fmt.Printf("%#d\n", pod)

  tunnel.watchPod("created")
}

func (tunnel *k8sTunnel) delete() {
  fmt.Printf("Deleting '%s'...", tunnel.podName)
  podClient := tunnel.client.CoreV1().Pods(namespace)
  deletePolicy := metav1.DeletePropagationForeground
  err := podClient.Delete(tunnel.ctx, tunnel.podName, metav1.DeleteOptions{
    PropagationPolicy: &deletePolicy,
  })
  if err != nil {
    panic(err)
  }
  tunnel.watchPod("deleted")
}

func (tunnel *k8sTunnel) startPortForward() {
	var wg sync.WaitGroup

	wg.Add(1)

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})
	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		close(stopCh)
		wg.Done()
	}()

	go func() {
    transport, upgrader, err := spdy.RoundTripperFor(tunnel.config)
    if err != nil {
      panic(err)
    }

    path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
      namespace, tunnel.podName)
    host := strings.TrimLeft(tunnel.config.Host, "https://")
    dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport},
      http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: host})

    ports := fmt.Sprintf("%d:%d", tunnel.localPort, tunnel.containerPort)
    fw, err := portforward.New(dialer, []string{ports}, stopCh, readyCh,
      nil, stream.ErrOut)
    if err != nil {
      panic(err)
    }

    fw.ForwardPorts()
	}()

  select {
    case <-readyCh:
      fmt.Printf("\nReady to receive traffic to localhost:%d\n", tunnel.localPort)
      fmt.Println("Press CTRL-C to quit..")
      break
  }

  wg.Wait()
}

func (tunnel *k8sTunnel) watchPod(w string) {
  stopCh := make(chan struct{}, 1)

  handlers := cache.ResourceEventHandlerFuncs{}
  if w == "created" {
    handlers = cache.ResourceEventHandlerFuncs{
      AddFunc: func(obj interface{}) {
        pod := obj.(*apiv1.Pod)
        if pod.ObjectMeta.Name == tunnel.podName {
          fmt.Println("done.")
          close(stopCh)
        }
      },
    }
  } else {
    handlers = cache.ResourceEventHandlerFuncs{
      DeleteFunc: func(obj interface{}) {
        pod := obj.(*apiv1.Pod)
        if pod.ObjectMeta.Name == tunnel.podName {
          fmt.Println("done.")
          close(stopCh)
        }
      },
    }
  }
	resyncPeriod := 20 * time.Second
	si := informers.NewSharedInformerFactory(tunnel.client, resyncPeriod)
	si.Core().V1().Pods().Informer().AddEventHandler(handlers)
	si.Start(stopCh)

  <- stopCh
}
