package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

/*
- Read all members of orgs
- Check what cross-match criteria is
- Create Student struct for each member
*/
func readRawCSV(fileName string) [][]string {
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return [][]string{}
	}

	rawStudents := [][]string{
		[]string{},
		[]string{},
		[]string{},
		[]string{},
	}

	reader := csv.NewReader(bufio.NewReader(f))
	for {
		line, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		year, err := strconv.Atoi(line[3])
		if err != nil {
			continue
		}
		rawStudents[3-(year-2019)] = append(
			rawStudents[3-(year-2019)],
			fmt.Sprintf("%s %d", line[0], year),
		)
	}

	return rawStudents
}
