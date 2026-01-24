package commands

import (
	"github.com/robzolkos/fizzy-cli/internal/response"
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

		// Build breadcrumbs
		breadcrumbs := []response.Breadcrumb{
			breadcrumb("status", "fizzy auth status", "Auth status"),
		}

		printSuccessWithBreadcrumbs(resp.Data, "", breadcrumbs)
	},
}

func init() {
	rootCmd.AddCommand(identityCmd)
	identityCmd.AddCommand(identityShowCmd)
}
