package main

import (
	"flag"
	"log"
	"time"

	kubeflow "github.com/StatCan/kubeflow-controller/pkg/generated/clientset/versioned"
	informers "github.com/StatCan/kubeflow-controller/pkg/generated/informers/externalversions"
	"github.com/statcan/profile-state-controller/pkg/controller"
	"github.com/statcan/profile-state-controller/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.Parse()
}

func main() {
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("error building kubeconfig: %v", err)
	}

	kubeclient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("error building kubernetes clientset: %v", err)
	}

	kubeflowclient, err := kubeflow.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("error building kubeflow client: %v", err)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeclient, time.Minute*5)
	kubeflowInformerFactory := informers.NewSharedInformerFactory(kubeflowclient, time.Minute*5)

	ctlr := controller.NewController(
		kubeclient,
		kubeflowclient,
		kubeflowInformerFactory.Kubeflow().V1().Profiles(),
		kubeInformerFactory.Core().V1().Pods(),
		kubeInformerFactory.Rbac().V1().RoleBindings(),
	)

	kubeInformerFactory.Start(stopCh)
	kubeflowInformerFactory.Start(stopCh)

	if err = ctlr.Run(2, stopCh); err != nil {
		log.Fatalf("error running controller: %v", err)
	}
}
