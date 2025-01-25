package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func getNodeDetails() ([]corev1.Node, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	nodeClient := clientset.CoreV1().Nodes()
	nodes, err := nodeClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return nodes.Items, nil
}

func GetNodeDetailsHandler(w http.ResponseWriter, r *http.Request) {
	nodes, err := getNodeDetails()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var nodeDetails []map[string]interface{}
	for _, node := range nodes {
		status := "Unready"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady {
				status = string("Ready")
				break
			}
		}

		nodeDetails = append(nodeDetails, map[string]interface{}{
			"name":   node.Name,
			"cpu":    node.Status.Capacity["cpu"],
			"memory": node.Status.Capacity["memory"],
			"status": status,
		})

		json.NewEncoder(w).Encode(map[string][]map[string]interface{}{"nodes": nodeDetails})
	}
}