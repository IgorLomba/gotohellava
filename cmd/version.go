package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Get AVA-CLI Version",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version %s", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
