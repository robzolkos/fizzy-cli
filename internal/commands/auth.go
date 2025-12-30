package commands

import (
	"github.com/robzolkos/fizzy-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Commands for managing API authentication.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login TOKEN",
	Short: "Save API token to config file",
	Long:  "Saves the provided API token to ~/.config/fizzy/config.yaml for future use.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := args[0]

		// Load existing config or create new
		globalCfg := config.LoadGlobal()
		globalCfg.Token = token

		if err := globalCfg.Save(); err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"authenticated": true,
			"message":       "Token saved to config file",
		})
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved credentials",
	Long:  "Removes the config file containing saved credentials.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.Delete(); err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"authenticated": false,
			"message":       "Logged out successfully",
		})
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  "Shows whether you are currently authenticated.",
	Run: func(cmd *cobra.Command, args []string) {
		effectiveCfg := cfg
		if effectiveCfg == nil {
			effectiveCfg = config.Load()
		}

		status := map[string]interface{}{
			"authenticated": effectiveCfg.Token != "",
		}

		if effectiveCfg.Token != "" {
			status["token_configured"] = true
			if effectiveCfg.Account != "" {
				status["account"] = effectiveCfg.Account
			}
			if effectiveCfg.APIURL != "" && effectiveCfg.APIURL != config.DefaultAPIURL {
				status["api_url"] = effectiveCfg.APIURL
			}
		}

		printSuccess(status)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
}
