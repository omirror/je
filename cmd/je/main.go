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

		dburi   string
		bind    string
		workers int
	)

	flag.BoolVar(&version, "v", false, "display version information")
	flag.BoolVar(&debug, "d", false, "enable debug logging")

	flag.StringVar(&dburi, "dburi", "memory://", "Database URI")
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "[int]:<port> to bind to")
	flag.IntVar(&workers, "workers", 16, "worker pool size")

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
		Workers: workers,
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
