package main

import (
	"github.com/johnamadeo/server"
)

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
