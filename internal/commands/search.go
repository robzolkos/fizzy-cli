package commands

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Search flags
var searchBoard string
var searchTag string
var searchAssignee string
var searchIndexedBy string
var searchSort string
var searchPage int
var searchAll bool

var searchCmd = &cobra.Command{
	Use:   "search QUERY",
	Short: "Search cards",
	Long:  "Searches cards by text. Multiple words are treated as separate terms (AND).",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		query := strings.Join(args, " ")

		client := getClient()
		path := "/cards.json"

		var params []string

		// Add search terms
		for _, term := range strings.Fields(query) {
			params = append(params, "terms[]="+term)
		}

		// Add optional filters
		boardID := defaultBoard(searchBoard)
		if boardID != "" {
			params = append(params, "board_ids[]="+boardID)
		}
		if searchTag != "" {
			params = append(params, "tag_ids[]="+searchTag)
		}
		if searchAssignee != "" {
			params = append(params, "assignee_ids[]="+searchAssignee)
		}
		if searchIndexedBy != "" {
			params = append(params, "indexed_by="+searchIndexedBy)
		}
		if searchSort != "" {
			params = append(params, "sorted_by="+searchSort)
		}
		if searchPage > 0 {
			params = append(params, "page="+strconv.Itoa(searchPage))
		}

		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}

		resp, err := client.GetWithPagination(path, searchAll)
		if err != nil {
			exitWithError(err)
		}

		hasNext := resp.LinkNext != ""
		printSuccessWithPagination(resp.Data, hasNext, resp.LinkNext)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVar(&searchBoard, "board", "", "Filter by board ID")
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "Filter by tag ID")
	searchCmd.Flags().StringVar(&searchAssignee, "assignee", "", "Filter by assignee ID")
	searchCmd.Flags().StringVar(&searchIndexedBy, "indexed-by", "", "Filter by status (all, closed, not_now, golden)")
	searchCmd.Flags().StringVar(&searchSort, "sort", "", "Sort order: newest, oldest, or latest (default)")
	searchCmd.Flags().IntVar(&searchPage, "page", 0, "Page number")
	searchCmd.Flags().BoolVar(&searchAll, "all", false, "Fetch all pages")
}
