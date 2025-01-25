package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getClusterName() (string, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return "", fmt.Errorf("kubeconfig file not found at: %s", kubeconfig)
	}

	kubeconfigFile, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return "", err
	}

	currentContext := kubeconfigFile.Contexts[kubeconfigFile.CurrentContext]
	if currentContext == nil {
		return "", fmt.Errorf("no current context set in kubeconfig")
	}

	clusterName := currentContext.Cluster
	return clusterName, nil
}

func GetClusterNameHandler(w http.ResponseWriter, r *http.Request) {
	clusterName, err := getClusterName()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"clusterName": clusterName})
}
