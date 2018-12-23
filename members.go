package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/johnamadeo/server"
)

const (
	MaxMemory            = 32 << 20
	CSVPath              = "src/github.com/johnamadeo/mealbot/csv/"
	FileFlag             = os.O_WRONLY | os.O_CREATE
	ReadWritePermissions = 0666
)

type Member struct {
	Organization string
	Id           string
	Name         string
	Metadata     map[string]string
	PairCounts   map[string]int
}

func MembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only POST requests are allowed at this route"))
		return
	}

	err := r.ParseMultipartForm(MaxMemory)
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

	filename := filepath.Join(build.Default.GOPATH, CSVPath, handler.Filename)

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

	err = createMembersFromCSV(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(server.StrToBytes("Successfully saved members from CSV"))
}

func isValidFormatCSV(headers []string) bool {
	for _, val := range headers {
		if strings.ToLower(val) == "name" {
			return true
		}
	}

	return false
}

func createMembersFromCSV(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	reader := csv.NewReader(bufio.NewReader(file))
	members := []Member{}

	headers := []string{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if len(headers) == 0 {
			if isValidFormatCSV(row) {
				for _, val := range row {
					headers = append(headers, strings.ToLower(val))
				}
				continue
			} else {
				return errors.New("CSV must have a column titled 'name'")
			}
		}

		name := ""
		metadata := map[string]string{}
		for i, val := range row {
			if headers[i] == "name" {
				name = val
			} else {
				metadata[headers[i]] = val
			}
		}

		members = append(members, Member{
			Organization: "test", // Need to grab from HTTP request
			Id:           uuid.New().String(),
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
			members[i].PairCounts[members[j].Id] = 0
		}
	}

	err = saveMembersInDB(members)
	if err != nil {
		return err
	}

	return nil
}

func saveMembersInDB(members []Member) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	placeholders := []string{}
	values := []interface{}{}
	for i, member := range members {
		placeholder := fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d)",
			i*5+1, i*5+2, i*5+3, i*5+4, i*5+5,
		)
		placeholders = append(placeholders, placeholder)

		values = append(values, member.Organization)
		values = append(values, member.Id)
		values = append(values, member.Name)

		bytes, err := json.Marshal(member.Metadata)
		if err != nil {
			return err
		}
		values = append(values, server.JSONB(bytes))

		bytes, err = json.Marshal(member.PairCounts)
		if err != nil {
			return err
		}
		values = append(values, server.JSONB(bytes))
	}

	query := fmt.Sprintf(
		"INSERT INTO members (organization, id, name, metadata, pair_counts) VALUES %s",
		strings.Join(placeholders, ", "),
	)

	_, err = db.Exec(query, values...)
	if err != nil {
		return err
	}

	return nil
}

