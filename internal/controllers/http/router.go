package httpcontroller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/dto"
	"github.com/che1nov/backend-trainee-assignment-autumn-2025/internal/usecases"
)

type RouterConfig struct {
	Message           string
	CreateUserUseCase *usecases.CreateUserUseCase
	Logger            *slog.Logger
}

func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(cfg.Message))
	})

	if cfg.CreateUserUseCase != nil {
		r.Post("/users", func(w http.ResponseWriter, r *http.Request) {
			var input dto.UserInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				cfg.Logger.Error("failed to decode user", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if input.ID == "" || input.Name == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := cfg.CreateUserUseCase.Execute(r.Context(), input.ID, input.Name); err != nil {
				cfg.Logger.Error("failed to create user", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(dto.UserOutput{ID: input.ID, Name: input.Name})
		})
	}

	return r
}
