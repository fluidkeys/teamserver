package main

import (
	"net/http"

	"github.com/fluidkeys/teamserver/models"
)

// RequestHandler is used to receive requests to join teams
type RequestHandler struct{}

// Handler takes a team UUID and database and then creates the appropriate
// record in the database.
func (h *RequestHandler) Handler(uuidString string, db models.Datastore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// TODO: Process the request
	})
}
