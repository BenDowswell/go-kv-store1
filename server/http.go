package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/BenDowswell/kvstore1/store"
)

const maxBodySize = 1 << 20 // 1 MB

type SetRequest struct {
	Value string `json:"value"`
}

type ResponsePayload struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type handler struct {
	svc *store.KVService
}

func NewHandler(svc *store.KVService) *handler {

	return &handler{
		svc: svc,
	}
}

func (h *handler) Routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.getHealth)
	mux.HandleFunc("GET /kv/{key}", h.getKey)
	mux.HandleFunc("PUT /kv/{key}", h.updateKey)
	mux.HandleFunc("DELETE /kv/{key}", h.deleteKey)
	return mux

}

func StartServer(ctx context.Context, addr string, svc *store.KVService) error {

	h := NewHandler(svc)
	return h.StartServer(ctx, addr)

}

func (h *handler) StartServer(ctx context.Context, addr string) error {

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	srv := &http.Server{
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logger.Info("Starting server", "addr", listener.Addr().String())

	serveErr := make(chan error, 1)
	go func() {
		if err := srv.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()

	select {
	case err := <-serveErr:
		if err != nil {
			logger.Error("server failed", "error", err)
		}
		return err
	case <-ctx.Done():
	}

	logger.Info("stopping server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		logger.Error("forced shutdown", "error", err)
	}
	return err
}

func (h *handler) getHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func (h *handler) getKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value, err := h.svc.Get(key)
	if err != nil {
		if errors.Is(err, store.ErrNoKeyFound) {
			writeJSONError(w, http.StatusNotFound, "key not found")
			return
		}
		if errors.Is(err, store.ErrEmptyKey) {
			writeJSONError(w, http.StatusBadRequest, "Validation failed empty key")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed getting record")
		return
	}

	data := ResponsePayload{
		Key:   key,
		Value: value,
	}

	writeJSON(w, http.StatusOK, data)
}

func (h *handler) updateKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var req SetRequest
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "Error decoding JSON")
		return
	}

	err = h.svc.Set(key, req.Value)

	if err != nil {
		if errors.Is(err, store.ErrEmptyKey) {
			writeJSONError(w, http.StatusBadRequest, "Validation failed empty key")
			return
		}
		if errors.Is(err, store.ErrEmptyValue) {
			writeJSONError(w, http.StatusBadRequest, "Validation failed empty value")
			return
		}
		if errors.Is(err, store.ErrSavingKeyValue) {
			writeJSONError(w, http.StatusInternalServerError, "key not saved")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed saving record")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) deleteKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	err := h.svc.Delete(key)

	if err != nil {
		if errors.Is(err, store.ErrEmptyKey) {
			writeJSONError(w, http.StatusBadRequest, "Validation failed empty key")
			return
		}

		if errors.Is(err, store.ErrDeletingKey) {
			writeJSONError(w, http.StatusInternalServerError, "key not deleted")
			return
		}

		writeJSONError(w, http.StatusInternalServerError, "Failed deleting record")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{Data: data})
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(Envelope{Error: message})
}
