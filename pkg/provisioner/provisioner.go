package provisioner

import (
	"flag"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	driver                        = "stevenklar/hetzner-cloud-driver"
	driverOptionToken             = "stevenklar/hetzner-cloud-driver/token"
)

var (
	kubeconfig  = flag.String("kubeconfig", "", "Absolute path to the kubeconfig file. Either this or master needs to be set if the provisioner is being run out of cluster.")
	master      = flag.String("master", "", "Master URL to build a client config from. Either this or kubeconfig needs to be set if the provisioner is being run out of cluster.")
	provisioner = flag.String("provisioner", "stevenklar/hetzner-storage-provisioner", "Name of the provisioner. The provisioner will only provision volumes for claims that request a StorageClass with a provisioner field set equal to this name.")
)

// Run executes the provisioner routine
func Run() {
	flag.Parse()
	err := flag.Set("logtostderr", "true")
	if err != nil {
		glog.Fatalf("Failed to set flag: %v", err)
	}

	var config *rest.Config
	if *master != "" || *kubeconfig != "" {
		glog.Infof("Either master or kubeconfig specified. Building kube config from that...")
		config, err = clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	} else {
		glog.Infof("Building kube configs for running in cluster...")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		glog.Fatalf("Failed to create config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("Failed to create client: %v", err)
	}

	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		glog.Fatalf("Error getting server version: %v", err)
	}

	hetznerProvisioner := NewProvisioner()
	pc := controller.NewProvisionController(
		clientset,
		*provisioner,
		hetznerProvisioner,
		serverVersion.GitVersion,
	)

	pc.Run(wait.NeverStop)
}
