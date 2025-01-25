package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/yaml"
)

func getDeployments(namespace string) ([]appsv1.Deployment, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	deploymentClient := clientset.AppsV1().Deployments(namespace)
	deployments, err := deploymentClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return deployments.Items, nil
}

func GetDeploymentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	if namespace == "" {
		http.Error(w, "namespace is required", http.StatusBadRequest)
		return
	}

	deployments, err := getDeployments(namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var deploymentNames []string
	for _, deployment := range deployments {
		deploymentNames = append(deploymentNames, deployment.Name)
	}

	json.NewEncoder(w).Encode(map[string][]string{"deployments": deploymentNames})
}

func createDeployment(namespace string, deployment *appsv1.Deployment) error {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	deploymentClient := clientset.AppsV1().Deployments(namespace)
	_, err = deploymentClient.Create(context.Background(), deployment, metav1.CreateOptions{})
	return err
}

func CreateDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	deploymentYaml, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	var deployment appsv1.Deployment
	err = yaml.UnmarshalStrict(deploymentYaml, &deployment)
	if err != nil {
		log.Fatalf("Error unmarshalling YAML: %s", err.Error())
		http.Error(w, fmt.Sprintf("Failed to unmarshal YAML: %v", err), http.StatusBadRequest)
		return
	}

	err = createDeployment(namespace, &deployment)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, fmt.Sprintf("Deployment with name '%s' already exists in namespace '%s'", deployment.Name, namespace), http.StatusConflict)
			return
		}

		http.Error(w, fmt.Sprintf("Error creating deployment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "deploymentName": deployment.Name})
}

func deleteDeployment(namespace, deploymentName string) error {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	deploymentClient := clientset.AppsV1().Deployments(namespace)
	deleteOptions := metav1.DeleteOptions{}
	err = deploymentClient.Delete(context.Background(), deploymentName, deleteOptions)
	return err
}

func DeleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	deploymentName := vars["deploymentName"]

	if namespace == "" || deploymentName == "" {
		http.Error(w, "Namespace and deploymentName are required", http.StatusBadRequest)
		return
	}

	err := deleteDeployment(namespace, deploymentName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "success", "deploymentName": deploymentName})
}
