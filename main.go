package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/vault/api"
)

var vaultClient *api.Client

func initVaultClient() *api.Client {
	vaultURL := os.Getenv("VAULT_SECRET_URL")
	vaultToken := os.Getenv("VAULT_SECRET_TOKEN")

	if vaultToken == "" {
		log.Fatal("VAULT_PARTNER_SECRET_TOKEN is not set in environment variables")
	}

	config := &api.Config{
		Address: vaultURL,
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}

	client.SetToken(vaultToken)
	return client
}

func getSecretsHandler(w http.ResponseWriter, r *http.Request) {
	secret, err := vaultClient.Logical().Read("secret/data/mysecret")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading from Vault: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Secret data map: %+v", secret.Data)

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		http.Error(w, "Could not parse data map", http.StatusInternalServerError)
		return
	}

	username, okUsername := data["username"].(string)
	password, okPassword := data["password"].(string)

	if !okUsername || !okPassword {
		http.Error(w, "Could not parse secrets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"username":"%s","password":"%s"}`, username, password)
}

func main() {
	vaultClient = initVaultClient()

	http.HandleFunc("/secrets", getSecretsHandler)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
