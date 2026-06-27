package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"connectionpool-sandbox/usecase"
)

type Handler struct {
	withPool    *usecase.RecordUsecase
	withoutPool *usecase.RecordUsecase
}

func New(withPool, withoutPool *usecase.RecordUsecase) *Handler {
	return &Handler{
		withPool:    withPool,
		withoutPool: withoutPool,
	}
}

type response struct {
	Mode       string  `json:"mode"`
	DurationMs float64 `json:"duration_ms"`
	Error      string  `json:"error,omitempty"`
}

func (h *Handler) WithPool(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	err := h.withPool.Execute(r.Context())
	write(w, "with-pool", start, err)
}

func (h *Handler) WithoutPool(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	err := h.withoutPool.Execute(r.Context())
	write(w, "without-pool", start, err)
}

func write(w http.ResponseWriter, mode string, start time.Time, err error) {
	resp := response{
		Mode:       mode,
		DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
	}
	if err != nil {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
