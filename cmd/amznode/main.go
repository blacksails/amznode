package main

import (
	"log"
	"net/http"

	"github.com/blacksails/amznode"
	"github.com/blacksails/amznode/pg"
)

func main() {
	storage, err := pg.NewFromEnv()
	if err != nil {
		log.Fatal(err)
	}
	server := amznode.New(storage)
	err = http.ListenAndServe(":8080", server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
