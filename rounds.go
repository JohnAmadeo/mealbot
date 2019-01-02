package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/johnamadeo/server"
)

const (
	// VERY rough check; need to improve
	NoMaxRoundErr = "converting driver.Value type <nil>"

	// Postgres timestamp string template patterns can be found in
	// https://www.postgresql.org/docs/8.1/functions-formatting.html
	TimestampFormat = "YYYY-MM-DD HH24:MI:ssZ"
)

type GetRoundsResponse struct {
	Rounds []string `json:"rounds"`
}

func AddRoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only POST requests are allowed at this route"))
		return
	}

	// TODO: Verify 'round' is a datestring in the YYYY-MM-DDTHH:mm:ss[Z] format
	values, err := getQueryParams(r, []string{"org", "round"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	orgname := values[0]
	roundDate := values[1]

	err = addRound(orgname, roundDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(server.StrToBytes("Successfully scheduled a new round"))
}

func GetRoundsHandler(w http.ResponseWriter, r *http.Request) {
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

	rounds, err := getRoundsFromDB(orgname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	resp := GetRoundsResponse{Rounds: rounds}
	bytes, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

func RoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		_, err := getQueryParam(r, "roundId")
		if err != nil {
			AddRoundHandler(w, r)
		} else {
			RescheduleRoundHandler(w, r)
		}
	} else if r.Method == "DELETE" {
		RemoveRoundHandler(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only POST and DELETE requests are allowed at this route"))
		return
	}
}

func getRoundsFromDB(orgname string) ([]string, error) {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return []string{}, err
	}

	rows, err := db.Query(
		"SELECT to_char(scheduled_date, $1) FROM rounds WHERE organization = $2 ORDER BY id ASC",
		TimestampFormat,
		orgname,
	)
	if err != nil {
		return []string{}, err
	}

	rounds := []string{}
	for rows.Next() {
		var roundDate string
		err = rows.Scan(&roundDate)
		if err != nil {
			return []string{}, err
		}

		rounds = append(rounds, roundDate)
	}

	return rounds, nil
}

func RemoveRoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only DELETE requests are allowed at this route"))
		return
	}

	values, err := getQueryParams(r, []string{"org", "roundId"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	orgname := values[0]
	roundId, err := strconv.Atoi(values[1])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	err = removeRound(orgname, roundId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(server.StrToBytes("Successfully removed round"))
}

func RescheduleRoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(server.StrToBytes("Only POST requests are allowed at this route"))
		return
	}

	values, err := getQueryParams(r, []string{"org", "round", "roundId"})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(server.ErrToBytes(err))
		return
	}

	orgname := values[0]
	roundDate := values[1]
	roundId, err := strconv.Atoi(values[2])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	err = rescheduleRound(orgname, roundDate, roundId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(server.StrToBytes("Succesfully changed date of the round"))
}

func addRound(orgname string, roundDate string) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT MAX(id) FROM rounds")
	if err != nil {
		return err
	}
	defer rows.Close()

	maxRoundId := -1
	for rows.Next() {
		err := rows.Scan(&maxRoundId)

		if err != nil && !strings.Contains(err.Error(), NoMaxRoundErr) {
			return err
		}
		break
	}

	_, err = db.Exec(
		"INSERT INTO rounds (organization, id, scheduled_date) VALUES ($1, $2, $3)",
		orgname,
		maxRoundId+1,
		roundDate,
	)
	if err != nil {
		return err
	}

	return nil
}

func removeRound(orgname string, roundId int) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"DELETE FROM rounds WHERE organization = $1 AND id = $2",
		orgname,
		roundId,
	)
	if err != nil {
		return err
	}

	for {
		result, err := db.Exec(
			"UPDATE rounds SET id = $1 WHERE organization = $2 AND id = $3",
			roundId,
			orgname,
			roundId+1,
		)
		if err != nil {
			return err
		}
		if numRows, _ := result.RowsAffected(); numRows == 0 {
			break
		}

		roundId += 1
	}

	return nil
}

func rescheduleRound(orgname string, roundDate string, roundId int) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE rounds SET scheduled_date = $1 WHERE organization = $2 AND id = $3",
		roundDate,
		orgname,
		roundId,
	)
	if err != nil {
		return err
	}

	return nil
}
