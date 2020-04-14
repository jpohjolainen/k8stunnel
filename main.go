package main

import (
  "os"
  "fmt"
  "time"
  "math/rand"
	"path/filepath"

  "golang.org/x/net/context"

  cli "github.com/urfave/cli/v2"

  "k8s.io/client-go/tools/clientcmd"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/util/homedir"

)

var (
  kubeconfig string
  namespace string

  cliFlags = []cli.Flag {
    &cli.StringFlag{
      Name: "kubeconfig",
      Aliases: []string{"k"},
      Value: filepath.Join(homedir.HomeDir(), ".kube", "config"),
      Usage: "Path to kubeconfig",
      EnvVars: []string{"KUBECONFIG"},
      Destination: &kubeconfig,
    },
    &cli.StringFlag{
      Name: "namespace",
      Aliases: []string{"n"},
      Value: "default",
      Usage: "Namespace in K8s to deploy Socat",
      Destination: &namespace,
    },
  }
)

func init() {
  	rand.Seed(time.Now().UnixNano())
}

func main() {
  tunnel := k8sTunnel{}

  app := &cli.App{
    Name: "k8stunnel",
    Version: "1.0.0",
    Compiled: time.Now(),
    Usage: "Create tunnel through K8s",
    UsageText: "k8stunnel [options] <host> <port> [localport]",
    EnableBashCompletion: true,
    Commands: []*cli.Command{},
    Flags: cliFlags,
    Action: func(c *cli.Context) error {
      if c.NArg() < 2 {
        fmt.Printf("You need to specify atleast 2 arguments.\n")
        os.Exit(2)
      }

      fmt.Println(c.NArg())

      tunnel.destinationHost = c.Args().Get(0)
      dPort := parseInt32OrExit(c.Args().Get(1))
      tunnel.destinationPort = int32(dPort)
      tunnel.containerPort = randomHighPort()

      if c.NArg() == 3 {
        lPort := parseInt32OrExit(c.Args().Get(2))
        tunnel.localPort = int32(lPort)
      } else {
        if dPort < 1000 {
          tunnel.localPort = randomHighPort()
        } else {
          tunnel.localPort = tunnel.destinationPort
        }
      }
      return nil
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    panic(err)
  }

  if tunnel.destinationHost == "" {
    os.Exit(2)
  }

  tunnel.ctx = context.Background()

	tunnel.config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
  if err != nil {
		fmt.Printf("Error: %s\n", err)
    os.Exit(2)
	}

	if tunnel.client, err = kubernetes.NewForConfig(tunnel.config); err != nil {
		panic(err)
	}

  tunnel.start()
}
