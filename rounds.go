package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/johnamadeo/server"
)

const (
	// NoMaxRoundErr : TODO
	NoMaxRoundErr = "converting driver.Value type <nil>"

	// TimestampFormat : Postgres timestamp string template patterns can be found in https://www.postgresql.org/docs/8.1/functions-formatting.html
	TimestampFormat = "YYYY-MM-DD HH24:MI:ssZ"
)

// GetRoundsResponse : Data structure for storing the dates for each round
type GetRoundsResponse struct {
	Rounds []string `json:"rounds"`
}

// AddRoundHandler : HTTP handler for scheduling a new round on a certain date
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

// GetRoundsHandler : HTTP handler for retrieving the dates for all rounds scheduled
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

// RoundHandler : Combined HTTP handler for rounds
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

// RemoveRoundHandler : HTTP handler for cancelling a previously scheduled round
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
	roundID, err := strconv.Atoi(values[1])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	err = removeRound(orgname, roundID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(server.StrToBytes("Successfully removed round"))
}

// RescheduleRoundHandler : HTTP handler for rescheduling the date of a particular round
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
	roundID, err := strconv.Atoi(values[2])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	err = rescheduleRound(orgname, roundDate, roundID)
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

	maxRoundID := -1
	for rows.Next() {
		err := rows.Scan(&maxRoundID)

		if err != nil && !strings.Contains(err.Error(), NoMaxRoundErr) {
			return err
		}
		break
	}

	_, err = db.Exec(
		"INSERT INTO rounds (organization, id, scheduled_date, done) VALUES ($1, $2, $3, $4)",
		orgname,
		maxRoundID+1,
		roundDate,
		false,
	)
	if err != nil {
		return err
	}

	return nil
}

func removeRound(orgname string, roundID int) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"DELETE FROM rounds WHERE organization = $1 AND id = $2",
		orgname,
		roundID,
	)
	if err != nil {
		return err
	}

	for {
		result, err := db.Exec(
			"UPDATE rounds SET id = $1 WHERE organization = $2 AND id = $3",
			roundID,
			orgname,
			roundID+1,
		)
		if err != nil {
			return err
		}
		if numRows, _ := result.RowsAffected(); numRows == 0 {
			break
		}

		roundID++
	}

	return nil
}

func rescheduleRound(orgname string, roundDate string, roundID int) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"UPDATE rounds SET scheduled_date = $1 WHERE organization = $2 AND id = $3",
		roundDate,
		orgname,
		roundID,
	)
	if err != nil {
		return err
	}

	return nil
}
