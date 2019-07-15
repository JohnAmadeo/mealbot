package main

import (
	"fmt"
	"encoding/json"

	"github.com/johnamadeo/server"
)

func createLastRoundWithFromPairCounts() error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT name FROM organizations")
	if err != nil {
		return err
	}

	organizations := []string{}
	for rows.Next() {
		var organization string
		err := rows.Scan(&organization)
		if err != nil {
			return err
		}

		organizations = append(organizations, organization)
	}

	for _, org := range organizations {
		membersMap := map[string]Member{}
		members, err := GetMembersFromDB(org, true)
		if err != nil {
			return err
		}

		for _, member := range members {
			membersMap[member.Email] = member
		}

		pairs, err := getPairsFromDB(org)
		if err != nil {
			return err
		}

		for round, roundPairs := range pairs {
			for _, pairing := range roundPairs {
				membersMap[pairing.Member1.Email].LastRoundWith[pairing.Member2.Email] = round
				membersMap[pairing.Member2.Email].LastRoundWith[pairing.Member1.Email] = round

				if pairing.ExtraMember.Email != "" {
					membersMap[pairing.Member1.Email].LastRoundWith[pairing.ExtraMember.Email] = round
					membersMap[pairing.Member2.Email].LastRoundWith[pairing.ExtraMember.Email] = round
					membersMap[pairing.ExtraMember.Email].LastRoundWith[pairing.Member1.Email] = round
					membersMap[pairing.ExtraMember.Email].LastRoundWith[pairing.Member2.Email] = round
				}
			}
		}

		for _, member := range membersMap {
			bytes, err := json.Marshal(member.LastRoundWith)
			if err != nil {
				return err
			}

			_, err = db.Exec(
				"UPDATE members SET last_round_with = $1 WHERE organization = $2 AND email = $3",
				server.JSONB(bytes),
				org,
				member.Email,
			)
			if err != nil {
				return err
			}

		}
	}

	fmt.Println("DONE!")
	return nil
}