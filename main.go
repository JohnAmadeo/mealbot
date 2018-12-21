package main

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/kr/pretty"
)

type Partner struct {
	name      string
	mealCount int
}

type StudentHistory struct {
	name       string
	year       int
	partners   []string
	mealCounts map[string]int
}

type Pair struct {
	name1 string
	name2 string
}

type Partners []Partner

func (p Partners) Len() int           { return len(p) }
func (p Partners) Less(i, j int) bool { return p[i].mealCount < p[j].mealCount }
func (p Partners) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func makePair(name1 string, name2 string) Pair {
	if name1 < name2 {
		return Pair{
			name1: name1,
			name2: name2,
		}
	} else {
		return Pair{
			name1: name2,
			name2: name1,
		}
	}
}

func getPartners(students [][]string, studentYear int) []string {
	partners := []string{}
	for year, _ := range students {
		if studentYear != year {
			partners = append(partners, students[year]...)
		}
	}
	return partners
}

func initHistory(partners []string) map[string]int {
	history := map[string]int{}
	for _, student := range partners {
		history[student] = 0
	}
	return history
}

func getNumStudents(students [][]string) int {
	num := 0
	for _, yearStudents := range students {
		num += len(yearStudents)
	}
	return num
}

func getStudentList(students [][]string) []string {
	flatStudents := []string{}
	for _, yearStudents := range students {
		for _, student := range yearStudents {
			flatStudents = append(flatStudents, student)
		}
	}
	return flatStudents
}

func getStudentHistories(students [][]string) map[string]StudentHistory {
	studentHistories := map[string]StudentHistory{}
	for year, yearStudents := range students {
		for _, student := range yearStudents {
			partners := getPartners(students, year)
			history := initHistory(partners)

			studentHistories[student] = StudentHistory{
				name:       student,
				year:       year,
				partners:   partners,
				mealCounts: history,
			}
		}
	}
	return studentHistories
}

func main() {
	// pairs := map[Pair]bool{}

	// pairs := map[string][]string{}

	raw := [][]string{
		[]string{"A1", "A2", "A3", "A4"},
		[]string{"B1", "B2"},
		[]string{"C1"},
		[]string{"D1", "D2", "D3"},

		// []string{"A1", "A2", "A3", "A4"},
		// []string{"B1", "B2", "B3", "B4"},
		// []string{"C1", "C2", "C3", "C4"},
		// []string{"D1", "D2", "D3", "D4"},
	}

	studentHistories := getStudentHistories(raw)
	students := getStudentList(raw)
	studentsByYear := raw

	fmt.Println(studentsByYear)

	rounds := 5
	for i := 0; i < rounds; i++ {
		fmt.Printf("-----------------\nRound %d\n-----------------\n", i)
		pairs := PairSet{}
		paired := map[string]bool{}

		// oddStudentOut := ""
		// if len(students)%2 == 1 {
		// 	oddStudentOut = flatStudents[rand.Intn(numStudents)]
		// }

		rand.Shuffle(len(students), func(i int, j int) {
			students[i], students[j] = students[j], students[i]
		})

		for _, student := range students {
			if _, ok := paired[student]; ok {
				continue
			}

			partnersWCount := []Partner{}
			studentHistory := studentHistories[student]
			// add all potential partners that don't have a meal pairing already
			for _, partner := range studentHistory.partners {
				if _, ok := paired[partner]; !ok {
					partnersWCount = append(partnersWCount, Partner{
						name:      partner,
						mealCount: studentHistory.mealCounts[partner],
					})
				}
			}

			// sort partners in ascending order of no. of meals had
			sort.Sort(Partners(partnersWCount))

			// fmt.Println(student, partnersWCount)

			// filter out top 50% of partners in terms of no. of meals had
			if len(partnersWCount) > 0 {
				partnersWCount = partnersWCount[:(len(partnersWCount)/2)+1]
			}

			// fmt.Println(student, partnersWCount)

			// pair student with a same-year partner since all other possible
			// partners are not available
			if len(partnersWCount) == 0 {
				for _, sameYearStudent := range studentsByYear[studentHistory.year] {
					if student == sameYearStudent {
						continue
					}
					if _, ok := paired[sameYearStudent]; ok {
						continue
					}
					partnersWCount = append(partnersWCount, Partner{
						name:      sameYearStudent,
						mealCount: studentHistory.mealCounts[sameYearStudent],
					})
				}
				rand.Shuffle(len(partnersWCount), Partners(partnersWCount).Swap)
			}

			// fmt.Println("Same year partners:", student, partnersWCount)

			minMeals := rounds + 1
			for _, partner := range partnersWCount {
				if partner.mealCount < minMeals {
					minMeals = partner.mealCount
				}
			}
			newPartnersWCount := []Partner{}
			for _, partner := range partnersWCount {
				if partner.mealCount == minMeals {
					newPartnersWCount = append(newPartnersWCount, partner)
				}
			}
			partnersWCount = newPartnersWCount

			// fmt.Println(student, partnersWCount)

			// partner := partnersWCount[rand.Intn(len(partnersWCount))].name
			partner := partnersWCount[0].name

			studentHistories[student].mealCounts[partner] += 1
			studentHistories[partner].mealCounts[student] += 1

			paired[student] = true
			paired[partner] = true

			pairs[makePair(student, partner)] = true
			fmt.Println(student, partner)

			// fmt.Println(student)
			// pairs.Print()
			// fmt.Println(paired)
		}

		// pairs.Print()
		// pretty.Println(studentHistories)
		// pretty.Println(pairs)
	}

	pretty.Println(studentHistories)
}
