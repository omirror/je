package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je"
	"git.mills.io/prologic/je/client"
)

// psCmd represents the run command
var psCmd = &cobra.Command{
	Use:     "ps [flags]",
	Aliases: []string{"ls", "list"},
	Short:   "List all running jobs",
	Long:    `This list all actively running jobs in a ps-like output`,
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := client.NewClient(uri, nil)

		os.Exit(ps(client))
	},
}

func init() {
	RootCmd.AddCommand(psCmd)
}

func ps(c *client.Client) int {
	res, err := c.Search(&client.SearchOptions{
		Filter: &client.SearchFilter{
			State: je.STATE_RUNNING,
		},
	})

	if err != nil {
		log.Errorf("error searching for active jobs: %s", err)
		return 1
	}

	if res == nil {
		return 0
	}

	w := tabwriter.NewWriter(os.Stdout, 10, 4, 8, ' ', 0)
	w.Write([]byte("ID\tNAME\tCREATED\tSTATE\n"))

	var (
		created string
		d       time.Duration
	)

	for _, job := range res {
		d = time.Since(job.CreatedAt)
		if (d / time.Second) < 1.0 {
			created = "just now"
		} else {
			created = fmt.Sprintf("%s ago", d.Truncate(time.Second))
		}

		w.Write(
			[]byte(
				fmt.Sprintf(
					"%d\t%s\t%s\t%s\n",
					job.ID,
					job.Name,
					created,
					job.State.String(),
				),
			),
		)
	}
	w.Flush()

	return 0
}
