package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/hackathon/hackhub/features/health"
	"github.com/hackathon/hackhub/features/projects"
	"github.com/hackathon/hackhub/features/user"
	"github.com/hackathon/hackhub/pkg/config"
	"github.com/hackathon/hackhub/pkg/database"
	"github.com/hackathon/hackhub/pkg/logs"
	"github.com/mongodb/mongo-go-driver/mongo"
	log "github.com/sirupsen/logrus"
)

// VERSION is passed in as a flag
var (
	VERSION = "0.0.0"
)

// Routes initializes the router with the given middlewares
func Routes(logger *log.Logger, config *config.Configuration, db database.DB) *chi.Mux {
	tokenAuth := jwtauth.New("HS256", []byte(config.Server.JWTKey), nil)
	router := chi.NewRouter()
	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		logs.NewStructuredLogger(logger),
		middleware.DefaultCompress,
		middleware.RedirectSlashes,
		middleware.Recoverer,
	)

	projectRouter := projects.Routes(logger, config, tokenAuth)
	healthRouter := health.Routes(logger, db, VERSION)
	userRouter := user.Routes(logger, config, tokenAuth)

	mainRouter := router.Route("/v1/api", func(r chi.Router) {
		r.Mount("/users", userRouter.Router)
		r.Mount("/projects", projectRouter.Router)
		r.Mount("/sys/health", healthRouter.Router)
	})

	mainRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		routes := make(map[string][]string)
		walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			routes[route] = append(routes[route], method)
			return nil
		}
		if err := chi.Walk(router, walkFunc); err != nil {
			logger.Fatalf("err walking routes: %s", err)
		}

		json.NewEncoder(w).Encode(&routes)
	})

	return router
}

func main() {
	logger := log.New()
	logger.Formatter = &log.JSONFormatter{
		// disable, as we set our own
		DisableTimestamp: true,
	}

	logger.Level = log.DebugLevel
	logger.Formatter = &log.JSONFormatter{DisableTimestamp: true}

	defaults := map[string]interface{}{
		"server": map[string]interface{}{
			"port":   8080,
			"APIKey": "fake-key",
			"JWTKey": "some-signing-key",
		},
		"database": map[string]interface{}{
			"port": 27017,
			"host": "mongo",
		},
	}

	viper, err := config.NewConfig("config", defaults)
	if err != nil {
		logger.Fatalf("err reading config: %s", err)
	}

	var config config.Configuration
	if err := viper.Unmarshal(&config); err != nil {
		logger.Fatalf("err reading config: %s", err)
	}

	db, err := mongo.Connect(
		context.TODO(),
		fmt.Sprintf("mongodb://%s:%d", config.Database.Host, config.Database.Port),
	)
	if err != nil {
		logger.Fatalf("err Connecting server: %s", err)
	}

	if err = db.Ping(context.Background(), nil); err != nil {
		logger.Fatalf("err connecting to server: %s", err)
	}

	logger.Infof("hackhub v%s on port: %d", VERSION, config.Server.Port)
	logger.Infof("Mongo connected to %s:%d", config.Database.Host, config.Database.Port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Server.Port), Routes(logger, &config, db)))
}
