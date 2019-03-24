// Package health provides a endpoint to query the status and readiness
// of the Poppin API
package health

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/go-chi/render"

	"github.com/go-chi/chi"
	"github.com/hackathon/hackhub/pkg/database"
)

// Health is the API response to health queries
type Health struct {
	Version        string `json:"version"`
	DatabaseStatus bool   `json:"database"`
	Ready          bool   `json:"ready"`
}

// Handler contains the chi.Mux, logrus.Logger, database.DB type object, and Version number
type Handler struct {
	Router  *chi.Mux
	Logger  *log.Logger
	DB      database.DB
	Version string
}

// Routes creates routes and registers the routes for the health endpoint
// It returns a handler with the needed dependancies passed from the caller
// or generates dependancies
func Routes(logger *log.Logger, db database.DB, version string) *Handler {
	router := chi.NewRouter()

	handler := &Handler{router, logger, db, version}

	// Routes for the location namespace
	router.Get("/", handler.Health)
	router.Handle("/metrics", promhttp.Handler())

	return handler
}

// Health returns the health of the application and status code to the caller
// This enables such things as readiness and liviness endpoints
// The Applcation is only ready and healthy if it can talk to the DB
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	resp := Health{}
	resp.DatabaseStatus = h.checkDbHealth()
	resp.Version = h.Version
	resp.Ready = h.ready(&resp)

	if !resp.Ready {
		render.Status(r, 500)
	}

	render.JSON(w, r, resp)
}

func (h *Handler) ready(resp *Health) bool {
	if !resp.DatabaseStatus {
		return false
	}
	return true
}

func (h *Handler) checkDbHealth() bool {
	// Check the connection
	if err := h.DB.Ping(context.TODO(), nil); err != nil {
		h.Logger.Errorf("database down: %s", err)
		return false
	}
	return true
}
