package commands

import (
	"strconv"

	"github.com/spf13/cobra"
)

var notificationCmd = &cobra.Command{
	Use:   "notification",
	Short: "Manage notifications",
	Long:  "Commands for managing your notifications.",
}

// Notification list flags
var notificationListPage int
var notificationListAll bool

var notificationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List notifications",
	Long:  "Lists your notifications.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		path := "/notifications.json"
		if notificationListPage > 0 {
			path += "?page=" + strconv.Itoa(notificationListPage)
		}

		resp, err := client.GetWithPagination(path, notificationListAll)
		if err != nil {
			exitWithError(err)
		}

		hasNext := resp.LinkNext != ""
		printSuccessWithPagination(resp.Data, hasNext, resp.LinkNext)
	},
}

var notificationReadCmd = &cobra.Command{
	Use:   "read NOTIFICATION_ID",
	Short: "Mark notification as read",
	Long:  "Marks a notification as read.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Post("/notifications/"+args[0]+"/read.json", nil)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

var notificationUnreadCmd = &cobra.Command{
	Use:   "unread NOTIFICATION_ID",
	Short: "Mark notification as unread",
	Long:  "Marks a notification as unread.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Post("/notifications/"+args[0]+"/unread.json", nil)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

var notificationReadAllCmd = &cobra.Command{
	Use:   "read-all",
	Short: "Mark all notifications as read",
	Long:  "Marks all notifications as read.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Post("/notifications/bulk_reading.json", nil)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

func init() {
	rootCmd.AddCommand(notificationCmd)

	// List
	notificationListCmd.Flags().IntVar(&notificationListPage, "page", 0, "Page number")
	notificationListCmd.Flags().BoolVar(&notificationListAll, "all", false, "Fetch all pages")
	notificationCmd.AddCommand(notificationListCmd)

	// Read/Unread
	notificationCmd.AddCommand(notificationReadCmd)
	notificationCmd.AddCommand(notificationUnreadCmd)
	notificationCmd.AddCommand(notificationReadAllCmd)
}
