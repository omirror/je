package je

import (
	"log"

	"github.com/asdine/storm"
)

var (
	db *storm.DB
)

func InitDB(path string) *storm.DB {
	var err error
	db, err = storm.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
