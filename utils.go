package main

import (
	"errors"
	"net/http"
)

func getQueryParam(r *http.Request, key string) (string, error) {
	queries, ok := r.URL.Query()[key]
	if !ok || len(queries) > 1 {
		return "", errors.New("Request query parameters must contain " + key)
	}
	return queries[0], nil
}
