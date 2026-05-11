package handlers

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/cQu1x/Auto-Scaler/internal/util"
)

type Loader interface {
	Load(ctx context.Context, amount int) error
}

type LoaderHandler struct {
	Loader Loader
}

func NewLoaderHandler(loader Loader) *LoaderHandler {
	return &LoaderHandler{Loader: loader}
}

func (h *LoaderHandler) Load(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload WorkRequest
	if err := util.DecodeJSON(r, &payload); err != nil {
		util.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if payload.Duration <= 0 {
		payload.Duration = 10
	}
	if payload.Amount <= 0 {
		payload.Amount = runtime.NumCPU()
	}
	if payload.Amount > 64 {
		util.WriteJSONError(w, "amount must be between 1 and 64", http.StatusBadRequest)
		return
	}
	if payload.Duration > 300 {
		util.WriteJSONError(w, "duration must be between 1 and 300 seconds", http.StatusBadRequest)
		return
	}
	if h.Loader == nil {
		util.WriteJSONError(w, "loader is not configured", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*time.Duration(payload.Duration))
	defer cancel()
	if err := h.Loader.Load(ctx, payload.Amount); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		util.WriteJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, map[string]any{
		"message":  "load completed",
		"amount":   payload.Amount,
		"duration": payload.Duration,
	}, http.StatusOK)
}

type CPULoader struct{}

func (l CPULoader) Load(ctx context.Context, amount int) error {
	wg := sync.WaitGroup{}
	wg.Add(amount)
	for i := 0; i < amount; i++ {
		go func() {
			defer wg.Done()
			cpuWork(ctx)
		}()
	}
	wg.Wait()
	return ctx.Err()
}

func cpuWork(ctx context.Context) {
	counter := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			counter++
		}
	}
}
