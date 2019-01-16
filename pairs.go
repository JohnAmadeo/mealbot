package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/johnamadeo/server"
)

type GetPairsResponsePair struct {
	Member1     Member `json:"member1"`
	Member2     Member `json:"member2"`
	ExtraMember Member `json:"extraMember"`
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

	members, err := getMemberFromDBInPairFormat(orgname)
	membersMap := map[string]Member{}
	for _, member := range members {
		membersMap[member.Email] = member
	}

	round := 0

	for {
		pairRows, err := db.Query(
			"SELECT id1, id2, extraId FROM pairs WHERE organization = $1 AND round = $2",
			orgname,
			round,
		)
		if err != nil {
			return roundPairs, err
		}

		numPairs := 0
		pairs := []GetPairsResponsePair{}
		for pairRows.Next() {
			var id1, id2 string
			var extraId sql.NullString
			err := pairRows.Scan(&id1, &id2, &extraId)
			if err != nil {
				return roundPairs, err
			}

			if extraId.Valid {
				pairs = append(pairs, GetPairsResponsePair{
					Member1:     membersMap[id1],
					Member2:     membersMap[id2],
					ExtraMember: membersMap[extraId.String],
				})
			} else {
				pairs = append(pairs, GetPairsResponsePair{
					Member1: membersMap[id1],
					Member2: membersMap[id2],
				})
			}

			numPairs += 1
		}

		if numPairs == 0 {
			break
		}

		roundPairs = append(roundPairs, pairs)
		round += 1
	}

	return roundPairs, nil
}
