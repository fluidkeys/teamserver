package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/fluidkeys/crypto/openpgp"
	"github.com/fluidkeys/teamserver/models"
)

// TeamsHandler is used to server up HTTP requests to `/teams`
type TeamsHandler struct {
}

func (h *TeamsHandler) ServeHTTP(res http.ResponseWriter, req *http.Request, db models.Datastore) {
	var uuid string
	uuid, req.URL.Path = shiftPath(req.URL.Path)
	fmt.Printf("uuid: %s\n", uuid)
	fmt.Printf("req.URL.Path: %s\n", req.URL.Path)
	if req.URL.Path == "/" && uuid == "" {
		switch req.Method {
		case "GET":
			h.handleIndexGet(db).ServeHTTP(res, req)
		case "POST":
			h.handleIndexPost(db).ServeHTTP(res, req)
		default:
			http.Error(res, "Only GET or POST are allowed", http.StatusMethodNotAllowed)
		}
	}
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

func (h *TeamsHandler) handleGet(uuid string) {
	fmt.Printf("UUID: %v\n", uuid)
}
