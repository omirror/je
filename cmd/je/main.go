package main

//go:generate rice embed-go

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/mmcloughlin/professor"

	"github.com/prologic/je"
)

func main() {
	var (
		version bool
		debug   bool

		datadir string
		dburi   string
		bind    string
		threads int
		backlog int
	)

	flag.BoolVar(&version, "v", false, "display version information")
	flag.BoolVar(&debug, "d", false, "enable debug logging")

	flag.StringVar(&datadir, "datadir", "./data", "data directory")
	flag.StringVar(&dburi, "dburi", "memory://", "database to use")
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "[int]:<port> to bind to")
	flag.IntVar(&threads, "threads", runtime.NumCPU(), "worker threads")
	flag.IntVar(&backlog, "backlog", runtime.NumCPU()*2, "backlog size")

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
		Backlog: backlog,
	}

	metrics := je.InitMetrics("je")

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

	server := je.NewServer(bind, opts)
	server.AddRoute("GET", "/metrics", metrics.Handler())

	log.Infof("je %s listening on %s", je.FullVersion(), bind)
	server.ListenAndServe()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	log.Infof("shuting down...")
	server.Shutdown()
}
