package main

import (
	"github.com/johnamadeo/server"
)

const DuplicateKeyErr = "duplicate key value violates unique constraint"

var LocalDBConnection = server.LocalDBConnection{
	User:   "johnamadeodaniswara",
	DBName: "mealbot",
}
