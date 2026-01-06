package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alifpay/temporal/models"
)

// payHandler accepts a JSON POST with payment data and starts a Temporal workflow.
// Example request body:
// {"ID":"123","Currency":"USD","Amount":100.5,"PayDate":"2025-01-20T00:00:00Z"}
func PayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var p models.Payment
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&p); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}
	if p.ID == "" {
		p.ID = fmt.Sprintf("payment-%d", time.Now().UnixNano())
	}

	startPaymentWorkflow(r.Context(), p)

	resp := map[string]string{
		"id":     p.ID,
		"status": "pending",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(resp)
}
