package main

import (
	"log"
)

func main() {
	db, err := connectDB()
	if err != nil {
		log.Panic(err)
	}

	if err := initDB(db); err != nil {
		log.Panic(err)
	}

	if err := riderDataFromS3(db); err != nil {
		log.Panic(err)
	}

	if err := tripDataFromS3(db); err != nil {
		log.Panic(err)
	}
}
