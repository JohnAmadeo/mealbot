package main

func getPartnerIds(rawStudents [][]string, studentYear int) []string {
	partners := []string{}
	for year, _ := range rawStudents {
		if studentYear != year {
			partners = append(partners, rawStudents[year]...)
		}
	}
	return partners
}

func initPairCounts(partners []string) map[string]int {
	pairCounts := map[string]int{}
	for _, studentId := range partners {
		pairCounts[studentId] = 0
	}
	return pairCounts
}

func getNumStudents(rawStudents [][]string) int {
	num := 0
	for _, yearStudents := range rawStudents {
		num += len(yearStudents)
	}
	return num
}

func getStudentIds(students [][]string) []string {
	flatStudents := []string{}
	for _, yearStudents := range students {
		for _, studentId := range yearStudents {
			flatStudents = append(flatStudents, studentId)
		}
	}
	return flatStudents
}

func initStudents(rawStudents [][]string) StudentMap {
	students := map[string]Student{}
	for year, yearStudents := range rawStudents {
		for _, studentId := range yearStudents {
			partnerIds := getPartnerIds(rawStudents, year)
			pairCounts := initPairCounts(partnerIds)
			yearmateIds := []string{}
			for _, yearStudentId := range yearStudents {
				if yearStudentId != studentId {
					yearmateIds = append(yearmateIds, yearStudentId)
				}
			}

			students[studentId] = Student{
				Id:          studentId,
				Year:        year,
				PartnerIds:  partnerIds,
				YearmateIds: yearmateIds,
				PairCounts:  pairCounts,
			}
		}
	}

	return students
}
