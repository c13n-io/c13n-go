package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/c13n-io/c13n-backend/app"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version and exit",
	Run: func(_ *cobra.Command, _ []string) {
		version := app.Version()
		commit, chash := app.BuildInfo()

		vString := "c13n version " + version + "\n" +
			"commit=" + commit + "\n" +
			"commit_hash=" + chash + "\n"

		fmt.Println(vString)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
