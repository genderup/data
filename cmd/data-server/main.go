package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/inappcloud/data"
	_ "github.com/lib/pq"
	"github.com/zenazn/goji"
)

func main() {
	url := os.Getenv("DATABASE_URL")

	if len(url) == 0 {
		log.Fatal(errors.New("You must set DATABASE_URL environment variable."))
	}

	db, err := sql.Open("postgres", url)

	if err != nil {
		log.Fatal(err)
	}

	goji.Handle("*", data.Mux(db))
	goji.Serve()
}
