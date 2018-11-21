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
	RequestHandler *RequestHandler
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
		case "/request":
			h.RequestHandler.Handler(uuid, db).ServeHTTP(res, req)
		default:
			http.Error(res, "Not Found", http.StatusNotFound)
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

func getFingerprintFromPublicKey(armoredPublicKey string) (string, error) {
	entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(armoredPublicKey))
	if err != nil {
		return "", fmt.Errorf("error reading armored key ring: %v", err)
	}
	if len(entityList) != 1 {
		return "", fmt.Errorf("expected 1 openpgp.Entity, got %d", len(entityList))
	}
	entity := entityList[0]

	fingerprint := fingerprintString(entity.PrimaryKey.Fingerprint)
	return fingerprint, nil
}

func fingerprintString(b [20]byte) string {
	return fmt.Sprintf(
		"%0X %0X %0X %0X %0X  %0X %0X %0X %0X %0X",
		b[0:2], b[2:4], b[4:6], b[6:8], b[8:10],
		b[10:12], b[12:14], b[14:16], b[16:18], b[18:20],
	)
}

func (h *TeamsHandler) handleGet(uuidString string, db models.Datastore) http.Handler {
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
		out, err := json.Marshal(team)
		if err != nil {
			http.Error(res, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(res, string(out))
	})
}
