package main

import (
	"github.com/johnamadeo/server"
)

type Organization struct {
	Name            string
	Admin           string
	CrossMatchTrait string
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

func getCrossMatchTrait(orgname string) (string, error) {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return "", err
	}

	rows, err := db.Query(
		"SELECT cross_match_trait FROM organizations WHERE name = $1",
		orgname,
	)
	if err != nil {
		return "", err
	}

	var crossMatchTrait string
	for rows.Next() {
		err := rows.Scan(&crossMatchTrait)
		if err != nil {
			return "", err
		}
		break
	}

	return crossMatchTrait, nil

}

func setCrossMatchTrait(orgname string, crossMatchTrait string) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE organizations SET cross_match_trait = $1 WHERE name = $2",
		crossMatchTrait,
		orgname,
	)
	if err != nil {
		return err
	}

	return nil
}
