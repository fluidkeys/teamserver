package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fluidkeys/crypto/openpgp"
	"github.com/fluidkeys/teamserver/models"

	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "teamserver"
	dbname = "teamserver_development"
)

var (
	password = os.Getenv("TEAMSERVER_PASSWORD")
)

// Env provides a way to hook into the database
type Env struct {
	db models.Datastore
}

func main() {
	db, err := models.NewDB(connStr())
	if err != nil {
		log.Panic(err)
	}

	env := &Env{db}

	http.HandleFunc("/teams", env.teamsIndex)
	err = http.ListenAndServe(Port(), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (env *Env) teamsIndex(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		teams, err := env.db.AllTeams()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		out, err := json.Marshal(teams)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(w, string(out))
	case http.MethodPost:
		decoder := json.NewDecoder(r.Body)
		var teamPost models.TeamsPOST
		err := decoder.Decode(&teamPost)
		if err != nil {
			panic(err)
		}

		entityList, err := openpgp.ReadArmoredKeyRing(strings.NewReader(teamPost.PublicKey))
		if err != nil {
			err := fmt.Sprintf("error reading armored key ring: %v", err)
			http.Error(w, formatAsJSONMessage(err), http.StatusInternalServerError)
			return
		}
		if len(entityList) != 1 {
			err := fmt.Sprintf("expected 1 openpgp.Entity, got %d!", len(entityList))
			http.Error(w, formatAsJSONMessage(err), http.StatusInternalServerError)
			return
		}
		entity := entityList[0]

		fingerprint := fingerprintString(entity.PrimaryKey.Fingerprint)

		_, err = env.db.CreatePublicKey(fingerprint, teamPost.PublicKey)
		if err != nil {
			http.Error(w, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}

		teamID, teamUUID, err := env.db.CreateTeam(teamPost.Name)
		if err != nil {
			http.Error(w, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}

		_, err = env.db.CreateTeamUser(teamID, fingerprint)
		if err != nil {
			http.Error(w, formatAsJSONMessage(err.Error()), http.StatusInternalServerError)
			return
		}

		out, err := json.Marshal(models.TeamUUID{UUID: teamUUID.String()})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		fmt.Fprintf(w, string(out))
	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Port retrieves the port from the environment so we can run on Heroku
func Port() string {
	var port = os.Getenv("PORT")
	// Set a default port if there is nothing in the environment
	if port == "" {
		port = "4747"
		fmt.Println("INFO: No PORT environment variable detected, defaulting to " + port)
	}
	return ":" + port
}

func connStr() string {
	herokuDatabaseURL, present := os.LookupEnv("DATABASE_URL")
	if present {
		return herokuDatabaseURL
	}
	return fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

func fingerprintString(b [20]byte) string {
	return fmt.Sprintf(
		"%0X %0X %0X %0X %0X  %0X %0X %0X %0X %0X",
		b[0:2], b[2:4], b[4:6], b[6:8], b[8:10],
		b[10:12], b[12:14], b[14:16], b[16:18], b[18:20],
	)
}

func formatAsJSONMessage(message string) string {
	bytes, _ := json.Marshal(map[string]string{"message": message})
	return string(bytes)
}
