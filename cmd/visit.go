package cmd

import (
	"gotohellava/cmd/ava"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	section = "#section-2"
)

var (
	username string
	password string
)

var visit = &cobra.Command{
	Use:   "visit '[url]' -u '[username]' -p '[password]'",
	Short: "Visit the course page clicking in all links and reloading until not finding any new link to click.",
	Long: `
visit 'https://ava.ufms.br/course/view.php?id=xxxx' -u 'username' -p 'password'
Visit the course page clicking in all links and reloading until not finding any new link to click.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Visiting the course page")
		url := args[0] + section
		ava.Visit(url, username, password)
	},
}

func init() {
	visit.PersistentFlags().StringVarP(&username, "username", "u", "", "Username to login")
	visit.PersistentFlags().StringVarP(&password, "password", "p", "", "Password to login")
	rootCmd.AddCommand(visit)
}
