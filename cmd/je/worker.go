package main

import (
	"net/http"
	"os"
	"os/signal"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/prologic/je"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var workerCmd = &cobra.Command{
	Use:     "worker [jobtype]",
	Aliases: []string{"agent", "minion"},
	Short:   "Run the worker agent",
	Long: `This runs the worker daemon that listens for new jobs on a queue,
processes them concurrently and sends results back to the logger. The worker
by default listens for jobs on a glocal topic called * -- Workers can be
packaged and configured to operate and run specific job types by specifying
the type of jobs as the [jobtype] argument.`,
	Args:      cobra.MaximumNArgs(1),
	ValidArgs: []string{"jobtype"},
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(runWorker(cmd, args...))
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)

	workerCmd.Flags().StringP(
		"bind", "b", ":9000",
		"[int]:<port> to bind to",
	)

	workerCmd.Flags().StringP(
		"data", "d", "je+http://localhost:8000/",
		"where to persist job data",
	)

	workerCmd.Flags().StringP(
		"logger", "l", "msgbus+http://localhost:8001/",
		"logger to use to send log streams",
	)

	workerCmd.Flags().StringP(
		"queue", "q", "msgbus+http://localhost:8001/",
		"message queue to use for subscribing to jobs and publishing results",
	)

	workerCmd.Flags().StringP(
		"store", "s", "je+http://localhost:8000/",
		"database store to persist job metadata",
	)

	workerCmd.Flags().IntP(
		"workers", "W", runtime.NumCPU(),
		"number of workers",
	)
	workerCmd.Flags().IntP(
		"buffer", "B", runtime.NumCPU()*2,
		"number of work items to buffer",
	)
}

func runWorker(cmd *cobra.Command, args ...string) int {
	var err error

	debug := viper.GetBool("debug")

	bind, err := cmd.Flags().GetString("bind")
	if err != nil {
		log.Errorf("error getting -b/--bind flag: %s", err)
		return 1
	}

	dataURI, err := cmd.Flags().GetString("data")
	if err != nil {
		log.Errorf("error getting -d/--data flag: %s", err)
		return 1
	}

	loggerURI, err := cmd.Flags().GetString("logger")
	if err != nil {
		log.Errorf("error getting -l/--logger flag: %s", err)
		return 1
	}

	queueURI, err := cmd.Flags().GetString("queue")
	if err != nil {
		log.Errorf("error getting -q/--queue flag: %s", err)
		return 1
	}

	storeURI, err := cmd.Flags().GetString("store")
	if err != nil {
		log.Errorf("error getting -s/--store flag: %s", err)
		return 1
	}

	workers, err := cmd.Flags().GetInt("workers")
	if err != nil {
		log.Errorf("error getting -W/--workers flag: %s", err)
		return 1
	}

	buffer, err := cmd.Flags().GetInt("buffer")
	if err != nil {
		log.Errorf("error getting -B/--buffer flag: %s", err)
		return 1
	}

	jobType := "*"
	if len(args) == 1 {
		jobType = args[0]
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	_, err = je.InitData(dataURI)
	if err != nil {
		log.Errorf("error initializing data: %s", err)
		return 1
	}

	store, err := je.InitStore(storeURI)
	if err != nil {
		log.Errorf("error initializing store: %s", err)
		return 1
	}
	defer store.Close()

	opts := je.BossOptions{
		JobType: jobType,
		Workers: workers,
		Buffer:  buffer,
	}
	boss := je.NewBoss(bind, loggerURI, queueURI, &opts)

	http.Handle("/", boss)
	http.Handle("/metrics", boss.Metrics().Handler())
	log.Infof("je worker %s listening on %s", je.FullVersion(), bind)
	log.Infof("operating on jobs of type: %s", jobType)
	log.Fatal(http.ListenAndServe(bind, nil))

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	log.Infof("shuting down...")
	boss.Shutdown()

	return 0
}
