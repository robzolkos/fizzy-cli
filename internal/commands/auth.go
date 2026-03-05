package commands

import (
	"fmt"
	"os"

	"github.com/basecamp/cli/output"
	"github.com/basecamp/fizzy-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Commands for managing API authentication.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login TOKEN",
	Short: "Save API token",
	Long:  "Saves the provided API token to the system keyring (or fallback file).",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		token := args[0]

		if creds != nil {
			if err := credsSaveToken(token); err != nil {
				return &output.Error{Code: output.CodeAPI, Message: err.Error()}
			}

			// Clear token from global YAML config if present (migration)
			globalCfg := config.LoadGlobal()
			if globalCfg.Token != "" {
				globalCfg.Token = ""
				_ = globalCfg.Save()
			}
		} else {
			// Fallback: save to config file (test mode or credstore unavailable)
			globalCfg := config.LoadGlobal()
			globalCfg.Token = token
			if err := globalCfg.Save(); err != nil {
				return &output.Error{Code: output.CodeAPI, Message: err.Error()}
			}
		}

		// Build breadcrumbs
		breadcrumbs := []Breadcrumb{
			breadcrumb("status", "fizzy auth status", "Check auth status"),
			breadcrumb("identity", "fizzy identity show", "View identity"),
			breadcrumb("boards", "fizzy board list", "List boards"),
		}

		result := map[string]any{
			"authenticated": true,
			"message":       "Token saved",
		}
		if creds != nil {
			if creds.UsingKeyring() {
				result["storage"] = "keyring"
			} else {
				result["storage"] = "file"
			}
		}

		printMutation(result, "", breadcrumbs)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove saved credentials",
	Long:  "Removes saved credentials from keyring and config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Delete from credstore
		if creds != nil {
			if err := creds.Delete("token"); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not remove token from credential store: %v\n", err)
			}
		}

		// Also clear any token in config file
		if err := config.Delete(); err != nil {
			return &output.Error{Code: output.CodeAPI, Message: err.Error()}
		}

		// Build breadcrumbs
		breadcrumbs := []Breadcrumb{
			breadcrumb("login", "fizzy auth login <token>", "Log in again"),
		}

		printMutation(map[string]any{
			"authenticated": false,
			"message":       "Logged out successfully",
		}, "", breadcrumbs)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Long:  "Shows whether you are currently authenticated.",
	RunE: func(cmd *cobra.Command, args []string) error {
		effectiveCfg := cfg
		if effectiveCfg == nil {
			effectiveCfg = config.Load()
		}

		status := map[string]any{
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

		if creds != nil {
			status["using_keyring"] = creds.UsingKeyring()
			if w := creds.FallbackWarning(); w != "" {
				status["credential_warning"] = w
			}
		}

		// Build breadcrumbs
		breadcrumbs := []Breadcrumb{
			breadcrumb("identity", "fizzy identity show", "View identity"),
			breadcrumb("logout", "fizzy auth logout", "Log out"),
		}

		printDetail(status, "", breadcrumbs)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
}
