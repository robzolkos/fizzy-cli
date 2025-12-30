package commands

import (
	"strconv"

	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
	Long:  "Commands for viewing tags in your account.",
}

// Tag list flags
var tagListPage int
var tagListAll bool

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags",
	Long:  "Lists all tags in your account.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		path := "/tags.json"
		if tagListPage > 0 {
			path += "?page=" + strconv.Itoa(tagListPage)
		}

		resp, err := client.GetWithPagination(path, tagListAll)
		if err != nil {
			exitWithError(err)
		}

		hasNext := resp.LinkNext != ""
		printSuccessWithPagination(resp.Data, hasNext, resp.LinkNext)
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)

	// List
	tagListCmd.Flags().IntVar(&tagListPage, "page", 0, "Page number")
	tagListCmd.Flags().BoolVar(&tagListAll, "all", false, "Fetch all pages")
	tagCmd.AddCommand(tagListCmd)
}
