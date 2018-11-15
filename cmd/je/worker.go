package main

import (
	"net/http"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/prologic/je"
	"github.com/prologic/je/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var workerCmd = &cobra.Command{
	Use:     "worker",
	Aliases: []string{"agent", "minion"},
	Short:   "Run the worker agent",
	Long: `This runs the worker daemon that listens for new jobs on a queue,
processes them concurrently and sends results back to the logger.`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(runWorker(cmd, args...))
	},
}

func init() {
	RootCmd.AddCommand(workerCmd)

	workerCmd.Flags().StringP(
		"bind", "b", ":32436",
		"[int]:<port> to bind to",
	)

	workerCmd.Flags().StringP(
		"logger", "l", "je+http://localhost:10706/",
		"message queue to use for subscribing to jobs and publishing results",
	)

	workerCmd.Flags().StringP(
		"queue", "q", "msgbus+http://localhost:58050/",
		"message queue to use for subscribing to jobs and publishing results",
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

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	queue, err := je.InitQueue(queueURI)
	if err != nil {
		log.Errorf("error initializing store: %s", err)
		return 1
	}
	defer queue.Close()

	opts := worker.Options{}
	boss := worker.NewBoss(bind, loggerURI, queueURI, &opts)

	http.Handle("/", boss)
	http.Handle("/metrics", boss.Metrics().Handler())
	log.Infof("je worker %s listening on %s", je.FullVersion(), bind)
	log.Fatal(http.ListenAndServe(bind, nil))

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	log.Infof("shuting down...")
	boss.Shutdown()

	return 0
}
