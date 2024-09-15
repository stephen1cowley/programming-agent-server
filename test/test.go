package test

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type inputSchema struct {
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/", helloHandler)

	fmt.Println("Server listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		fmt.Println("Received a POST request!")
		var requestData inputSchema
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, "Error decoding request data:", err)
			return
		}

		// Process data after successful unmarshalling
		responseMessage := "Received POST request with data: " + requestData.Message

		// Same schema as input (for now)
		jsonResponse := inputSchema{Message: responseMessage}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Error encoding JSON response", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed")
	}
}
