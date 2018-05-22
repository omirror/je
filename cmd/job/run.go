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
		raw := viper.GetBool("raw")
		interactive := viper.GetBool("interactive")

		client := client.NewClient(uri, nil)

		if interactive {
			run(client, args[0], args[1:], os.Stdin, raw)
		} else {
			run(client, args[0], args[1:], nil, raw)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolP(
		"raw", "r", false,
		"Output job response in raw form (output only)",
	)
	viper.BindPFlag("raw", runCmd.Flags().Lookup("raw"))
	viper.SetDefault("raw", false)

	runCmd.Flags().BoolP(
		"interactive", "i", false,
		"Pass stdin as input to job",
	)
	viper.BindPFlag("interactive", runCmd.Flags().Lookup("interactive"))
	viper.SetDefault("interactive", false)

}

func run(client *client.Client, name string, args []string, input io.Reader, raw bool) {
	res, err := client.Run(name, args, input)
	if err != nil {
		log.Errorf("error running job %s: %s", name, err)
		return
	}

	if raw {
		fmt.Print(res[0].Output)
	} else {
		out, err := json.Marshal(res)
		if err != nil {
			log.Errorf("error encoding job result: %s", err)
			return
		}
		fmt.Printf(string(out))
	}
}
