package handler

import (
	"encoding/json"
	"math"
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

func (h *Handler) Heavy(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	count := sieveOfEratosthenes(1_000_000)
	resp := struct {
		Mode       string  `json:"mode"`
		DurationMs float64 `json:"duration_ms"`
		PrimeCount int     `json:"prime_count"`
	}{
		Mode:       "heavy",
		DurationMs: float64(time.Since(start).Microseconds()) / 1000.0,
		PrimeCount: count,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// sieveOfEratosthenes returns the count of primes up to n.
// Intentionally named for easy identification in Pyroscope flame graphs.
func sieveOfEratosthenes(n int) int {
	composite := make([]bool, n+1)
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if !composite[i] {
			for j := i * i; j <= n; j += i {
				composite[j] = true
			}
		}
	}
	count := 0
	for i := 2; i <= n; i++ {
		if !composite[i] {
			count++
		}
	}
	return count
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
