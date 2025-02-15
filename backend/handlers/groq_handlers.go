package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func SendGroqQueryHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		http.Error(w, "could not find value for 'GROQ_API_KEY'", http.StatusInternalServerError)
		return
	}

	var request map[string]string
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "request format incorrect", http.StatusBadRequest)
		return
	}

	yamlType, existsYaml := request["yamlType"]
	query, existsQuery := request["query"]
	if !existsYaml || !existsQuery {
		http.Error(w, "yamlType and query parameters are required in the request", http.StatusBadRequest)
		return
	}

	payload, err := createPayload(yamlType, query)
	if err != nil {
		http.Error(w, "could not create payload", http.StatusInternalServerError)
		return
	}

	groqURL := "https://api.groq.com/openai/v1/chat/completions"
	req, err := http.NewRequest("POST", groqURL, bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, "could not create request", http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "could not make request to Groq API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}

func createPayload(yamlType string, query string) ([]byte, error) {
	requestBody := map[string]interface{}{
		"model": "llama-3.3-70b-versatile",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": "only provide the yaml, and no extra text / formatting (i.e. ```) for the following - create a " + yamlType + " yaml for: " + query + ". if you are unsure, just provide a basic sample yaml.",
			},
		},
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
