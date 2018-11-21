package models

// A Member represents a Fluidkeys user on the teamserver
type Member struct {
	PublicKey string `json:"publicKey,omitempty"`
	IsAdmin   bool   `json:"isAdmin,omitempty"`
}

// GetTeamMembers returns all users for a particular team id
func (db *DB) GetTeamMembers(teamID int) ([]*Member, error) {
	members := make([]*Member, 0)
	rows, err := db.Query(`SELECT pk.armoredpublickey, tu.is_admin FROM
		public_keys pk, team_users tu
		WHERE team_id=$1 AND pk.fingerprint=tu.fingerprint`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		member := Member{}
		err = rows.Scan(&member.PublicKey, &member.IsAdmin)
		if err != nil {
			return nil, err
		}
		members = append(members, &member)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return members, nil
}
