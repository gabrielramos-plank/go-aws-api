package items

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"go-aws-api/internal/httpx"
)

type Repository interface {
	List(ctx context.Context) ([]Item, error)
	Create(ctx context.Context, name string) (Item, error)
}

type Handler struct {
	repo   Repository
	logger *slog.Logger
}

func NewHandler(repo Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/items", h.list)
	mux.HandleFunc("POST /api/v1/items", h.create)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		h.logger.Error("list items", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if items == nil {
		items = []Item{}
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

type createRequest struct {
	Name string `json:"name"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createRequest
	if err := dec.Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 200 {
		httpx.WriteError(w, http.StatusBadRequest, "name is required and must be at most 200 characters")
		return
	}

	it, err := h.repo.Create(r.Context(), name)
	if err != nil {
		h.logger.Error("create item", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, it)
}
