package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

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
	db           models.Datastore
	TeamsHandler *TeamsHandler
}

func (env *Env) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	var head string
	head, req.URL.Path = shiftPath(req.URL.Path)
	if head == "teams" {
		env.TeamsHandler.ServeHTTP(res, req, env.db)
		return
	}
	http.Error(res, "Not Found", http.StatusNotFound)
}

func main() {
	db, err := models.NewDB(connStr())
	if err != nil {
		log.Panic(err)
	}
	env := &Env{db, new(TeamsHandler)}

	err = http.ListenAndServe(Port(), env)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
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

// ShiftPath splits off the first component of p, which will be cleaned of
// relative components before processing. head will never contain a slash and
// tail will always be a rooted path without trailing slash.
func shiftPath(p string) (head, tail string) {
	p = path.Clean("/" + p)
	i := strings.Index(p[1:], "/") + 1
	if i <= 0 {
		return p[1:], "/"
	}
	return p[1:i], p[i:]
}
