package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fluidkeys/teamserver/models"
	uuid "github.com/satori/go.uuid"
)

// SummaryHandler is used to server up HTTP requests to `/teams/{uuid}/summary`
type SummaryHandler struct{}

// Handler takes a team UUID and database and then looks up the record in the
// database, writing JSON back.
func (h *SummaryHandler) Handler(uuidString string, db models.Datastore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		uuid, err := uuid.FromString(uuidString)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}
		team, err := db.GetTeam(uuid)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}
		out, err := json.Marshal(models.TeamSummary{
			Team: team,
		})
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(res, string(out))
	})
}
