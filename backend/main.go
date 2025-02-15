package main

import (
	corshandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	handlers "github.com/sameersaeed/cluster-manager/handlers"
)

type Cluster struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

func handleCORS(next http.Handler) http.Handler {
	return corshandlers.CORS(
		corshandlers.AllowedOrigins([]string{"http://localhost:3000"}),
		corshandlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		corshandlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(next)
}

func main() {
	router := mux.NewRouter()

	// try loading from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("couldn't find any .env file, using system env vars")
	}

	// if .env not found, try loading from system env vars
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		log.Fatal("missing GROQ_API_KEY value")
	}

	router.HandleFunc("/api/cluster-name", handlers.GetClusterNameHandler).Methods("GET")
	router.HandleFunc("/api/node-details", handlers.GetNodeDetailsHandler).Methods("GET")
	router.HandleFunc("/api/namespaces", handlers.GetNamespacesHandler).Methods("GET")

	router.HandleFunc("/api/deployments/{namespace}", handlers.GetDeploymentsHandler).Methods("GET")
	router.HandleFunc("/api/deployment/{namespace}/{deploymentName}", handlers.CreateDeploymentHandler).Methods("POST")
	router.HandleFunc("/api/deployment/{namespace}/{deploymentName}", handlers.DeleteDeploymentHandler).Methods("DELETE")

	router.HandleFunc("/api/pods/{namespace}", handlers.GetPodsHandler).Methods("GET")
	router.HandleFunc("/api/pod/{namespace}/{podName}", handlers.CreatePodHandler).Methods("POST")
	router.HandleFunc("/api/pod/{namespace}/{podName}", handlers.EditPodYamlHandler).Methods("PUT")
	router.HandleFunc("/api/pod/{namespace}/{podName}", handlers.DeletePodHandler).Methods("DELETE")
	router.HandleFunc("/api/pod/{namespace}/{podName}/yaml", handlers.GetPodYamlHandler).Methods("GET")
	router.HandleFunc("/api/pod/{namespace}/{podName}/logs", handlers.GetPodLogsHandler).Methods("GET")

	router.HandleFunc("/api/groq", handlers.SendGroqQueryHandler).Methods("POST")

	http.ListenAndServe(":8080", handleCORS(router))
}
