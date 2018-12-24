package main

import (
	"github.com/johnamadeo/server"
)

type Organization struct {
	Name               string
	Admin              string
	CrossMatchCriteria string
}

func createOrganization(name string, admin string) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO organizations (name, admin) VALUES ($1, $2)",
		name,
		admin,
	)
	if err != nil {
		return err
	}

	return nil
}

func getCrossMatchCriteria(orgname string) (string, error) {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return "", err
	}

	rows, err := db.Query(
		"SELECT cross_match_criteria FROM organizations WHERE name = $1",
		orgname,
	)
	if err != nil {
		return "", err
	}

	var crossMatchCriteria string
	for rows.Next() {
		err := rows.Scan(&crossMatchCriteria)
		if err != nil {
			return "", err
		}
		break
	}

	return crossMatchCriteria, nil

}

func setCrossMatchCriteria(orgname string, crossMatchCriteria string) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE organizations SET cross_match_criteria = $1 WHERE name = $2",
		crossMatchCriteria,
		orgname,
	)
	if err != nil {
		return err
	}

	return nil
}
