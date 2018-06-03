package main

//go:generate rice embed-go

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/mmcloughlin/professor"

	"git.mills.io/prologic/je"
)

func main() {
	var (
		version bool
		debug   bool

		datadir string
		dburi   string
		bind    string
		threads int
	)

	flag.BoolVar(&version, "v", false, "display version information")
	flag.BoolVar(&debug, "d", false, "enable debug logging")

	flag.StringVar(&datadir, "datadir", "./data", "data directory")
	flag.StringVar(&dburi, "dburi", "memory://", "database to use")
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "[int]:<port> to bind to")
	flag.IntVar(&threads, "threads", 16, "worker threads")

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

	opts := &je.Options{
		Data:    datadir,
		Threads: threads,
	}

	_, err := je.InitData(datadir)
	if err != nil {
		log.Errorf("error initializing data storage: %s", err)
		os.Exit(1)
	}

	db, err := je.InitDB(dburi)
	if err != nil {
		log.Errorf("error initializing database: %s", err)
		os.Exit(1)
	}

	defer db.Close()

	log.Infof("je %s listening on %s", je.FullVersion(), bind)
	je.NewServer(bind, opts).ListenAndServe()
}
