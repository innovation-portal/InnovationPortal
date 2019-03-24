// Package projects has the actions for the Project resource
package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/render"

	"github.com/go-chi/jwtauth"

	"github.com/hackathon/hackhub/pkg/config"

	log "github.com/sirupsen/logrus"

	"github.com/go-chi/chi"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"

	uuid "github.com/satori/go.uuid"
)

// Project represents one location in the collection
type Project struct {
	ID              string   `bson:"_id,omitempty"`
	Name            string   `bson:"name"`
	Description     string   `bson:"description"`
	Tags            []string `bson:"tags"`
	TagLine         string   `bson:"tag_line"`
	Members         []string `bson:"members"`
	Photo           string   `bson:"photo"`
	ApplicationArea []string `bson:"application_area"`
	Winner          bool     `bson:"winner"`
	WinnerType      string   `bson:"winner_type"`
	Hackathon       string   `bson:"hackathon"`
	Year            int      `bson:"year"`
}

// Handler contains the chi.Mux, logrus.Logger,
// database.Collection for users and locations, config,
// maps.Client and jwt auth
type Handler struct {
	Router     *chi.Mux
	Collection *mongo.Collection
	Logger     *log.Logger
	Config     *config.Configuration
	JWT        *jwtauth.JWTAuth
}

// Routes creates routes for the location module
func Routes(logger *log.Logger, config *config.Configuration, tokenAuth *jwtauth.JWTAuth) *Handler {
	router := chi.NewRouter()

	// Connect to database
	db, err := mongo.Connect(
		context.TODO(),
		fmt.Sprintf("mongodb://%s:%d", config.Database.Host, config.Database.Port),
	)
	if err != nil {
		logger.Fatalf("err Connecting database: %s", err)
	}

	// Check the connection
	if err = db.Ping(context.TODO(), nil); err != nil {
		logger.Fatalf("err pinging database: %s", err)
	}

	collection := db.Database("hackhub").Collection("projects")
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	// indexes
	IDIndex := yieldIndexModel("_id", true)
	collection.Indexes().CreateOne(context.Background(), IDIndex, opts)

	NameIndex := yieldIndexModel("name", true)
	collection.Indexes().CreateOne(context.Background(), NameIndex, opts)

	handler := &Handler{router, collection, logger, config, tokenAuth}

	// This is authentication implemented later
	// router.Use(
	// 	jwtauth.Verifier(tokenAuth),
	// 	jwtauth.Authenticator,
	// )

	// Routes for the location namespace
	router.Get("/", handler.GetProjects)
	router.Post("/", handler.PostProject)
	router.Get("/{ProjectID}", handler.GetProject)
	router.Put("/{ProjectID}", handler.UpdateProject)
	router.Delete("/{ProjectID}", handler.DeleteProject)

	return handler
}

// PostProject will post the users location along with metrics from the user
func (h *Handler) PostProject(w http.ResponseWriter, r *http.Request) {
	var project Project
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&project); err != nil {
		h.Logger.Errorf("err decoding user request: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	project.ID = uuid.NewV4().String()
	if _, err := h.Collection.InsertOne(context.Background(), project); err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			h.Logger.Warnf("Project %v Already exists: %s", project, err)
			http.Error(w, "Project with that name already exists", http.StatusConflict)
			return
		}
		h.Logger.Errorf("err writing to db request: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	render.JSON(w, r, project)
}

// UpdateProject will update an existing project
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	var project Project
	projectID := chi.URLParam(r, "ProjectID")

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&project); err != nil {
		h.Logger.Errorf("err decoding user request: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	opts := options.Update()
	filter := bson.M{"_id": projectID}
	result, err := h.Collection.UpdateOne(context.Background(), filter, project, opts)
	if err != nil {
		h.Logger.Errorf("err decoding user request: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if result.MatchedCount > 0 {
		render.JSON(w, r, map[string]string{"message": "User updated"})
	}
}

// GetProject retrieves one location from PlaceID
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	var project Project
	projectID := chi.URLParam(r, "ProjectID")
	filter := bson.M{"_id": projectID}
	result := h.Collection.FindOne(context.TODO(), filter)
	if err := result.Decode(&project); err != nil {
		if strings.Contains(err.Error(), "mongo: no documents in result") {
			http.Error(w, "Project doesn't exist", 404)
			return
		}
		h.Logger.Errorf("err decoding: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	render.JSON(w, r, project) // A chi router helper for serializing and returning json
}

// GetProjects retrieves a list of projects
func (h *Handler) GetProjects(w http.ResponseWriter, r *http.Request) {
	opts := options.Find()
	var projects []Project
	cur, err := h.Collection.Find(context.TODO(), bson.D{{}}, opts)
	if err != nil {
		h.Logger.Errorf("err retrieving cursor item: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	for cur.Next(context.TODO()) {
		project := &Project{}
		err := cur.Decode(&project)
		if err != nil {
			h.Logger.Errorf("err decoding item: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		projects = append(projects, *project)
	}
	render.JSON(w, r, projects)
}

// DeleteProject retrieves a list of projects
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	opts := options.Delete()
	projectID := chi.URLParam(r, "ProjectID")
	filter := bson.M{"_id": projectID}
	result, err := h.Collection.DeleteOne(context.TODO(), filter, opts)
	if err != nil {
		if strings.Contains(err.Error(), "mongo: no documents in result") {
			http.Error(w, "Project doesn't exist", 404)
			return
		}
		h.Logger.Errorf("err decoding: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Project doesn't exist", 404)
		return
	}
	render.JSON(w, r, map[string]string{"message": fmt.Sprintf("project %s deleted", projectID)}) // A chi router helper for serializing and returning json
}

func yieldIndexModel(key string, unique bool) mongo.IndexModel {
	keys := bsonx.Doc{{Key: key, Value: bsonx.Int32(int32(1))}}
	index := mongo.IndexModel{}
	index.Keys = keys
	index.Options = &options.IndexOptions{Unique: &unique}
	return index
}
