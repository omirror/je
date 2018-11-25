package main

import (
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/prologic/je"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var daemonCmd = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"server", "api"},
	Short:   "Run the server daemon",
	Long: `This runs the server daemon that provides the API for workers
and clients.`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(daemon(cmd, args...))
	},
}

func init() {
	RootCmd.AddCommand(daemonCmd)

	daemonCmd.Flags().StringP(
		"bind", "b", ":8000",
		"[int]:<port> to bind to",
	)

	daemonCmd.Flags().StringP(
		"data", "d", "file://data",
		"where to persist job data",
	)

	daemonCmd.Flags().StringP(
		"queue", "q", "msgbus+http://localhost:8001/",
		"message queue to use for publishing jobs",
	)

	daemonCmd.Flags().StringP(
		"store", "s", "memory://",
		"database store to persist job metadata",
	)
}

func daemon(cmd *cobra.Command, args ...string) int {
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

	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	metrics := je.InitMetrics("je")

	data, err := je.InitData(dataURI)
	if err != nil {
		log.Errorf("error initializing data: %s", err)
		return 1
	}

	queue, err := je.InitQueue(queueURI)
	if err != nil {
		log.Errorf("error initializing store: %s", err)
		return 1
	}
	defer queue.Close()

	store, err := je.InitStore(storeURI)
	if err != nil {
		log.Errorf("error initializing store: %s", err)
		return 1
	}
	defer store.Close()

	server := je.NewServer(bind, data, queue, store)
	server.AddRoute("GET", "/metrics", metrics.Handler())

	log.Infof("je %s listening on %s", je.FullVersion(), bind)
	server.ListenAndServe()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	log.Infof("shuting down...")
	server.Shutdown()

	return 0
}
