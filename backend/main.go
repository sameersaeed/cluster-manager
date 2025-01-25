package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Cluster struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

func handleCORS(next http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000"}), 
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(next)
}

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

func getClusterNameHandler(w http.ResponseWriter, r *http.Request) {
    clusterName, err := getClusterName()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"clusterName": clusterName})
}

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

func getNodeDetailsHandler(w http.ResponseWriter, r *http.Request) {
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
			"name":		node.Name,
			"cpu":      node.Status.Capacity["cpu"],
			"memory":   node.Status.Capacity["memory"],
			"status":   status, 
		})

	json.NewEncoder(w).Encode(map[string][]map[string]interface{}{"nodes": nodeDetails})
	}
}

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

func getNamespacesHandler(w http.ResponseWriter, r *http.Request) {
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

func getDeploymentsHandler(w http.ResponseWriter, r *http.Request) {
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

func createDeploymentHandler(w http.ResponseWriter, r *http.Request) {
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

func deleteDeploymentHandler(w http.ResponseWriter, r *http.Request) {
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

func getPods(namespace string) ([]corev1.Pod, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	podClient := clientset.CoreV1().Pods(namespace)
	pods, err := podClient.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func getPodsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]

	if namespace == "" {
		http.Error(w, "Namespace is required", http.StatusBadRequest)
		return
	}

	pods, err := getPods(namespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var podDetails []map[string]interface{}
	for _, pod := range pods {
		status := "Unknown"
		if pod.Status.Phase != "" {
			status = string(pod.Status.Phase)
		}

		podDetails = append(podDetails, map[string]interface{}{
			"name":   pod.Name,
			"status": status,
		})
	}

	json.NewEncoder(w).Encode(map[string][]map[string]interface{}{"pods": podDetails})
}

func createPod(namespace string, pod *corev1.Pod) error {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	podClient := clientset.CoreV1().Pods(namespace)
	_, err = podClient.Create(context.Background(), pod, metav1.CreateOptions{})
	return err
}

func createPodHandler(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	namespace := vars["namespace"]

	podYaml, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	var pod corev1.Pod 
	err = yaml.UnmarshalStrict([]byte(podYaml), &pod)
	if err != nil {
		log.Fatalf("Error unmarshalling YAML: %s", err.Error())
	}

	err = createPod(namespace, &pod)
	if err != nil {
		log.Fatalf("Error creating pod: %s", err.Error())
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"podName": pod.Name,
	})
}

func editPodYaml(namespace string, pod *corev1.Pod) error {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	podClient := clientset.CoreV1().Pods(namespace)
	//podClient.Delete(context.TODO(), pod.Name, metav1.DeleteOptions{GracePeriodSeconds: nil})
	//_, err = podClient.Update(context.Background(), pod, metav1.CreateOptions{})
	_, err = podClient.Update(context.Background(), pod, metav1.UpdateOptions{})
	return err
}

func editPodYamlHandler(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	namespace := vars["namespace"]

	podYaml, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	var pod corev1.Pod
	err = yaml.UnmarshalStrict(podYaml, &pod)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error unmarshalling YAML: %s", err.Error()), http.StatusBadRequest)
		return
	}

	err = editPodYaml(namespace, &pod)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update pod: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"podName": pod.Name,
	})
}

func deletePod(namespace, podName string) error {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	podClient := clientset.CoreV1().Pods(namespace)
	deleteOptions := metav1.DeleteOptions{}
	err = podClient.Delete(context.Background(), podName, deleteOptions)
	return err
}

func deletePodHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	podName := vars["podName"]

	if namespace == "" || podName == "" {
		http.Error(w, "Namespace and podName are required", http.StatusBadRequest)
		return
	}

	err := deletePod(namespace, podName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "deletedPod": podName})
}

func getPodLogs(namespace, podName string) (string, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return "", err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	podClient := clientset.CoreV1().Pods(namespace)
	logOptions := &corev1.PodLogOptions{}
	logs, err := podClient.GetLogs(podName, logOptions).DoRaw(context.Background())
	if err != nil {
		return "", err
	}

	return string(logs), nil
}

func getPodLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	podName := vars["podName"]

	if namespace == "" || podName == "" {
		http.Error(w, "Namespace and podName are required", http.StatusBadRequest)
		return
	}

	logs, err := getPodLogs(namespace, podName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"logs": logs})
}

func getPodYamlHandler(w http.ResponseWriter, r *http.Request) {
	namespace := mux.Vars(r)["namespace"]
	podName := mux.Vars(r)["podName"]

	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to build config from path: %v", kubeconfig), http.StatusInternalServerError)
		return 
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Kubernetes client: %v", err), http.StatusInternalServerError)
		return
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get pod: %v", err), http.StatusInternalServerError)
		return
	}

	yamlData, err := yaml.Marshal(pod)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal pod to YAML: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(yamlData)
}


func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/cluster-name", getClusterNameHandler).Methods("GET")
	router.HandleFunc("/api/node-details", getNodeDetailsHandler).Methods("GET")
	router.HandleFunc("/api/namespaces", getNamespacesHandler).Methods("GET")

	router.HandleFunc("/api/deployments/{namespace}", getDeploymentsHandler).Methods("GET") 
	router.HandleFunc("/api/deployment/{namespace}/{deploymentName}", createDeploymentHandler).Methods("POST")  
	router.HandleFunc("/api/deployment/{namespace}/{deploymentName}", deleteDeploymentHandler).Methods("DELETE")  

	router.HandleFunc("/api/pods/{namespace}", getPodsHandler).Methods("GET") 
	router.HandleFunc("/api/pod/{namespace}/{podName}", createPodHandler).Methods("POST")
	router.HandleFunc("/api/pod/{namespace}/{podName}", editPodYamlHandler).Methods("PUT")
	router.HandleFunc("/api/pod/{namespace}/{podName}", deletePodHandler).Methods("DELETE")
	router.HandleFunc("/api/pod/{namespace}/{podName}/yaml", getPodYamlHandler).Methods("GET")
	router.HandleFunc("/api/pod/{namespace}/{podName}/logs", getPodLogsHandler).Methods("GET")

	http.ListenAndServe(":8080", handleCORS(router))
}