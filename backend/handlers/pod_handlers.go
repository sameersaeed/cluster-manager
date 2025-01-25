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
	"time"

	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"sigs.k8s.io/yaml"
)

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

func GetPodsHandler(w http.ResponseWriter, r *http.Request) {
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

func CreatePodHandler(w http.ResponseWriter, r *http.Request) {
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
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, fmt.Sprintf("Pod with name '%s' already exists in namespace '%s'", pod.Name, namespace), http.StatusConflict)
			return
		}

		http.Error(w, fmt.Sprintf("Error creating pod: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"podName": pod.Name,
	})
}

func EditPodYamlHandler(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)
	namespace := vars["namespace"]
	podName := vars["podName"]

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

	// certain pod properties are immutable
	// to get around this we first delete the pod
	err = deletePod(namespace, podName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "deletedPod": podName})
	time.Sleep(2 * time.Second)

	// then we recreate the pod with the provided yaml
	err = createPod(namespace, &pod)
	if err != nil {
		log.Fatalf("Error updating pod: %s", err.Error())
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

func DeletePodHandler(w http.ResponseWriter, r *http.Request) {
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

func GetPodLogsHandler(w http.ResponseWriter, r *http.Request) {
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

func GetPodYamlHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, fmt.Sprintf("Failed to marshal pod spec to YAML: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(yamlData)
}