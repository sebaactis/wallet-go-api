package health

import (
	"encoding/json"
	"net/http"
)

type HealthHandler struct{}

func New() *HealthHandler { return &HealthHandler{} }

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"OK": true})
}
