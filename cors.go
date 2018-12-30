package main

import (
	"net/http"
)

const (
	AcessControlAllowOrigin   = "http://localhost:3000" // change to Heroku
	AccessControlAllowHeaders = "Authorization, Content-Type, Origin, Accept, token"
)

func GetCorsHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", AcessControlAllowOrigin)
		w.Header().Add("Access-Control-Allow-Headers", AccessControlAllowHeaders)

		if r.Method == "OPTIONS" {
			return
		}

		handler.ServeHTTP(w, r)
	})
}