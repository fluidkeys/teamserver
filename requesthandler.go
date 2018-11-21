package main

import (
	"encoding/json"
	"net/http"

	"github.com/fluidkeys/teamserver/models"
)

// RequestHandler is used to receive requests to join teams
type RequestHandler struct{}

// Handler takes a team UUID and database and then creates the appropriate
// record in the database.
func (h *RequestHandler) Handler(uuidString string, db models.Datastore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var teamPost models.RequestPOST
		err := decoder.Decode(&teamPost)
		if err != nil {
			panic(err)
		}

		fingerprint, err := getFingerprintFromPublicKey(teamPost.PublicKey)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}

		_, err = db.CreatePublicKey(fingerprint, teamPost.PublicKey)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}
		_, err = db.CreateTeamJoinRequest(fingerprint, uuidString)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusCreated)
	})
}
