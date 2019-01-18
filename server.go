package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

type Middleware struct {
	MiddlewareHandlers [](func(handler http.Handler) http.Handler)
}

func (mw *Middleware) Apply(
	coreHandler func(w http.ResponseWriter, r *http.Request),
) http.Handler {
	handler := http.Handler(http.HandlerFunc(coreHandler))
	for _, nextHandler := range mw.MiddlewareHandlers {
		handler = nextHandler(handler)
	}

	return handler
}

func (mw Middleware) ApplyFake(
	coreHandler func(w http.ResponseWriter, r *http.Request),
) http.Handler {
	return http.Handler(http.HandlerFunc(coreHandler))
}

func runTestSequence(testMode bool) {
	err := createOrganization("test", "johnamadeo.daniswara@yale.edu")
	if err != nil {
		fmt.Println(err)
	}

	_, err = createMembersFromCSV("test", "./csv/test_john.csv")
	if err != nil {
		fmt.Println(err)
	}

	err = addRound("test", "2019-01-02 15:55:00")
	if err != nil {
		fmt.Println(err)
	}
	// actually send out emails!
	err = runPairingRound("test", 0, testMode)
	if err != nil {
		fmt.Println(err)
	}

	_, err = createMembersFromCSV("test", "./csv/test_john2.csv")
	if err != nil {
		fmt.Println(err)
	}

	err = addRound("test", "2019-01-02 19:55:00")
	if err != nil {
		fmt.Println(err)
	}
	// actually send out emails!
	err = runPairingRound("test", 1, testMode)
	if err != nil {
		fmt.Println(err)
	}

	_, err = createMembersFromCSV("test", "./csv/test_john3.csv")
	if err != nil {
		fmt.Println(err)
	}

	err = addRound("test", "2019-01-02 20:55:00")
	if err != nil {
		fmt.Println(err)
	}
	// actually send out emails!
	err = runPairingRound("test", 2, testMode)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	args := os.Args
	if len(args) == 2 && args[1] == "pair" {
		// runTestSequence(true)
		err := runPairingScheduler(false)
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	mw := Middleware{
		MiddlewareHandlers: [](func(handler http.Handler) http.Handler){
			GetAuthHandler,
			GetCorsHandler,
		},
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/members", mw.Apply(MembersHandler))
	serveMux.Handle("/orgs", mw.Apply(GetOrganizationsHandler))
	serveMux.Handle("/org", mw.Apply(CreateOrganizationHandler))
	serveMux.Handle("/crossmatchtrait", mw.Apply(CrossMatchTraitHandler))
	serveMux.Handle("/rounds", mw.Apply(GetRoundsHandler))
	serveMux.Handle("/round", mw.Apply(RoundHandler))
	serveMux.Handle("/pairs", mw.Apply(GetPairsHandler))
	serveMux.Handle("/", http.FileServer(http.Dir("./static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(":"+port, serveMux))
}
