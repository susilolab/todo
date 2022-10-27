package models

import (
	"log"

	"github.com/boltdb/bolt"
)

var DB *bolt.DB

func SetupDB(dbPath string) {
	db, err := bolt.Open(dbPath, 0777, nil)
	if err != nil {
		log.Fatalf("Tidak dapat membuka db %s: %s\n", dbPath, err.Error())
	}

	DB = db
}
