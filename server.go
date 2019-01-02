package main

import (
	"fmt"
	"go/build"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func main() {
	args := os.Args
	if len(args) == 2 && args[1] == "pair" {
		err := createOrganization("ysc", "johnamadeo.daniswara@yale.edu")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(filepath.Join(build.Default.GOPATH, CSVPath, "ysc.csv"))
		_, err := createMembersFromCSV("ysc", filepath.Join(build.Default.GOPATH, CSVPath, "ysc.csv"))
		if err != nil {
			fmt.Println(err)
		}
		err = addRound("ysc", "1999-01-08 04:05:06")
		if err != nil {
			fmt.Println(err)
		}
		err = rescheduleRound("ysc", "1999-10-08 04:05:06", 0)
		if err != nil {
			fmt.Println(err)
		}
		err = addRound("ysc", "2019-01-08 04:05:06")
		if err != nil {
			fmt.Println(err)
		}
		err = runPairingRound("ysc", 0, true)
		if err != nil {
			fmt.Println(err)
		}
		err = runPairingRound("ysc", 1, true)
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
