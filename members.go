package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/johnamadeo/server"
)

const (
	MaxMemory            = 32 << 20
	CSVPath              = "./csv/"
	FileFlag             = os.O_WRONLY | os.O_CREATE
	ReadWritePermissions = 0666
)

type Member struct {
	Organization string
	Email        string
	Name         string
	Metadata     map[string]string
	PairCounts   map[string]int
}

type MemberResponse map[string]string
type CreateMembersResponse struct {
	Members []MemberResponse `json:"members"`
	Traits  []string         `json:"traits"`
}

type GetMembersResponse struct {
	Members         []MemberResponse `json:"members"`
	Traits          []string         `json:"traits"`
	CrossMatchTrait string           `json:"crossMatchTrait"`
}

func MembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		CreateMembersHandler(w, r)
	} else if r.Method == "GET" {
		GetMembersHandler(w, r)
	}
}

func GetMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only GET requests are allowed at this route"))
		return
	}

	orgname, err := getQueryParam(r, "org")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	members, err := getMembersFromDB(orgname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	traits := []string{}
	if len(members) > 0 {
		for trait, _ := range members[0] {
			if trait != "name" && trait != "email" {
				traits = append(traits, trait)
			}
		}
	}

	crossMatchTrait, err := getCrossMatchTrait(orgname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	resp := GetMembersResponse{
		Members: members,
		Traits:  traits,
	}
	if crossMatchTrait != "" {
		resp.CrossMatchTrait = crossMatchTrait
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func CreateMembersHandler(w http.ResponseWriter, r *http.Request) {
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

	err = r.ParseMultipartForm(MaxMemory)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	formFile, handler, err := r.FormFile("members")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	defer formFile.Close()

	filename := filepath.Join(CSVPath, handler.Filename)

	file, err := os.OpenFile(filename, FileFlag, ReadWritePermissions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}
	defer file.Close()

	_, err = io.Copy(file, formFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	members, err := createMembersFromCSV(orgname, filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	var traits []string
	for trait, _ := range members[0] {
		if trait != "name" && trait != "email" {
			traits = append(traits, trait)
		}
	}

	resp := CreateMembersResponse{
		Members: members,
		Traits:  traits,
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(bytes)
}

func isValidFormatCSV(headers []string) bool {
	for _, val := range headers {
		if strings.ToLower(val) == "name" {
			return true
		}
	}

	return false
}

func createMembersFromCSV(orgname string, filename string) ([]MemberResponse, error) {
	file, err := os.Open(filename)
	if err != nil {
		return []MemberResponse{}, err
	}

	reader := csv.NewReader(bufio.NewReader(file))
	members := []Member{}

	headers := []string{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return []MemberResponse{}, err
		}

		if len(headers) == 0 {
			if isValidFormatCSV(row) {
				for _, val := range row {
					headers = append(headers, strings.ToLower(val))
				}
				continue
			} else {
				err = errors.New("CSV must have a column titled 'name'")
				return []MemberResponse{}, err
			}
		}

		name := ""
		email := ""
		metadata := map[string]string{}
		for i, val := range row {
			if headers[i] == "name" {
				name = val
			} else if headers[i] == "email" {
				email = val
			} else {
				metadata[headers[i]] = val
			}
		}

		members = append(members, Member{
			Organization: orgname, // Need to grab from HTTP request
			Email:        email,
			Name:         name,
			Metadata:     metadata,
			PairCounts:   map[string]int{}, // fill in later
		})
	}

	for i, _ := range members {
		for j, _ := range members {
			if i == j {
				continue
			}
			members[i].PairCounts[members[j].Email] = 0
		}
	}

	err = saveMembersInDB(members)
	if err != nil {
		return []MemberResponse{}, err
	}

	var membersJson []MemberResponse
	for _, member := range members {
		memberJson := member.Metadata
		memberJson["name"] = member.Name
		memberJson["email"] = member.Email
		membersJson = append(membersJson, memberJson)
	}

	return membersJson, nil
}

// TODO: Batch insert
func saveMembersInDB(members []Member) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	for _, member := range members {
		metadataBytes, err := json.Marshal(member.Metadata)
		if err != nil {
			return err
		}

		pairCountsBytes, err := json.Marshal(member.PairCounts)
		if err != nil {
			return err
		}

		columns := "(organization, email, name, metadata, pair_counts)"
		placeholders := "($1, $2, $3, $4, $5)"
		_, err = db.Exec(
			fmt.Sprintf(
				"INSERT INTO members %s VALUES %s",
				columns,
				placeholders,
			),
			member.Organization,
			member.Email,
			member.Name,
			server.JSONB(metadataBytes),
			server.JSONB(pairCountsBytes),
		)

		if err != nil && !strings.Contains(err.Error(), DuplicateKeyErr) {
			return err
		}
	}

	return nil
}

func getMembersFromDB(orgname string) ([]MemberResponse, error) {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return []MemberResponse{}, err
	}

	members := []MemberResponse{}

	rows, err := db.Query(
		"SELECT name, email, metadata FROM members WHERE organization = $1 ORDER BY name",
		orgname,
	)
	if err != nil {
		return members, err
	}

	for rows.Next() {
		var name, email string
		var metadataJson server.JSONB
		err := rows.Scan(&name, &email, &metadataJson)
		if err != nil {
			return members, err
		}

		bytes, err := metadataJson.MarshalJSON()
		if err != nil {
			return members, err
		}

		var member map[string]string
		err = json.Unmarshal(bytes, &member)
		if err != nil {
			return members, err
		}

		member["name"] = name
		member["email"] = email

		members = append(members, member)
	}

	return members, nil
}

// Helper function used in getPairsFromDB
func getMemberFromDB(orgname string, email string, db *sql.DB) (Member, error) {
	memberRows, err := db.Query(
		"SELECT name FROM members WHERE organization = $1 AND email = $2",
		orgname,
		email,
	)
	if err != nil {
		return Member{}, err
	}

	var member Member
	for memberRows.Next() {
		var name string
		err := memberRows.Scan(&name)
		if err != nil {
			return Member{}, err
		}

		member = Member{
			Name:  name,
			Email: email,
		}
		break
	}

	return member, nil
}

func readCSV_TEST(fileName string) [][]string {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return [][]string{}
	}

	rawStudents := [][]string{
		[]string{},
		[]string{},
		[]string{},
		[]string{},
	}

	reader := csv.NewReader(bufio.NewReader(f))
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		year, err := strconv.Atoi(line[3])
		if err != nil {
			continue
		}
		rawStudents[3-(year-2019)] = append(
			rawStudents[3-(year-2019)],
			fmt.Sprintf("%s %d", line[0], year),
		)
	}

	return rawStudents
}
