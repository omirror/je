package main

//go:generate rice embed-go

import (
	"fmt"
	"log"
	"os"

	"github.com/asdine/storm"
	"github.com/namsral/flag"

	"github.com/prologic/je"
)

var (
	cfg je.Config
	db  *storm.DB
)

func main() {
	var (
		version bool
		config  string
		dbpath  string
		bind    string
	)

	flag.BoolVar(&version, "v", false, "display version information")
	flag.StringVar(&config, "config", "", "config file")
	flag.StringVar(&dbpath, "dbpath", "je.db", "Database path")
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "[int]:<port> to bind to")
	flag.Parse()

	if version {
		fmt.Printf("je v%s", FullVersion())
		os.Exit(0)
	}

	je.NewServer(bind, cfg).ListenAndServe()
}
