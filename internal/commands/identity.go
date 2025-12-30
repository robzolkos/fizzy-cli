package commands

import (
	"github.com/spf13/cobra"
)

var identityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Manage identity",
	Long:  "Commands for viewing your identity and accessible accounts.",
}

var identityShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show your identity and accessible accounts",
	Long:  "Displays your user identity and all accounts you have access to.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuth(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		// Identity endpoint doesn't use account prefix
		resp, err := client.Get(cfg.APIURL + "/my/identity.json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

func init() {
	rootCmd.AddCommand(identityCmd)
	identityCmd.AddCommand(identityShowCmd)
}
