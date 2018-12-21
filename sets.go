package main

import "fmt"

type PairSet map[Pair]bool

// Change to a Stringer interface so we can print from fmt.Println
func (ps PairSet) Print() {
	fmt.Println("Pairs:")
	for pair, _ := range ps {
		fmt.Printf("%s %s\n", pair.name1, pair.name2)
	}
	fmt.Println("")
}

type StringSet struct {
	set map[string]bool
}

func NewStringSet() *StringSet {
	return &StringSet{make(map[string]bool)}
}

func (s *StringSet) Add(str string) {
	s.set[str] = true
}
