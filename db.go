package je

import (
	"log"

	"github.com/asdine/storm"
	"github.com/asdine/storm/codec/msgpack"
)

var (
	db *storm.DB
)

func InitDB(path string) *storm.DB {
	var err error

	db, err = storm.Open(path, storm.Codec(msgpack.Codec), storm.Batch())
	if err != nil {
		log.Fatal(err)
	}
	return db
}
