package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/johnamadeo/server"
)

// Organization :
type Organization struct {
	Name            string
	Admin           string
	CrossMatchTrait string
}

// CreateOrganizationRequestBody :
type CreateOrganizationRequestBody struct {
	Organization string `json:"org"`
}

// SetCrossMatchTraitRequestBody :
type SetCrossMatchTraitRequestBody struct {
	Trait string `json:"trait"`
}

// GetOrganizationsHandler : HTTP Handler for fetching all the organizations an admin manages
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

// CreateOrganizationHandler : HTTP handler for creating a new organization
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

// CrossMatchTraitHandler : HTTP handler for changing or setting a cross match trait for an organization
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

// GetOrganizations :
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
	if name == "" {
		return errors.New("Organization name cannot be an empty string")
	}

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

// GetCrossMatchTrait : Placeholder
func GetCrossMatchTrait(orgname string) (string, error) {
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

	var crossMatchTraitSQL sql.NullString
	for rows.Next() {
		err := rows.Scan(&crossMatchTraitSQL)
		if err != nil {
			return "", err
		}
		break
	}

	crossMatchTrait := ""
	if crossMatchTraitSQL.Valid {
		crossMatchTrait = crossMatchTraitSQL.String
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
