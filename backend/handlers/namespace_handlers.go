package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getNamespaces() ([]corev1.Namespace, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	namespaceClient := clientset.CoreV1().Namespaces()
	namespaces, err := namespaceClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return namespaces.Items, nil
}

func GetNamespacesHandler(w http.ResponseWriter, r *http.Request) {
	namespaces, err := getNamespaces()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var namespaceNames []string
	for _, namespace := range namespaces {
		namespaceNames = append(namespaceNames, namespace.Name)
	}

	json.NewEncoder(w).Encode(map[string][]string{"namespaces": namespaceNames})
}