package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/johnamadeo/server"
	mailgun "github.com/mailgun/mailgun-go"
)

const (
	// MaxTries is the max. number of times to attempt re-matching if a non-optimal matching is initially made
	MaxTries = 50
	// RecentRoundRange : max. no. of rounds between 2 pairings that is considered recent
	RecentRoundRange = 5
	// EmailIntro : Text to put in the beginning of the email body
	EmailIntro = "Your Mealbot group this week is:"
	// EmailFooter : Text to put at the end of the email body
	EmailFooter = "Feel free to reply all in this thread for scheduling. I'm a robot, so I can only read 1's and 0's.\n\n Sent by your friendly neighborhood Mealbot! Learn more about me at https://mealbot-web.herokuapp.com"
	// EmailSubject : Subject of the email
	EmailSubject = "Your new Mealbot group"
)

// Pair :
type Pair struct {
	ID1     string
	ID2     string
	ExtraID string
}

// NewPair :
func NewPair(name1 string, name2 string) Pair {
	if name1 < name2 {
		return Pair{
			ID1:     name1,
			ID2:     name2,
			ExtraID: "",
		}
	}
	return Pair{
		ID1:     name2,
		ID2:     name1,
		ExtraID: "",
	}
}

func (p Pair) String() string {
	return fmt.Sprintf("%-40s %-40s %-40s", p.ID1, p.ID2, p.ExtraID)
}

// Round : Data structure for storing info on one round of pairings between members
type Round struct {
	Number           int
	Pairs            map[Pair]bool
	Paired           map[string]bool
	DrawOrderedPairs []Pair // for logging; keeps order in which pairs were made
}

// NewRound :
func NewRound(roundNumber int) Round {
	return Round{
		Number:           roundNumber,
		Pairs:            map[Pair]bool{},
		Paired:           map[string]bool{},
		DrawOrderedPairs: []Pair{},
	}
}

// AddPair : Record a new pairing made during the round
func (r *Round) AddPair(member1 string, member2 string) {
	r.Paired[member1] = true
	r.Paired[member2] = true

	pair := NewPair(member1, member2)
	r.Pairs[pair] = true

	r.DrawOrderedPairs = append(r.DrawOrderedPairs, pair)
}

// IsPaired : Check if member has already been paired this round
func (r Round) IsPaired(member string) bool {
	_, ok := r.Paired[member]
	return ok
}

func (r Round) String() string {
	header := fmt.Sprintf("---------\nRound %d\n---------\n", r.Number)

	pairs := []string{}
	for _, pair := range r.DrawOrderedPairs {
		pairs = append(pairs, pair.String())
	}

	return header + strings.Join(pairs, "\n") + "\n"
}

// MinimalMember : A pared down version of the Member struct (see members.go) for the pairing process
type MinimalMember struct {
	ID            string
	Name          string
	Trait         string
	LastRoundWith map[string]int
}

// MembersMap :
type MembersMap map[string]MinimalMember

// UpdateLastRoundWith : Updates the last round 2 members were matched with each other
func (mm *MembersMap) UpdateLastRoundWith(id1 string, id2 string, roundNum int) {
	(*mm)[id1].LastRoundWith[id2] = roundNum
	(*mm)[id2].LastRoundWith[id1] = roundNum
}

func (mm MembersMap) String() string {
	str := "X's last round with Y\n"
	for memberID := range mm {
		str += fmt.Sprintf("%s\n", memberID)

		for partnerID, lastRound := range mm[memberID].LastRoundWith {
			str += fmt.Sprintf("\t%-30s : %d\n", partnerID, lastRound)
		}
		str += "\n"
	}
	return str
}

type RandomIntGenerator func(int) int

func runPairingAlgorithm(members MembersMap, roundNum int, genRandomInt RandomIntGenerator) (MembersMap, Round) {
	memberIds := []string{}
	for id := range members {
		memberIds = append(memberIds, id)
	}

	tries := 0
	var round Round
	// retry until a) a round w/o repeats is found, or b) MaxTries is reached
	for {
		// Make a deep copy of the members map
		bytes, _ := json.Marshal(members)
		var tempMembers MembersMap
		json.Unmarshal(bytes, &tempMembers)

		round = NewRound(roundNum)
		numRecentMatches := 0

		// hold out odd member out and add it back in at the end of round
		var extraMemberID string
		if len(tempMembers)%2 == 1 {
			extraMemberID = memberIds[genRandomInt(len(memberIds))]
		}

		for _, memberID := range memberIds {
			if round.IsPaired(memberID) || memberID == extraMemberID {
				continue
			}

			member := tempMembers[memberID]

			var goodCandidates, okCandidates, badCandidates []MinimalMember
			for candidateID, lastRound := range member.LastRoundWith {
				if round.IsPaired(candidateID) || candidateID == extraMemberID {
					continue
				}

				notRecentlyMatched := lastRound == -1 || round.Number-lastRound > RecentRoundRange
				hasDifferentTraits := member.Trait == members[candidateID].Trait

				switch {
				case notRecentlyMatched && hasDifferentTraits:
					goodCandidates = append(goodCandidates, members[candidateID])
				case notRecentlyMatched && !hasDifferentTraits:
					okCandidates = append(okCandidates, members[candidateID])
				default:
					badCandidates = append(badCandidates, members[candidateID])
				}
			}

			var partner MinimalMember
			switch {
			case len(goodCandidates) > 0:
				partner = selectRandomPartner(genRandomInt, goodCandidates)
			case len(okCandidates) > 0:
				partner = selectRandomPartner(genRandomInt, okCandidates)
			default:
				partner = selectRandomPartner(genRandomInt, badCandidates)
				numRecentMatches++
			}

			round.AddPair(memberID, partner.ID)
			tempMembers.UpdateLastRoundWith(memberID, partner.ID, round.Number)
		}

		tries++

		if numRecentMatches == 0 || tries == MaxTries {
			members = tempMembers
			fmt.Println(round)
			fmt.Printf("%d recent matches\n", numRecentMatches)
			break
		}
	}

	return members, round
}

func runPairingRound(orgname string, roundNum int, testMode bool) error {
	members, err := getMinimalMembersFromDB(orgname)
	if err != nil {
		return err
	}

	members, round := runPairingAlgorithm(members, roundNum, rand.Intn)

	if !testMode {
		err = sendEmails(orgname, round, members)
		if err != nil {
			return err
		}
	}

	err = saveRoundInDB(round, members, orgname)
	if err != nil {
		return err
	}

	return nil
}

func sendEmails(orgname string, round Round, members MembersMap) error {
	for pair := range round.Pairs {
		toEmails := []string{pair.ID1, pair.ID2}
		toNames := []string{members[pair.ID1].Name, members[pair.ID2].Name}

		err := sendEmail(orgname, toEmails, toNames)
		if err != nil {
			return err
		}
	}

	return nil
}

func selectRandomPartner(genRandomInt RandomIntGenerator, members []MinimalMember) MinimalMember {
	return members[genRandomInt(len(members))]
}

func getMinimalMembersFromDB(orgname string) (MembersMap, error) {
	crossMatchTrait, err := GetCrossMatchTrait(orgname)
	if err != nil {
		return MembersMap{}, err
	}

	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return MembersMap{}, err
	}

	minimalMembers := MembersMap{}
	members, err := GetMembersFromDB(orgname, true)
	if err != nil {
		return MembersMap{}, err
	}

	for _, member := range members {
		minimalMembers[member.Email] = MinimalMember{
			ID:            member.Email,
			Name:          member.Name,
			Trait:         member.Metadata[crossMatchTrait],
			LastRoundWith: member.LastRoundWith,
		}
	}

	return minimalMembers, nil
}

func saveRoundInDB(round Round, members MembersMap, orgname string) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	for pair := range round.Pairs {
		columns := "(organization, id1, id2, extraId, round)"
		placeholder := "($1, $2, $3, $4, $5)"

		_, err := db.Exec(
			fmt.Sprintf("INSERT INTO pairs %s VALUES %s", columns, placeholder),
			orgname,
			pair.ID1,
			pair.ID2,
			pair.ExtraID,
			round.Number,
		)

		if err != nil {
			return err
		}
	}

	for _, member := range members {
		bytes, err := json.Marshal(member.LastRoundWith)
		if err != nil {
			return err
		}

		_, err = db.Exec(
			"UPDATE members SET last_round_with = $1 WHERE organization = $2 AND email = $3",
			server.JSONB(bytes),
			orgname,
			member.ID,
		)
		if err != nil {
			return err
		}
	}

	_, err = db.Exec(
		"UPDATE rounds SET done = $1 WHERE organization = $2 AND id = $3",
		true,
		orgname,
		round.Number,
	)
	if err != nil {
		return err
	}

	return nil
}

func sendEmail(orgname string, toEmails []string, toNames []string) error {
	smtpAddress, ok := os.LookupEnv("MAILGUN_SMTP_LOGIN")
	if !ok {
		return errors.New("environment variable MAILGUN_SMTP_LOGIN not set")
	}
	domain, ok := os.LookupEnv("MAILGUN_DOMAIN")
	if !ok {
		return errors.New("environment variable MAILGUN_DOMAIN not set")
	}
	apiKey, ok := os.LookupEnv("MAILGUN_API_KEY")
	if !ok {
		return errors.New("environment variable MAILGUN_API_KEY not set")
	}

	from := fmt.Sprintf("%s Mealbot <%s>", orgname, smtpAddress)
	text := fmt.Sprintf(
		"%s \n%s\n\n %s",
		EmailIntro,
		strings.Join(toNames, "\r\n"),
		EmailFooter,
	)

	mg := mailgun.NewMailgun(domain, apiKey)
	message := mg.NewMessage(
		from,
		EmailSubject,
		text,
		toEmails...,
	)

	_, _, err := mg.Send(message)

	if err != nil {
		return err
	}

	fmt.Println("Emails queued on Mailgun!")
	return nil
}

func runPairingScheduler(testMode bool) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	rows, err := db.Query(
		"SELECT organization, id FROM rounds WHERE scheduled_date < now() AT TIME ZONE 'utc' AND done = false",
	)
	if err != nil {
		return err
	}

	for rows.Next() {
		var orgname string
		var roundNum int
		err := rows.Scan(&orgname, &roundNum)
		if err != nil {
			return err
		}

		err = runPairingRound(orgname, roundNum, testMode)
		if err != nil {
			return err
		}
	}

	return nil
}
