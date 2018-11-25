package main

import (
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prologic/je"
)

// waitCmd represents the run command
var waitCmd = &cobra.Command{
	Use:     "wait [flags] <id>",
	Aliases: []string{"join"},
	Short:   "Waits for a job to complete",
	Long: `This waits for the given job id to complete before returning and
displaying the job's exit status`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := je.NewClient(uri, nil)

		interval, err := cmd.Flags().GetDuration("interval")
		if err != nil {
			log.Errorf("error getting -i/--interval flag: %s", err)
			os.Exit(1)
		}

		timeout, err := cmd.Flags().GetDuration("timeout")
		if err != nil {
			log.Errorf("error getting -t/--timeout flag: %s", err)
			os.Exit(1)
		}

		id := args[0]

		os.Exit(wait(client, id, interval, timeout))
	},
}

func init() {
	RootCmd.AddCommand(waitCmd)

	waitCmd.Flags().DurationP(
		"interval", "i", 5*time.Second,
		"Poll interval duration",
	)

	waitCmd.Flags().DurationP(
		"timeout", "t", 30*time.Second,
		"Timeout after specified duration",
	)
}

func wait(client *je.Client, id string, interval, timeout time.Duration) int {
	res, err := client.GetJobByID(id)
	if err != nil {
		log.Errorf("error retrieving information for job #%s: %s", id, err)
		return 1
	}
	state := res[0].State
	if state == je.STATE_STOPPED || state == je.STATE_ERRORED {
		fmt.Print(res[0].Status)
		return 0
	}

	t1 := time.NewTicker(interval)
	t2 := time.NewTimer(timeout)

	for {
		select {
		case <-t1.C:
			res, err := client.GetJobByID(id)
			if err != nil {
				log.Errorf("error retrieving information for job #%s: %s", id, err)
				return 1
			}
			state := res[0].State
			if state == je.STATE_STOPPED || state == je.STATE_ERRORED {
				fmt.Print(res[0].Status)
				return 0
			}
		case <-t2.C:
			log.Errorf("timed out waiting for job #%s after %s", id, timeout)
			return 2
		}
	}
}
