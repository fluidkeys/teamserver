package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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
		var team models.Team
		err := decoder.Decode(&team)
		if err != nil {
			panic(err)
		}
		env.db.CreateTeam(team.Name)
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
	herokuDatabaseUrl, present := os.LookupEnv("DATABASE_URL")
	if present {
		return herokuDatabaseUrl
	} else {
		return fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}
}
