package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fluidkeys/crypto/openpgp"
	"github.com/fluidkeys/teamserver/models"
	uuid "github.com/satori/go.uuid"
)

// TeamsHandler is used to server up HTTP requests to `/teams`
type TeamsHandler struct {
	SummaryHandler *SummaryHandler
}

func (h *TeamsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request, db models.Datastore) {
	var uuid string
	uuid, tail := shiftPath(req.URL.Path)
	if uuid == "" {
		switch req.Method {
		case "GET":
			h.handleIndexGet(db).ServeHTTP(res, req)
		case "POST":
			h.handleIndexPost(db).ServeHTTP(res, req)
		default:
			http.Error(res, "Only GET and POST are allowed", http.StatusMethodNotAllowed)
		}
	} else {
		switch tail {
		case "/":
			h.handleGet(uuid, db).ServeHTTP(res, req)
		case "/summary":
			h.SummaryHandler.Handler(uuid, db).ServeHTTP(res, req)
		default:
			http.Error(res, "Not Found", http.StatusNotFound)
		}
	} else {
		switch req.Method {
		case "GET":
			h.handleGet(uuid, db).ServeHTTP(res, req)
		default:
			http.Error(res, "Only GET is allowed", http.StatusMethodNotAllowed)
		}
	}
	return
}

func (h *TeamsHandler) handleIndexGet(db models.Datastore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		teams, err := db.AllTeams()
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		out, err := json.Marshal(teams)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(res, string(out))
	})
}

func (h *TeamsHandler) handleIndexPost(db models.Datastore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		var teamPost models.TeamsPOST
		err := decoder.Decode(&teamPost)
		if err != nil {
			panic(err)
		}

		entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(teamPost.PublicKey))
		if err != nil {
			err := fmt.Sprintf("error reading armored key ring: %v", err)
			http.Error(res, formatAsJSONMessage(err), http.StatusInternalServerError)
			return
		}
		if len(entityList) != 1 {
			err := fmt.Sprintf("expected 1 openpgp.Entity, got %d!", len(entityList))
			http.Error(res, formatAsJSONMessage(err), http.StatusInternalServerError)
			return
		}
		entity := entityList[0]

		fingerprint := fingerprintString(entity.PrimaryKey.Fingerprint)

		_, err = db.CreatePublicKey(fingerprint, teamPost.PublicKey)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}

		teamID, teamUUID, err := db.CreateTeam(teamPost.Name)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}

		_, err = db.CreateTeamUser(teamID, fingerprint)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}

		out, err := json.Marshal(models.TeamUUID{UUID: teamUUID.String()})
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(res, string(out))
	})
}

func (h *TeamsHandler) handleGet(uuidString string, db models.Datastore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		uuid, err := uuid.FromString(uuidString)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		team, err := db.GetTeam(uuid)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		out, err := json.Marshal(team)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(res, string(out))
	})
}

// SummaryHandler is used to server up HTTP requests to `/teams/{uuid}/summary`
type SummaryHandler struct{}

// Handler takes a team UUID and database and then looks up the record in the
// database, writing JSON back.
func (h *SummaryHandler) Handler(uuidString string, db models.Datastore) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		uuid, err := uuid.FromString(uuidString)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		team, err := db.GetTeam(uuid)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		out, err := json.Marshal(models.TeamSummary{
			Team: team,
		})
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(res, string(out))
	})
}
