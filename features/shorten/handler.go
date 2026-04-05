package shorten

import (
	"encoding/json"
	"net/http"

	"github.com/spinelli/encurtador-links/shared/dto"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ServeHTTP godoc
// @Summary Encurtar URL
// @Description Recebe uma URL longa e retorna uma URL curta
// @Tags shorten
// @Accept json
// @Produce json
// @Param body body ShortenCommand true "URL para encurtar"
// @Success 201 {object} ShortenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 405 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /shorten [post]
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "método não permitido"})
		return
	}

	var cmd ShortenCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "JSON inválido"})
		return
	}

	if cmd.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "campo 'url' é obrigatório"})
		return
	}

	resp, err := h.service.Execute(r.Context(), cmd)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(dto.ErrorResponse{Error: "erro interno", Details: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
