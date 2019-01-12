package main

import (
	"net/http"
)

const (
	// AcessControlAllowOrigin = "http://localhost:3000"
	AccessControlAllowOrigin  = "https://mealbot-web.herokuapp.com"
	AccessControlAllowHeaders = "Authorization, Content-Type, Origin, Accept, token"
)

func GetCorsHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", AccessControlAllowOrigin)
		w.Header().Add("Access-Control-Allow-Headers", AccessControlAllowHeaders)
		w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE")

		if r.Method == "OPTIONS" {
			return
		}

		handler.ServeHTTP(w, r)
	})
}
