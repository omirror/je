package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"git.mills.io/prologic/je/client"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run [flags] <name> [--] [args]",
	Aliases: []string{"exec"},
	Short:   "Runs the given job type by name",
	Long: `This runs the job given by the provided name argument and waits
for it to complete before returning and printing the result of the job and
its output log. Arguments to the job can be provided but if those arguments
are themselves command-line options to an execute use -- [args].

Input can also be provided to the job by using the -i/--interactive option
to pass stadard input to the job.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")

		client := client.NewClient(uri, nil)

		if interactive {
			os.Exit(run(client, args[0], args[1:], os.Stdin, raw))
		} else {
			os.Exit(run(client, args[0], args[1:], nil, raw))
		}
	},
}

var (
	raw         bool
	interactive bool
)

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVarP(&raw,
		"raw", "r", false,
		"Output job response in raw form (output only)",
	)

	runCmd.Flags().BoolVarP(&interactive,
		"interactive", "i", false,
		"Pass stdin as input to job",
	)
}

func run(client *client.Client, name string, args []string, input io.Reader, raw bool) int {
	res, err := client.Create(name, args, input, true)
	if err != nil {
		log.Errorf("error running job %s: %s", name, err)
		return 1
	}

	if raw {
		return output(client, fmt.Sprintf("%d", res[0].ID), false)
	}

	out, err := json.Marshal(res)
	if err != nil {
		log.Errorf("error encoding job result: %s", err)
		return 1
	}
	fmt.Printf(string(out))

	return 0
}
