package main

import (
	"encoding/json"
	"net/http"

	"github.com/johnamadeo/server"
)

type GetPairsResponsePair struct {
	Member1 Member `json:"member1"`
	Member2 Member `json:"member2"`
}

type GetPairsResponse struct {
	RoundPairs [][]GetPairsResponsePair `json:"roundPairs"`
}

func GetPairsHandler(w http.ResponseWriter, r *http.Request) {
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

	roundPairs, err := getPairsFromDB(orgname)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(server.ErrToBytes(err))
		return
	}

	resp := GetPairsResponse{
		RoundPairs: roundPairs,
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

func getPairsFromDB(orgname string) ([][]GetPairsResponsePair, error) {
	roundPairs := [][]GetPairsResponsePair{}

	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return roundPairs, err
	}

	roundNum := 0

	for {
		pairRows, err := db.Query(
			"SELECT id1, id2 FROM pairs WHERE organization = $1 AND round = $2",
			orgname,
			roundNum,
		)
		if err != nil {
			return roundPairs, err
		}

		pairs := []GetPairsResponsePair{}
		numRows := 0
		for pairRows.Next() {
			var id1, id2 string
			err := pairRows.Scan(&id1, &id2)
			if err != nil {
				return roundPairs, err
			}

			member, err := getMemberFromDB(orgname, id1)
			if err != nil {
				return roundPairs, err
			}
			member1 := Member{
				Name:  member.Name,
				Email: member.Email,
			}

			member, err = getMemberFromDB(orgname, id2)
			if err != nil {
				return roundPairs, err
			}
			member2 := Member{
				Name:  member.Name,
				Email: member.Email,
			}

			pairs = append(pairs, GetPairsResponsePair{
				Member1: member1,
				Member2: member2,
			})

			numRows += 1
		}

		if numRows == 0 {
			break
		}

		roundPairs = append(roundPairs, pairs)
		roundNum += 1
	}

	return roundPairs, nil
}
