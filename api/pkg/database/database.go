package database

import (
	"context"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
)

// DB is the interface for interactions with the DB
type DB interface {
	Ping(context.Context, *readpref.ReadPref) error
}
