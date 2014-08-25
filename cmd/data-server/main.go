package main

import (
	"errors"
	"log"
	"os"

	"github.com/inappcloud/data"
	"github.com/inappcloud/jsonapi"
	"github.com/zenazn/goji"
)

func main() {
	dbUrl := os.Getenv("DATABASE_URL")

	if len(dbUrl) == 0 {
		log.Fatal(errors.New("You must set DATABASE_URL environment variable."))
	}

	db, err := data.Open(dbUrl)

	if err != nil {
		log.Fatal(err)
	}

	goji.Handle("*", data.Mux(db))
	goji.Serve()
}
