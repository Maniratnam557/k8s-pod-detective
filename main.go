package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Maniratnam557/k8s-pod-detective/pkg/detector"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var kubeconfig *string

	// Only offer kubeconfig flag when running outside cluster
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "",
			"absolute path to the kubeconfig file")
	}

	namespace := flag.String("namespace", "default",
		"Kubernetes namespace to monitor")

	flag.Parse()

	// ===== BUILD CONFIG (WORKS BOTH IN-CLUSTER AND OUT-OF-CLUSTER) =====
	config, err := buildConfig(*kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building config: %v\n", err)
		os.Exit(1)
	}

	// Debug output
	fmt.Printf("[DEBUG] ==========================================\n")
	fmt.Printf("[DEBUG] API Server URL: %s\n", config.Host)
	fmt.Printf("[DEBUG] ==========================================\n\n")

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}

	// Print banner
	printBanner()

	// Start detector
	podDetector := detector.New(clientset)
	if err := podDetector.WatchPods(*namespace); err != nil {
		fmt.Fprintf(os.Stderr, "Error watching pods: %v\n", err)
		os.Exit(1)
	}
}

// buildConfig creates Kubernetes config from kubeconfig file or in-cluster config
func buildConfig(kubeconfigPath string) (*rest.Config, error) {
	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		fmt.Println("[INFO] Using in-cluster configuration")
		return config, nil
	}

	// Use kubeconfig - DON'T set ExplicitPath, let it find the file
	fmt.Printf("[INFO] Using kubeconfig file\n")

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// DON'T SET ExplicitPath - let it use default discovery
	// loadingRules.ExplicitPath = kubeconfigPath  ← REMOVE THIS!

	configOverrides := &clientcmd.ConfigOverrides{}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	return kubeConfig.ClientConfig()
}

func printBanner() {
	banner := `
╔══════════════════════════════════════════════╗
║   🔍 Kubernetes Pod Failure Detective       ║
║   Detecting and explaining pod failures     ║
╚══════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
