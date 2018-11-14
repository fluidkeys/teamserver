package models

import "github.com/satori/go.uuid"

// A Team represents a Fluidkeys team that use the server
type Team struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"teamName,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

// AllTeams reads all the teams in the database
func (db *DB) AllTeams() ([]*Team, error) {
	teams := make([]*Team, 0)
	rows, err := db.Query(`
		SELECT id, name, uuid FROM teams`)
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

// CreateTeam inserts a record for the given teamName in the database
func (db *DB) CreateTeam(teamName string) (int64, error) {
	uuid := uuid.NewV4()
	sqlStatement := `INSERT INTO teams (name, uuid) VALUES ($1, $2) RETURNING id`
	var id int64
	err := db.QueryRow(sqlStatement, teamName, uuid).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
