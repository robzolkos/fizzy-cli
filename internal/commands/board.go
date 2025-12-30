package commands

import (
	"github.com/robzolkos/fizzy-cli/internal/errors"
	"github.com/spf13/cobra"
)

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Manage boards",
	Long:  "Commands for managing Fizzy boards.",
}

// Board list flags
var boardListPage int
var boardListAll bool

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all boards",
	Long:  "Lists all boards you have access to.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		path := "/boards.json"
		if boardListPage > 0 {
			path += "?page=" + string(rune(boardListPage+'0'))
		}

		resp, err := client.GetWithPagination(path, boardListAll)
		if err != nil {
			exitWithError(err)
		}

		hasNext := resp.LinkNext != ""
		printSuccessWithPagination(resp.Data, hasNext, resp.LinkNext)
	},
}

var boardShowCmd = &cobra.Command{
	Use:   "show BOARD_ID",
	Short: "Show a board",
	Long:  "Shows details of a specific board.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Get("/boards/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Board create flags
var boardCreateName string
var boardCreateAllAccess string
var boardCreateAutoPostponePeriod int

var boardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a board",
	Long:  "Creates a new board.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if boardCreateName == "" {
			exitWithError(newRequiredFlagError("name"))
		}

		body := map[string]interface{}{
			"name": boardCreateName,
		}

		if boardCreateAllAccess != "" {
			body["all_access"] = boardCreateAllAccess == "true"
		}
		if boardCreateAutoPostponePeriod > 0 {
			body["auto_postpone_period"] = boardCreateAutoPostponePeriod
		}

		client := getClient()
		resp, err := client.Post("/boards.json", body)
		if err != nil {
			exitWithError(err)
		}

		// Create returns location header - follow it to get the created resource
		if resp.Location != "" {
			followResp, err := client.FollowLocation(resp.Location)
			if err == nil && followResp != nil {
				printSuccessWithLocation(followResp.Data, resp.Location)
				return
			}
			// If follow fails, just return success with location
			printSuccessWithLocation(nil, resp.Location)
			return
		}

		printSuccess(resp.Data)
	},
}

// Board update flags
var boardUpdateName string
var boardUpdateAllAccess string
var boardUpdateAutoPostponePeriod int

var boardUpdateCmd = &cobra.Command{
	Use:   "update BOARD_ID",
	Short: "Update a board",
	Long:  "Updates an existing board.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		body := make(map[string]interface{})

		if boardUpdateName != "" {
			body["name"] = boardUpdateName
		}
		if boardUpdateAllAccess != "" {
			body["all_access"] = boardUpdateAllAccess == "true"
		}
		if boardUpdateAutoPostponePeriod > 0 {
			body["auto_postpone_period"] = boardUpdateAutoPostponePeriod
		}

		client := getClient()
		resp, err := client.Patch("/boards/"+args[0]+".json", body)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

var boardDeleteCmd = &cobra.Command{
	Use:   "delete BOARD_ID",
	Short: "Delete a board",
	Long:  "Deletes a board.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		_, err := client.Delete("/boards/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"deleted": true,
		})
	},
}

func init() {
	rootCmd.AddCommand(boardCmd)

	// List
	boardListCmd.Flags().IntVar(&boardListPage, "page", 0, "Page number")
	boardListCmd.Flags().BoolVar(&boardListAll, "all", false, "Fetch all pages")
	boardCmd.AddCommand(boardListCmd)

	// Show
	boardCmd.AddCommand(boardShowCmd)

	// Create
	boardCreateCmd.Flags().StringVar(&boardCreateName, "name", "", "Board name (required)")
	boardCreateCmd.Flags().StringVar(&boardCreateAllAccess, "all_access", "", "Allow all team members access (true/false)")
	boardCreateCmd.Flags().IntVar(&boardCreateAutoPostponePeriod, "auto_postpone_period", 0, "Auto postpone period in days")
	boardCmd.AddCommand(boardCreateCmd)

	// Update
	boardUpdateCmd.Flags().StringVar(&boardUpdateName, "name", "", "Board name")
	boardUpdateCmd.Flags().StringVar(&boardUpdateAllAccess, "all_access", "", "Allow all team members access (true/false)")
	boardUpdateCmd.Flags().IntVar(&boardUpdateAutoPostponePeriod, "auto_postpone_period", 0, "Auto postpone period in days")
	boardCmd.AddCommand(boardUpdateCmd)

	// Delete
	boardCmd.AddCommand(boardDeleteCmd)
}

// Helper function for required flag errors
func newRequiredFlagError(flag string) error {
	return errors.NewInvalidArgsError("required flag --" + flag + " not provided")
}
