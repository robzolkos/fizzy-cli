package commands

import (
	"github.com/robzolkos/fizzy-cli/internal/errors"
	"github.com/spf13/cobra"
)

var columnCmd = &cobra.Command{
	Use:   "column",
	Short: "Manage columns",
	Long:  "Commands for managing board columns.",
}

// Column list flags
var columnListBoard string

var columnListCmd = &cobra.Command{
	Use:   "list",
	Short: "List columns for a board",
	Long:  "Lists all columns for a specific board.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		boardID, err := requireBoard(columnListBoard)
		if err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Get("/boards/" + boardID + "/columns.json")
		if err != nil {
			exitWithError(err)
		}

		data, ok := resp.Data.([]interface{})
		if !ok {
			printSuccess(resp.Data)
			return
		}

		cols := make([]interface{}, 0, len(data)+3)
		cols = append(cols, pseudoColumnObject(pseudoColumnNotYet), pseudoColumnObject(pseudoColumnMaybe))
		cols = append(cols, data...)
		cols = append(cols, pseudoColumnObject(pseudoColumnDone))

		printSuccess(cols)
	},
}

// Column show flags
var columnShowBoard string

var columnShowCmd = &cobra.Command{
	Use:   "show COLUMN_ID",
	Short: "Show a column",
	Long:  "Shows details of a specific column.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if pseudo, ok := parsePseudoColumnID(args[0]); ok {
			printSuccess(pseudoColumnObject(pseudo))
			return
		}

		boardID, err := requireBoard(columnShowBoard)
		if err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Get("/boards/" + boardID + "/columns/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Column create flags
var columnCreateBoard string
var columnCreateName string
var columnCreateColor string

var columnCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a column",
	Long:  "Creates a new column in a board.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		boardID, err := requireBoard(columnCreateBoard)
		if err != nil {
			exitWithError(err)
		}
		if columnCreateName == "" {
			exitWithError(newRequiredFlagError("name"))
		}

		body := map[string]interface{}{
			"name": columnCreateName,
		}
		if columnCreateColor != "" {
			body["color"] = columnCreateColor
		}

		client := getClient()
		resp, err := client.Post("/boards/"+boardID+"/columns.json", body)
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
			printSuccessWithLocation(nil, resp.Location)
			return
		}

		printSuccess(resp.Data)
	},
}

// Column update flags
var columnUpdateBoard string
var columnUpdateName string
var columnUpdateColor string

var columnUpdateCmd = &cobra.Command{
	Use:   "update COLUMN_ID",
	Short: "Update a column",
	Long:  "Updates an existing column.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if _, ok := parsePseudoColumnID(args[0]); ok {
			exitWithError(errors.NewInvalidArgsError("cannot update pseudo columns (Not Yet, Maybe?, Done)"))
		}

		boardID, err := requireBoard(columnUpdateBoard)
		if err != nil {
			exitWithError(err)
		}

		body := make(map[string]interface{})
		if columnUpdateName != "" {
			body["name"] = columnUpdateName
		}
		if columnUpdateColor != "" {
			body["color"] = columnUpdateColor
		}

		client := getClient()
		resp, err := client.Patch("/boards/"+boardID+"/columns/"+args[0]+".json", body)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Column delete flags
var columnDeleteBoard string

var columnDeleteCmd = &cobra.Command{
	Use:   "delete COLUMN_ID",
	Short: "Delete a column",
	Long:  "Deletes a column from a board.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if _, ok := parsePseudoColumnID(args[0]); ok {
			exitWithError(errors.NewInvalidArgsError("cannot delete pseudo columns (Not Yet, Maybe?, Done)"))
		}

		boardID, err := requireBoard(columnDeleteBoard)
		if err != nil {
			exitWithError(err)
		}

		client := getClient()
		_, err = client.Delete("/boards/" + boardID + "/columns/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"deleted": true,
		})
	},
}

func init() {
	rootCmd.AddCommand(columnCmd)

	// List
	columnListCmd.Flags().StringVar(&columnListBoard, "board", "", "Board ID (required)")
	columnCmd.AddCommand(columnListCmd)

	// Show
	columnShowCmd.Flags().StringVar(&columnShowBoard, "board", "", "Board ID (required)")
	columnCmd.AddCommand(columnShowCmd)

	// Create
	columnCreateCmd.Flags().StringVar(&columnCreateBoard, "board", "", "Board ID (required)")
	columnCreateCmd.Flags().StringVar(&columnCreateName, "name", "", "Column name (required)")
	columnCreateCmd.Flags().StringVar(&columnCreateColor, "color", "", "Column color")
	columnCmd.AddCommand(columnCreateCmd)

	// Update
	columnUpdateCmd.Flags().StringVar(&columnUpdateBoard, "board", "", "Board ID (required)")
	columnUpdateCmd.Flags().StringVar(&columnUpdateName, "name", "", "Column name")
	columnUpdateCmd.Flags().StringVar(&columnUpdateColor, "color", "", "Column color")
	columnCmd.AddCommand(columnUpdateCmd)

	// Delete
	columnDeleteCmd.Flags().StringVar(&columnDeleteBoard, "board", "", "Board ID (required)")
	columnCmd.AddCommand(columnDeleteCmd)
}
