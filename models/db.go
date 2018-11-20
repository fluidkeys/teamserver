package models

import (
	"database/sql"

	uuid "github.com/satori/go.uuid"
)

// Datastore is an interface specifiying all the ways of interacting with the
// database
type Datastore interface {
	AllTeams() ([]*Team, error)
	CreateTeam(string) (int64, *uuid.UUID, error)
	CreateTeamUser(int64, string) (int64, error)
	CreatePublicKey(string, string) (int64, error)
	GetTeam(uuid.UUID) (*Team, error)
}

// DB is a struct the points at a sql database
type DB struct {
	*sql.DB
}

// NewDB populates the global db variable with an opened postgres database
func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
