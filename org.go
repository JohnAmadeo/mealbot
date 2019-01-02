package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/johnamadeo/server"
)

type Organization struct {
	Name            string
	Admin           string
	CrossMatchTrait string
}

type CreateOrganizationRequestBody struct {
	Organization string `json:"org"`
}

type SetCrossMatchTraitRequestBody struct {
	Trait string `json:"trait"`
}

func GetOrganizationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "" {
		fmt.Println(r.Method + " Only GET requests are allowed at this route")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only GET requests are allowed at this route"))
		return
	}

	queries, ok := r.URL.Query()["admin"]
	if !ok || len(queries) > 1 {
		fmt.Println("Request query parameters must contain admin")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.StrToBytes("Request query parameters must contain admin"))
		return
	}

	organizations, err := getOrganizations(queries[0])
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	resp := map[string][]string{"orgs": organizations}
	bytes, err := json.Marshal(resp)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only POST requests are allowed at this route"))
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.StrToBytes("Malformed body."))
		return
	}
	defer r.Body.Close()

	var body CreateOrganizationRequestBody
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.StrToBytes("Request body is malformed"))
		return
	}

	queries, ok := r.URL.Query()["admin"]
	if !ok || len(queries) > 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.StrToBytes("Request query parameters must contain admin"))
		return
	}
	admin := queries[0]

	fmt.Println(body.Organization, admin)

	err = createOrganization(body.Organization, admin)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(server.StrToBytes("Successfully created new organization"))
}

func CrossMatchTraitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only POST requests are allowed at this route"))
		return
	}

	orgname, err := getQueryParam(r, "org")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.StrToBytes("Malformed body."))
		return
	}
	defer r.Body.Close()

	var body SetCrossMatchTraitRequestBody
	err = json.Unmarshal(bytes, &body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.StrToBytes("Request body is malformed"))
		return
	}

	err = setCrossMatchTrait(orgname, body.Trait)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(server.StrToBytes("Successfully set the cross match trait"))
}

func getOrganizations(admin string) ([]string, error) {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return []string{}, err
	}

	rows, err := db.Query(
		"SELECT name FROM organizations WHERE admin = $1",
		admin,
	)
	if err != nil {
		return []string{}, err
	}

	organizations := []string{}
	for rows.Next() {
		var organization string
		err := rows.Scan(&organization)
		if err != nil {
			return []string{}, err
		}

		organizations = append(organizations, organization)
	}

	return organizations, nil
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

	var crossMatchTraitSql sql.NullString
	for rows.Next() {
		err := rows.Scan(&crossMatchTraitSql)
		if err != nil {
			return "", err
		}
		break
	}

	crossMatchTrait := ""
	if crossMatchTraitSql.Valid {
		crossMatchTrait = crossMatchTraitSql.String
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
