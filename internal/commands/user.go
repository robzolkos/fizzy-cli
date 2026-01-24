package commands

import (
	"fmt"
	"strconv"

	"github.com/robzolkos/fizzy-cli/internal/response"
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

		// Build summary
		count := 0
		if arr, ok := resp.Data.([]interface{}); ok {
			count = len(arr)
		}
		summary := fmt.Sprintf("%d users", count)
		if userListAll {
			summary += " (all)"
		} else if userListPage > 0 {
			summary += fmt.Sprintf(" (page %d)", userListPage)
		}

		// Build breadcrumbs
		breadcrumbs := []response.Breadcrumb{
			breadcrumb("show", "fizzy user show <id>", "View user details"),
			breadcrumb("assign", "fizzy card assign <number> --user <user_id>", "Assign user to card"),
		}

		hasNext := resp.LinkNext != ""
		if hasNext {
			nextPage := userListPage + 1
			if nextPage == 0 {
				nextPage = 2
			}
			breadcrumbs = append(breadcrumbs, breadcrumb("next", fmt.Sprintf("fizzy user list --page %d", nextPage), "Next page"))
		}

		printSuccessWithPaginationAndBreadcrumbs(resp.Data, hasNext, resp.LinkNext, summary, breadcrumbs)
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

		userID := args[0]

		client := getClient()
		resp, err := client.Get("/users/" + userID + ".json")
		if err != nil {
			exitWithError(err)
		}

		// Build breadcrumbs
		breadcrumbs := []response.Breadcrumb{
			breadcrumb("people", "fizzy user list", "List users"),
			breadcrumb("assign", fmt.Sprintf("fizzy card assign <number> --user %s", userID), "Assign to card"),
		}

		printSuccessWithBreadcrumbs(resp.Data, "", breadcrumbs)
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
