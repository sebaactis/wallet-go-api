package token

import (
	"net/http"

	"github.com/sebaactis/wallet-go-api/internal/httputil"
)

type HTTPHandler struct {
	service *Service
}

func NewHTTPHandler(service *Service) *HTTPHandler {
	return &HTTPHandler{service: service}
}

func (h *HTTPHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tokens, err := h.service.GetAll(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalida request", nil)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, ToResponseMany(tokens))

}
