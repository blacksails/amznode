package main

import (
	"log"
	"net/http"
	"time"

	"github.com/blacksails/amznode"
	"github.com/blacksails/amznode/pg"
)

func main() {
	var (
		storage amznode.Storage
		err     error
	)
	ticker := time.NewTicker(time.Second)
	tries := 0
	for {
		<-ticker.C
		storage, err = pg.NewFromEnv()
		if err == nil {
			break
		}
		tries++
		if tries > 10 {
			log.Fatal(err)
		}
	}
	ticker.Stop()
	server := amznode.New(storage)
	err = http.ListenAndServe(":8080", server.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
