package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const (
	Border    = "------------------------"
	MaxTries  = 5
	NumRounds = 6
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
	Id          string
	Year        int
	PartnerIds  []string
	YearmateIds []string
	PairCounts  map[string]int
}

type StudentMap map[string]Student

func (sm *StudentMap) AddPair(studentId string, partnerId string) bool {
	isRepeat := false
	if (*sm)[studentId].PairCounts[partnerId] > 0 {
		fmt.Printf("REPEAT: %s %s\n", studentId, partnerId)
		isRepeat = true
	}

	(*sm)[studentId].PairCounts[partnerId] += 1
	(*sm)[partnerId].PairCounts[studentId] += 1

	return isRepeat
}

func (sm *StudentMap) AddExtraStudentToPair(pair Pair, studentId string) {
	(*sm)[pair.Id1].PairCounts[studentId] += 1
	(*sm)[pair.Id2].PairCounts[studentId] += 1
	(*sm)[studentId].PairCounts[pair.Id1] += 1
	(*sm)[studentId].PairCounts[pair.Id2] += 1
}

func (sm StudentMap) String() string {
	str := ""
	for studentId, _ := range sm {
		str += fmt.Sprintf("%s\n", studentId)
		for partnerId, count := range sm[studentId].PairCounts {
			str += fmt.Sprintf("\t%-30s : %d\n", partnerId, count)
		}
		str += "\n"
	}
	return str
}

func (sm StudentMap) Repeats() ([]Pair, int) {
	repeats := []Pair{}
	for _, student := range sm {
		for partnerId, count := range student.PairCounts {
			if count > 1 {
				repeats = append(repeats, NewPair(student.Id, partnerId))
			}
		}
	}
	return repeats, len(repeats) / 2
}

type Partner struct {
	Id        string
	MealCount int
}

type Partners []Partner

func (p Partners) Len() int           { return len(p) }
func (p Partners) Less(i, j int) bool { return p[i].MealCount < p[j].MealCount }
func (p Partners) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func sortPartnersByMealCount(partners *[]Partner) {
	sort.Sort(Partners(*partners))
}

func findUnpairedPartners(
	partners *[]Partner,
	student Student,
	round Round,
	extraStudentId string,
) {
	for _, partnerId := range student.PartnerIds {
		if !round.IsPaired(partnerId) && partnerId != extraStudentId {
			*partners = append(*partners, Partner{
				Id:        partnerId,
				MealCount: student.PairCounts[partnerId],
			})
		}
	}
}

func findSameYearPartners(partners *[]Partner, student Student, round Round) {
	for _, partnerId := range student.YearmateIds {
		if student.Id == partnerId || round.IsPaired(partnerId) {
			continue
		}
		*partners = append(*partners, Partner{
			Id:        partnerId,
			MealCount: student.PairCounts[partnerId],
		})
	}
	rand.Shuffle(len(*partners), Partners(*partners).Swap)
}

func findLeastPairedPartners(partners *[]Partner, round Round) {
	minMeals := round.Number + 1
	for _, partner := range *partners {
		if partner.MealCount < minMeals {
			minMeals = partner.MealCount
		}
	}
	leastPairedPartners := []Partner{}
	for _, partner := range *partners {
		if partner.MealCount == minMeals {
			leastPairedPartners = append(leastPairedPartners, partner)
		}
	}
	*partners = leastPairedPartners
}

func selectRandomPartner(partners []Partner) Partner {
	return partners[rand.Intn(len(partners))]
}

func main() {
	// TODO: Need to dynamically decide NumRounds & generalize

	rand.Seed(time.Now().UTC().UnixNano())
	raw := readRawCSV("test.csv")
	// raw := [][]string{
	// 	[]string{"A1", "A2"},
	// 	[]string{"B1", "B2"},
	// }

	students := initStudents(raw)
	studentIds := getStudentIds(raw)

	for roundNum := 0; roundNum < NumRounds; roundNum++ {
		tries := 0
		// retry until a) a round w/o repeats is found, or b) MaxTries is reached
		for {
			studentBytes, _ := json.Marshal(students)
			var tempStudents StudentMap
			json.Unmarshal(studentBytes, &tempStudents)

			round := NewRound(roundNum)
			numRepeats := 0

			rand.Shuffle(len(studentIds), func(i int, j int) {
				studentIds[i], studentIds[j] = studentIds[j], studentIds[i]
			})

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
				// fmt.Println("unpaired", studentId, partners)

				if len(partners) == 0 {
					findSameYearPartners(&partners, student, round)
					// fmt.Println("same year", studentId, partners)
				}

				findLeastPairedPartners(&partners, round)

				partnerId := selectRandomPartner(partners).Id

				round.AddPair(studentId, partnerId)
				repeat := tempStudents.AddPair(studentId, partnerId)
				if repeat {
					numRepeats += 1
				}
			}

			if extraStudentId != "" {
				pair, _ := round.GetPairForExtraStudent(tempStudents[extraStudentId])
				tempStudents.AddExtraStudentToPair(pair, extraStudentId)
				round.AddExtraStudentToPair(pair, extraStudentId)
			}

			tries += 1

			if numRepeats == 0 || tries == MaxTries {
				students = tempStudents
				fmt.Println(round)
				break
			}
		}
	}

	_, repeats := students.Repeats()
	fmt.Println(repeats)
}
