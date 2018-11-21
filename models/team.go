package models

import (
	"github.com/satori/go.uuid"
)

// A Team represents a Fluidkeys team that use the server
type Team struct {
	ID      string    `json:"id,omitempty"`
	Name    string    `json:"teamName,omitempty"`
	UUID    string    `json:"uuid,omitempty"`
	Members []*Member `json:"members,omitempty"`
}

// A TeamUUID represents a simple json structure used in response
type TeamUUID struct {
	UUID string `json:"teamUuid,omitempty"`
}

// A TeamsPOST represents a simple json structure when creating a team
type TeamsPOST struct {
	Name      string `json:"teamName,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
}

// A RequestPOST represents a simple json structure requesting to join a team
// posted to the teams UUID, containing the public key
type RequestPOST struct {
	PublicKey string `json:"publicKey,omitempty"`
}

// A TeamSummary represents a simplified team output
type TeamSummary struct {
	*Team
	ID   omit `json:"id,omitempty"`
	UUID omit `json:"uuid,omitempty"`
}

type omit *struct{}

// AllTeams reads all the teams in the database
func (db *DB) AllTeams() ([]*Team, error) {
	teams := make([]*Team, 0)
	rows, err := db.Query(`SELECT id, name, uuid FROM teams`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		team := Team{}
		err = rows.Scan(&team.ID, &team.Name, &team.UUID)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return teams, nil
}

// CreateTeam inserts a record for the given teamName in the database returning
// the ID of the record
func (db *DB) CreateTeam(teamName string) (int64, *uuid.UUID, error) {
	uuid := uuid.NewV4()
	sqlStatement := `INSERT INTO teams (name, uuid) VALUES ($1, $2) RETURNING id`
	writeDB, err := db.Begin()
	if err != nil {
		writeDB.Rollback()
		return 0, nil, err
	}
	var teamID int64
	err = writeDB.QueryRow(sqlStatement, teamName, uuid).Scan(&teamID)
	if err != nil {
		writeDB.Rollback()
		return 0, nil, err
	}
	return teamID, &uuid, writeDB.Commit()
}

// CreateTeamUser inserts a record for the given user in the database, returning
// the ID.
func (db *DB) CreateTeamUser(teamID int64, fingerprint string) (int64, error) {
	sqlStatement := `INSERT INTO team_users (team_id, fingerprint, is_admin) VALUES ($1, $2, $3) RETURNING id`
	writeDB, err := db.Begin()
	if err != nil {
		writeDB.Rollback()
		return 0, err
	}
	var teamUserID int64
	err = writeDB.QueryRow(sqlStatement, teamID, fingerprint, true).Scan(&teamUserID)
	if err != nil {
		writeDB.Rollback()
		return 0, err
	}
	return teamUserID, writeDB.Commit()
}

// CreatePublicKey takes a fingerprint and publickey and creates a record in the
// database, returning the ID.
func (db *DB) CreatePublicKey(fingerprint string, publicKey string) (int64, error) {
	sqlStatement := `INSERT INTO public_keys (fingerprint, armoredPublicKey)
		VALUES ($1, $2) ON CONFLICT ON CONSTRAINT public_keys_pkey
		DO UPDATE SET fingerprint = $1 RETURNING id`
	// TODO: To ensure we get the return id, I've added the 'ON CONFLICT' clause
	// I don't really think this is the best approach, but for now it works.
	writeDB, err := db.Begin()
	if err != nil {
		writeDB.Rollback()
		return 0, err
	}
	var publicKeyID int64
	err = writeDB.QueryRow(sqlStatement, fingerprint, publicKey).Scan(&publicKeyID)
	if err != nil {
		writeDB.Rollback()
		return 0, err
	}
	return publicKeyID, writeDB.Commit()
}

// GetTeam uses a uuid to retrieve a single team from the database.
func (db *DB) GetTeam(uuid uuid.UUID) (*Team, error) {
	sqlStatement := `SELECT id, name, uuid FROM teams WHERE uuid=$1`
	rows, err := db.Query(sqlStatement, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	team := Team{}
	for rows.Next() {
		err = rows.Scan(&team.ID, &team.Name, &team.UUID)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// CreateTeamJoinRequest creates a record team_join_requests record in the
// database, finding the team id using the passed UUID.
func (db *DB) CreateTeamJoinRequest(fingerprint string, uuid string) (int64, error) {
	sqlStatement := `INSERT INTO team_join_requests (team_id, fingerprint)
		SELECT t.id, $2 FROM teams t WHERE uuid=$1 RETURNING id`
	writeDB, err := db.Begin()
	if err != nil {
		writeDB.Rollback()
		return 0, err
	}
	var teamJoinRequestID int64
	err = writeDB.QueryRow(sqlStatement, uuid, fingerprint).Scan(&teamJoinRequestID)
	if err != nil {
		writeDB.Rollback()
		return 0, err
	}
	return teamJoinRequestID, writeDB.Commit()
}
