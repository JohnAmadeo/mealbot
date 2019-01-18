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
	MaxTries  = 5
	NumRounds = 6

	EmailIntro   = "Your Mealbot group this week is:"
	EmailFooter  = "Feel free to reply all in this thread for scheduling. I'm a robot, so I can only read 1's and 0's.\n\n Sent by your friendly neighborhood Mealbot! Learn more about me at https://mealbot-web.herokuapp.com"
	EmailSubject = "Your new Mealbot group"
)

type Pair struct {
	Id1     string
	Id2     string
	ExtraId string
}

func NewPair(name1 string, name2 string) Pair {
	if name1 < name2 {
		return Pair{
			Id1:     name1,
			Id2:     name2,
			ExtraId: "",
		}
	} else {
		return Pair{
			Id1:     name2,
			Id2:     name1,
			ExtraId: "",
		}
	}
}

func (p Pair) String() string {
	if p.ExtraId == "" {
		return fmt.Sprintf("%-30s %-30s", p.Id1, p.Id2)
	} else {
		return fmt.Sprintf("%-30s %-30s %-30s", p.Id1, p.Id2, p.ExtraId)
	}
}

func (p *Pair) AddExtraPerson(id string) error {
	if p.ExtraId != "" {
		return errors.New("Pair already has a 3rd person!")
	}
	p.ExtraId = id
	return nil
}

type Round struct {
	Number           int
	Pairs            map[Pair]bool
	Paired           map[string]bool
	DrawOrderedPairs []Pair // for logging; keeps order in which pairs were made
}

func NewRound(roundNumber int) Round {
	return Round{
		Number:           roundNumber,
		Pairs:            map[Pair]bool{},
		Paired:           map[string]bool{},
		DrawOrderedPairs: []Pair{},
	}
}

func (r *Round) AddPair(student1 string, student2 string) {
	r.Paired[student1] = true
	r.Paired[student2] = true

	pair := NewPair(student1, student2)
	r.Pairs[pair] = true

	r.DrawOrderedPairs = append(r.DrawOrderedPairs, pair)
}

func (r Round) IsPaired(student string) bool {
	if _, ok := r.Paired[student]; ok {
		return true
	}
	return false
}

func (r Round) GetPairForExtraStudent(student Student) (Pair, error) {
	if len(r.Pairs) == 0 {
		return Pair{}, errors.New("No pair of 2 people can be found this round")
	}

	selectedPair := Pair{}
	// Pick random pair as fallback
	for pair, _ := range r.Pairs {
		if pair.ExtraId != "" {
			continue
		}
		selectedPair = pair
		break
	}

	// Pick a pair that contains someone the extra student is least-paired w/
	minCount := r.Number + 1
	for _, partnerId := range student.PartnerIds {
		if student.PairCounts[partnerId] < minCount {
			minCount = student.PairCounts[partnerId]
		}
	}

	for pair, _ := range r.Pairs {
		id1Pairs := student.PairCounts[pair.Id1]
		id2Pairs := student.PairCounts[pair.Id2]
		if id1Pairs == minCount && id2Pairs == minCount {
			selectedPair = pair
			break
		}
	}

	return selectedPair, nil

}

func (r Round) AddExtraStudentToPair(pair Pair, studentId string) error {
	if _, ok := r.Pairs[pair]; !ok {
		return errors.New("Specified pair does not exist in the round")
	}

	for i, currPair := range r.DrawOrderedPairs {
		if currPair == pair {
			r.DrawOrderedPairs[i].AddExtraPerson(studentId)
		}
	}

	delete(r.Pairs, pair)
	pair.AddExtraPerson(studentId)
	r.Pairs[pair] = true
	r.Paired[studentId] = true

	return nil
}

func (r Round) String() string {
	header := fmt.Sprintf("---------\nRound %d\n---------\n", r.Number)

	pairs := []string{}
	for _, pair := range r.DrawOrderedPairs {
		pairs = append(pairs, pair.String())
	}

	return header + strings.Join(pairs, "\n") + "\n"
}

type Student struct {
	Id         string
	Name       string
	Trait      string
	PartnerIds []string // list of ideal partners
	BackupIds  []string // list of backups
	PairCounts map[string]int
}

type StudentMap map[string]Student

func (sm *StudentMap) AddPair(studentId string, partnerId string) int {
	repeats := 0
	if (*sm)[studentId].PairCounts[partnerId] > 0 {
		fmt.Printf("REPEAT: %s %s\n", studentId, partnerId)
		repeats += 1
	}

	(*sm)[studentId].PairCounts[partnerId] += 1
	(*sm)[partnerId].PairCounts[studentId] += 1

	return repeats
}

func (sm *StudentMap) AddExtraStudentToPair(pair Pair, studentId string) int {
	repeats := 0

	if (*sm)[pair.Id1].PairCounts[studentId] > 0 {
		fmt.Printf("REPEAT: %s %s\n", studentId, pair.Id1)
		repeats += 1
	}

	if (*sm)[pair.Id2].PairCounts[studentId] > 0 {
		fmt.Printf("REPEAT: %s %s\n", studentId, pair.Id2)
		repeats += 1
	}

	(*sm)[pair.Id1].PairCounts[studentId] += 1
	(*sm)[pair.Id2].PairCounts[studentId] += 1
	(*sm)[studentId].PairCounts[pair.Id1] += 1
	(*sm)[studentId].PairCounts[pair.Id2] += 1

	return repeats
}

func (sm StudentMap) String() string {
	str := ""
	for studentId, _ := range sm {
		str += fmt.Sprintf("%s\n", studentId)

		str += "Partners:\n"
		for _, partnerId := range sm[studentId].PartnerIds {
			str += fmt.Sprintf("%s ", partnerId)
		}
		str += "\n"

		str += "Backups:\n"
		for _, backupId := range sm[studentId].BackupIds {
			str += fmt.Sprintf("%s ", backupId)
		}
		str += "\n"

		for partnerId, count := range sm[studentId].PairCounts {
			str += fmt.Sprintf("\t%-30s : %d\n", partnerId, count)
		}
		str += "\n"
	}
	return str
}

func (sm StudentMap) Repeats() (map[Pair]bool, int) {
	repeats := map[Pair]bool{}
	for _, student := range sm {
		for partnerId, count := range student.PairCounts {
			if count > 1 {
				repeats[NewPair(student.Id, partnerId)] = true
			}
		}
	}
	return repeats, len(repeats) / 2
}

type Partner struct {
	Id        string
	PairCount int
}

type Partners []Partner

func (p Partners) Len() int           { return len(p) }
func (p Partners) Less(i, j int) bool { return p[i].PairCount < p[j].PairCount }
func (p Partners) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func runPairingRound(orgname string, roundNum int, testMode bool) error {
	students, err := getStudentsFromDB(orgname)
	if err != nil {
		return err
	}

	studentIds := []string{}
	for id, _ := range students {
		studentIds = append(studentIds, id)
	}

	tries := 0
	var round Round
	// retry until a) a round w/o repeats is found, or b) MaxTries is reached
	for {
		studentBytes, _ := json.Marshal(students)
		var tempStudents StudentMap
		json.Unmarshal(studentBytes, &tempStudents)

		round = NewRound(roundNum)
		numRepeats := 0

		// hold out odd student out and add it back in at the end of round
		extraStudentId := ""
		if len(tempStudents)%2 == 1 {
			extraStudentId = studentIds[rand.Intn(len(studentIds))]
		}

		for _, studentId := range studentIds {
			if round.IsPaired(studentId) || studentId == extraStudentId {
				continue
			}

			partners := []Partner{}
			student := tempStudents[studentId]

			findUnpairedPartners(&partners, student, round, extraStudentId)
			if len(partners) == 0 {
				findBackupPartners(&partners, student, round, extraStudentId)
			}
			findLeastPairedPartners(&partners, round)

			partnerId := selectRandomPartner(partners).Id

			round.AddPair(studentId, partnerId)
			numRepeats += tempStudents.AddPair(studentId, partnerId)
		}

		if extraStudentId != "" {
			pair, _ := round.GetPairForExtraStudent(
				tempStudents[extraStudentId],
			)
			round.AddExtraStudentToPair(pair, extraStudentId)
			numRepeats += tempStudents.AddExtraStudentToPair(
				pair,
				extraStudentId,
			)
		}

		tries += 1

		if numRepeats == 0 || tries == MaxTries {
			students = tempStudents
			fmt.Println(round)
			fmt.Printf("%d repeats\n", numRepeats)
			break
		}
	}

	for pair, _ := range round.Pairs {
		toEmails := []string{pair.Id1, pair.Id2}
		toNames := []string{students[pair.Id1].Name, students[pair.Id2].Name}

		if pair.ExtraId != "" {
			toEmails = append(toEmails, pair.ExtraId)
			toNames = append(toNames, students[pair.ExtraId].Name)
		}

		if !testMode {
			err := sendEmails(orgname, toEmails, toNames)
			if err != nil {
				return err
			}
		}
	}

	err = saveRoundInDB(round, students, orgname)
	if err != nil {
		return err
	}

	return nil
}

func findUnpairedPartners(
	partners *[]Partner,
	student Student,
	round Round,
	extraStudentId string,
) {
	for _, partnerId := range student.PartnerIds {
		if round.IsPaired(partnerId) || partnerId == extraStudentId {
			continue
		}

		*partners = append(*partners, Partner{
			Id:        partnerId,
			PairCount: student.PairCounts[partnerId],
		})
	}
}

func findBackupPartners(
	partners *[]Partner,
	student Student,
	round Round,
	extraStudentId string,
) {
	for _, partnerId := range student.BackupIds {
		if student.Id == partnerId ||
			round.IsPaired(partnerId) ||
			partnerId == extraStudentId {
			continue
		}

		*partners = append(*partners, Partner{
			Id:        partnerId,
			PairCount: student.PairCounts[partnerId],
		})
	}
	rand.Shuffle(len(*partners), Partners(*partners).Swap)
}

func findLeastPairedPartners(partners *[]Partner, round Round) {
	minPairs := round.Number + 1
	for _, partner := range *partners {
		if partner.PairCount < minPairs {
			minPairs = partner.PairCount
		}
	}
	leastPairedPartners := []Partner{}
	for _, partner := range *partners {
		if partner.PairCount == minPairs {
			leastPairedPartners = append(leastPairedPartners, partner)
		}
	}
	*partners = leastPairedPartners
}

func selectRandomPartner(partners []Partner) Partner {
	return partners[rand.Intn(len(partners))]
}

func getStudentsFromDB(orgname string) (StudentMap, error) {
	crossMatchTrait, err := getCrossMatchTrait(orgname)
	if err != nil {
		return StudentMap{}, err
	}

	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return StudentMap{}, err
	}

	// rows, err := db.Query(
	// 	"SELECT * FROM members WHERE organization = $1",
	// 	orgname,
	// )
	// if err != nil {
	// 	return StudentMap{}, err
	// }
	// defer rows.Close()
	//
	// students := StudentMap{}
	// for rows.Next() {
	// 	var organization, email, name string
	// 	var metadataJson, pairCountsJson server.JSONB
	//
	// 	err := rows.Scan(
	// 		&organization,
	// 		&email,
	// 		&name,
	// 		&metadataJson,
	// 		&pairCountsJson,
	// 	)
	// 	if err != nil {
	// 		return StudentMap{}, err
	// 	}
	//
	// 	bytes, err := metadataJson.MarshalJSON()
	// 	if err != nil {
	// 		return StudentMap{}, err
	// 	}
	//
	// 	var metadata map[string]string
	// 	err = json.Unmarshal(bytes, &metadata)
	// 	if err != nil {
	// 		return StudentMap{}, err
	// 	}
	//
	// 	bytes, err = pairCountsJson.MarshalJSON()
	// 	if err != nil {
	// 		return StudentMap{}, err
	// 	}
	//
	// 	var pairCounts map[string]int
	// 	err = json.Unmarshal(bytes, &pairCounts)
	// 	if err != nil {
	// 		return StudentMap{}, err
	// 	}

	students := StudentMap{}
	members, err := getMembersFromDB(orgname, true)
	if err != nil {
		return StudentMap{}, err
	}

	for _, member := range members {
		students[member.Email] = Student{
			Id:         member.Email,
			Name:       member.Name,
			Trait:      member.Metadata[crossMatchTrait],
			PartnerIds: []string{},
			BackupIds:  []string{},
			PairCounts: member.PairCounts,
		}
	}

	for i, student := range students {
		partnerIds := []string{}
		backupIds := []string{}
		for j, _ := range students {
			if i == j {
				continue
			}

			if students[i].Trait != students[j].Trait {
				partnerIds = append(partnerIds, students[j].Id)
			} else {
				backupIds = append(backupIds, students[j].Id)
			}
		}
		student.PartnerIds = partnerIds
		student.BackupIds = backupIds
		students[i] = student
	}

	return students, nil
}

func saveRoundInDB(round Round, students StudentMap, orgname string) error {
	db, err := server.CreateDBConnection(LocalDBConnection)
	defer db.Close()
	if err != nil {
		return err
	}

	for pair, _ := range round.Pairs {
		columns := "(organization, id1, id2, extraId, round)"
		placeholder := "($1, $2, $3, $4, $5)"

		_, err := db.Exec(
			fmt.Sprintf("INSERT INTO pairs %s VALUES %s", columns, placeholder),
			orgname,
			pair.Id1,
			pair.Id2,
			pair.ExtraId,
			round.Number,
		)

		if err != nil {
			return err
		}
	}

	for _, student := range students {
		bytes, err := json.Marshal(student.PairCounts)
		if err != nil {
			return err
		}

		_, err = db.Exec(
			"UPDATE members SET pair_counts = $1 WHERE email = $2",
			server.JSONB(bytes),
			student.Id,
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

func sendEmails(orgname string, toEmails []string, toNames []string) error {
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

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	// defer cancel()
	//
	// // Send the message	with a 10 second timeout
	// _, _, err := mg.Send(ctx, message)

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
