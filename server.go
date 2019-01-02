package main

import (
	"fmt"
	"net/http"
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
	fmt.Println("Hello!")
	// createOrganization("ysc", "johnamadeo.daniswara@yale.edu")
	// createMembersFromCSV("ysc", filepath.Join(build.Default.GOPATH, CSVPath, "ysc.csv"))
	// addRound("ysc", "1999-01-08 04:05:06")
	// rescheduleRound("ysc", "1999-10-08 04:05:06", 0)
	// addRound("ysc", "2019-01-08 04:05:06")
	// runPairingRound("ysc", 0, true)
	// runPairingRound("ysc", 1, true)

	// mw := Middleware{
	// 	MiddlewareHandlers: [](func(handler http.Handler) http.Handler){
	// 		GetAuthHandler,
	// 		GetCorsHandler,
	// 	},
	// }
	//
	// serveMux := http.NewServeMux()
	// serveMux.Handle("/members", mw.Apply(MembersHandler))
	// serveMux.Handle("/orgs", mw.Apply(GetOrganizationsHandler))
	// serveMux.Handle("/org", mw.Apply(CreateOrganizationHandler))
	// serveMux.Handle("/crossmatchtrait", mw.Apply(CrossMatchTraitHandler))
	// serveMux.Handle("/rounds", mw.Apply(GetRoundsHandler))
	// serveMux.Handle("/round", mw.Apply(RoundHandler))
	// serveMux.Handle("/pairs", mw.Apply(GetPairsHandler))
	// serveMux.Handle("/", http.FileServer(http.Dir("./static")))
	//
	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080"
	// }
	//
	// log.Fatal(http.ListenAndServe(":"+port, serveMux))
}
