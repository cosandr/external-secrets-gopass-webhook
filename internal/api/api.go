package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cosandr/external-secrets-gopass-webhook/internal/gopass"
	log "github.com/sirupsen/logrus"
)

type PostRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func HandleGetSecret(gp *gopass.Gopass, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the secret name from the 'name' URL query parameter
	secretName := r.URL.Query().Get("name")
	if secretName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing 'name' query parameter"})
		return
	}
	log.Debugf("received GET request for secret '%s'", secretName)
	val, err := gp.GetSecret(r.Context(), secretName)
	if err != nil {
		var notFoundErr *gopass.ErrSecretNotFound
		if errors.As(err, &notFoundErr) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"name": secretName, "error": "Secret not found"})
		} else {
			log.Errorf("error retrieving secret '%s': %v", secretName, err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"name": secretName, "error": "An error occurred retrieving secret"})
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"name": secretName, "value": val})
	log.Debugf("completed GET request for secret '%s'", secretName)
}

func HandlePostSecret(gp *gopass.Gopass, pushEnabled bool, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !pushEnabled {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Push is not enabled, set GIT_PUSH_ENABLED=true"})
		return
	}

	var input PostRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&input)
	if err != nil {
		log.Errorf("error parsing request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse request body"})
		return
	}
	log.Debugf("parsed POST request: %v", input)
	if input.Name == "" || input.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Body missing 'name' and/or 'value'"})
		return
	}
	// TODO: Implement pushing
}
