package redirect

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/spinelli/encurtador-links/shared/dto"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ServeHTTP godoc
// @Summary Redirecionar URL curta
// @Description Recebe um código curto e redireciona (302) para a URL original
// @Tags redirect
// @Param shortCode path string true "Código curto da URL"
// @Success 302 "Redirect para URL original"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /{shortCode} [get]
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	shortCode := strings.TrimPrefix(r.URL.Path, "/")

	if shortCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "código não informado"})
		return
	}

	longURL, err := h.service.Execute(r.Context(), RedirectQuery{ShortCode: shortCode})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "URL não encontrada"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "erro interno"})
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}
