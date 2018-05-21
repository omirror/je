package main

//go:generate rice embed-go

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/mmcloughlin/professor"
	"github.com/namsral/flag"

	"git.mills.io/prologic/je"
)

var (
	cfg je.Config
)

func main() {
	var (
		version bool
		debug   bool
		config  string
		dbpath  string
		bind    string
	)

	flag.BoolVar(&version, "v", false, "display version information")
	flag.BoolVar(&debug, "d", false, "enable debug logging")

	flag.StringVar(&config, "config", "", "config file")
	flag.StringVar(&dbpath, "dbpath", "je.db", "Database path")
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "[int]:<port> to bind to")

	flag.Parse()

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if version {
		fmt.Printf("je v%s", je.FullVersion())
		os.Exit(0)
	}

	if debug {
		go professor.Launch(":6060")
	}

	db := je.InitDB(dbpath)
	defer db.Close()

	log.Infof("je %s listening on %s", je.FullVersion(), bind)
	je.NewServer(bind, cfg).ListenAndServe()
}
