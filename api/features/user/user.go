// Package user handles user data
package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/jwtauth"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
	"github.com/hackathon/hackhub/pkg/config"
)

// User is the user model for the application
type User struct {
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
	Email     string `json:"email" bson:"_id,omitempty"`
	Location  string `json:"location" bson:"location"`
	Password  string `json:"password" bson:"password"`
}

// Handler contains the router and DB client
type Handler struct {
	Router     *chi.Mux
	Collection *mongo.Collection
	Logger     *log.Logger
	Config     *config.Configuration
	JWT        *jwtauth.JWTAuth
}

// Routes creates routes for the user module
func Routes(logger *log.Logger, config *config.Configuration, tokenAuth *jwtauth.JWTAuth) *Handler {
	router := chi.NewRouter()

	// Connect to database
	db, err := mongo.Connect(
		context.TODO(),
		fmt.Sprintf("mongodb://%s:%d", config.Database.Host, config.Database.Port),
	)
	if err != nil {
		logger.Fatalf("err Connecting server: %s", err)
	}

	// Check the connection
	if err = db.Ping(context.TODO(), nil); err != nil {
		logger.Fatalf("err pinging server: %s", err)
	}

	collection := db.Database("poppin").Collection("users")
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	index := yieldIndexModel()
	collection.Indexes().CreateOne(context.Background(), index, opts)

	handler := &Handler{router, collection, logger, config, tokenAuth}

	router.Group(func(router chi.Router) {
		router.Use(
			jwtauth.Verifier(tokenAuth),
			jwtauth.Authenticator,
		)

		// Routes protected by jwt
		router.Get("/", handler.GetUsers)
		router.Get("/{email}", handler.GetUser)
	})

	// Public route
	router.Post("/auth", handler.AuthUser)
	router.Post("/", handler.PostUser) // For now creating new user does not require token

	return handler
}

// GetUser returns one users
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")

	var user User
	res := h.Collection.FindOne(context.TODO(), bson.M{"_id": email}, options.FindOne())

	if err := res.Decode(&user); err != nil {
		if strings.Contains(err.Error(), "mongo: no documents in result") {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		h.Logger.Errorf("err decoding item: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	user.Password = ""      // Never return password hashes
	render.JSON(w, r, user) // A chi router helper for serializing and returning json
}

// GetUsers returns a list of users
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	cur, err := h.Collection.Find(context.TODO(), bson.D{{}}, options.Find())
	if err != nil {
		h.Logger.Errorf("err retrieving cursor item: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	for cur.Next(context.TODO()) {
		user := &User{}
		err := cur.Decode(&user)
		if err != nil {
			h.Logger.Errorf("err decoding item: %s", err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		user.Password = "" // Never return password hashes
		users = append(users, *user)
	}
	render.JSON(w, r, users) // A chi router helper for serializing and returning json
}

// PostUser will add a new user
func (h *Handler) PostUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var newUser User
	if err := decoder.Decode(&newUser); err != nil {
		if err.Error() == "EOF" {
			http.Error(w, "Empty user request", 400)
			return
		}
		h.Logger.Errorf("err decoding user request: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if newUser.Email == "" || newUser.Password == "" {
		http.Error(w, "Invalid email or password", 400)
		return
	}

	// Hash password
	bytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		h.Logger.Errorf("err hashing password: %s", err)
	}
	newUser.Password = string(bytes)

	_, err = h.Collection.InsertOne(context.TODO(), newUser)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			http.Error(w, "User already exists", 400)
			return
		}
		h.Logger.Errorf("error creating user: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	newUser.Password = ""

	render.JSON(w, r, newUser)
}

// AuthUser will check is user is autenticated
func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var user User
	if err := decoder.Decode(&user); err != nil {
		if err.Error() == "EOF" {
			http.Error(w, "Empty user request", 400)
			return
		}
		h.Logger.Errorf("err decoding user request: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if user.Email == "" || user.Password == "" {
		http.Error(w, "Invalid email or password", 400)
		return
	}

	filter := bson.M{"_id": user.Email}
	result := h.Collection.FindOne(context.TODO(), filter)

	var dbUser User
	if err := result.Decode(&dbUser); err != nil {
		if strings.Contains(err.Error(), "no document") {
			http.Error(w, "Invalid email or password", 400)
			return
		}
		h.Logger.Errorf("err decoding user request: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid email or password", 400)
		return
	}
	user.Password = ""

	// JWT
	_, token, err := h.JWT.Encode(jwt.MapClaims{
		"email": user.Email,
		"exp":   time.Now().UTC().Add(4 * time.Hour).Unix(),
	})
	if err != nil {
		h.Logger.Errorf("err creating JWT: %s", err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	render.JSON(w, r, map[string]string{"Token": token})
}

func yieldIndexModel() mongo.IndexModel {
	keys := bsonx.Doc{{Key: "Email", Value: bsonx.Int32(int32(1))}}
	index := mongo.IndexModel{}
	index.Keys = keys
	return index
}
