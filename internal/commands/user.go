package commands

import (
	"strconv"

	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  "Commands for viewing users in your account.",
}

// User list flags
var userListPage int
var userListAll bool

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Long:  "Lists all users in your account.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		path := "/users.json"
		if userListPage > 0 {
			path += "?page=" + strconv.Itoa(userListPage)
		}

		resp, err := client.GetWithPagination(path, userListAll)
		if err != nil {
			exitWithError(err)
		}

		hasNext := resp.LinkNext != ""
		printSuccessWithPagination(resp.Data, hasNext, resp.LinkNext)
	},
}

var userShowCmd = &cobra.Command{
	Use:   "show USER_ID",
	Short: "Show a user",
	Long:  "Shows details of a specific user.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Get("/users/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

func init() {
	rootCmd.AddCommand(userCmd)

	// List
	userListCmd.Flags().IntVar(&userListPage, "page", 0, "Page number")
	userListCmd.Flags().BoolVar(&userListAll, "all", false, "Fetch all pages")
	userCmd.AddCommand(userListCmd)

	// Show
	userCmd.AddCommand(userShowCmd)
}
