package main

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prologic/je"
)

// searchCmd represents the run command
var searchCmd = &cobra.Command{
	Use:     "search [flags] [<id>]",
	Aliases: []string{"find"},
	Short:   "Search for matching jobs",
	Long: `This searches for matching jobs given a criteria.
Jobs can be searched by id, name, one or more arguments or by state.
If no jobs are found matching the criteria, an empy list is returned.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		uri := viper.GetString("uri")
		client := je.NewClient(uri, nil)

		var id string

		if len(args) == 1 {
			id = args[0]
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Errorf("error getting -n/--name flag: %s", err)
			os.Exit(1)
		}

		state, err := cmd.Flags().GetString("state")
		if err != nil {
			log.Errorf("error getting -s/--state flag: %s", err)
			os.Exit(1)
		}

		os.Exit(search(client, id, name, state))
	},
}

func init() {
	RootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringP(
		"name", "n", "",
		"Search for jobs by name",
	)

	searchCmd.Flags().StringP(
		"state", "s", "",
		"Search for jobs by state",
	)
}

func search(c *je.Client, id, name, state string) int {
	res, err := c.Search(&je.SearchOptions{
		Filter: &je.SearchFilter{
			ID:    id,
			Name:  name,
			State: state,
		},
	})

	if err != nil {
		log.Errorf("error searching for active jobs: %s", err)
		return 1
	}

	out, err := json.Marshal(res)
	if err != nil {
		log.Errorf("error encoding job results: %s", err)
		return 1
	}

	fmt.Print(string(out))

	return 0
}
