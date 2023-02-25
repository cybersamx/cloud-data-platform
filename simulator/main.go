package main

import (
	"log"
)

func main() {
	if err := riderDataFromS3(); err != nil {
		log.Panic(err)
	}

	if err := initDB(); err != nil {
		log.Panic(err)
	}
}
