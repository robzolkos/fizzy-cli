package commands

import (
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Manage cards",
	Long:  "Commands for managing Fizzy cards.",
}

// Card list flags
var cardListBoard string
var cardListTag string
var cardListStatus string
var cardListAssignee string
var cardListPage int
var cardListAll bool

var cardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cards",
	Long:  "Lists cards with optional filters.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		boardID := defaultBoard(cardListBoard)

		client := getClient()
		path := "/cards.json"
		params := []string{}

		if boardID != "" {
			params = append(params, "board_id="+boardID)
		}
		if cardListTag != "" {
			params = append(params, "tag_id="+cardListTag)
		}
		if cardListStatus != "" {
			params = append(params, "status="+cardListStatus)
		}
		if cardListAssignee != "" {
			params = append(params, "assignee_id="+cardListAssignee)
		}
		if cardListPage > 0 {
			params = append(params, "page="+strconv.Itoa(cardListPage))
		}

		if len(params) > 0 {
			path += "?"
			for i, p := range params {
				if i > 0 {
					path += "&"
				}
				path += p
			}
		}

		resp, err := client.GetWithPagination(path, cardListAll)
		if err != nil {
			exitWithError(err)
		}

		hasNext := resp.LinkNext != ""
		printSuccessWithPagination(resp.Data, hasNext, resp.LinkNext)
	},
}

var cardShowCmd = &cobra.Command{
	Use:   "show CARD_NUMBER",
	Short: "Show a card",
	Long:  "Shows details of a specific card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Get("/cards/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Card create flags
var cardCreateBoard string
var cardCreateTitle string
var cardCreateDescription string
var cardCreateDescriptionFile string
var cardCreateTagIDs string
var cardCreateImage string
var cardCreateCreatedAt string

var cardCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a card",
	Long:  "Creates a new card in a board.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		boardID, err := requireBoard(cardCreateBoard)
		if err != nil {
			exitWithError(err)
		}
		if cardCreateTitle == "" {
			exitWithError(newRequiredFlagError("title"))
		}

		body := map[string]interface{}{
			"board_id": boardID,
			"title":    cardCreateTitle,
		}

		// Handle description
		if cardCreateDescriptionFile != "" {
			content, err := os.ReadFile(cardCreateDescriptionFile)
			if err != nil {
				exitWithError(err)
			}
			body["description"] = string(content)
		} else if cardCreateDescription != "" {
			body["description"] = cardCreateDescription
		}

		if cardCreateTagIDs != "" {
			body["tag_ids"] = cardCreateTagIDs
		}
		if cardCreateImage != "" {
			body["image"] = cardCreateImage
		}
		if cardCreateCreatedAt != "" {
			body["created_at"] = cardCreateCreatedAt
		}

		client := getClient()
		resp, err := client.Post("/cards.json", body)
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

// Card update flags
var cardUpdateTitle string
var cardUpdateDescription string
var cardUpdateDescriptionFile string
var cardUpdateCreatedAt string

var cardUpdateCmd = &cobra.Command{
	Use:   "update CARD_NUMBER",
	Short: "Update a card",
	Long:  "Updates an existing card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		body := make(map[string]interface{})

		if cardUpdateTitle != "" {
			body["title"] = cardUpdateTitle
		}
		if cardUpdateDescriptionFile != "" {
			content, err := os.ReadFile(cardUpdateDescriptionFile)
			if err != nil {
				exitWithError(err)
			}
			body["description"] = string(content)
		} else if cardUpdateDescription != "" {
			body["description"] = cardUpdateDescription
		}
		if cardUpdateCreatedAt != "" {
			body["created_at"] = cardUpdateCreatedAt
		}

		client := getClient()
		resp, err := client.Patch("/cards/"+args[0]+".json", body)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

var cardDeleteCmd = &cobra.Command{
	Use:   "delete CARD_NUMBER",
	Short: "Delete a card",
	Long:  "Deletes a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		_, err := client.Delete("/cards/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"deleted": true,
		})
	},
}

var cardCloseCmd = &cobra.Command{
	Use:   "close CARD_NUMBER",
	Short: "Close a card",
	Long:  "Closes a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Post("/cards/"+args[0]+"/closure.json", nil)
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

var cardReopenCmd = &cobra.Command{
	Use:   "reopen CARD_NUMBER",
	Short: "Reopen a card",
	Long:  "Reopens a closed card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Delete("/cards/" + args[0] + "/closure.json")
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

var cardPostponeCmd = &cobra.Command{
	Use:   "postpone CARD_NUMBER",
	Short: "Postpone a card",
	Long:  "Moves a card to 'Not Now'.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Post("/cards/"+args[0]+"/not_now.json", nil)
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

// Card column flags
var cardColumnColumn string

var cardColumnCmd = &cobra.Command{
	Use:   "column CARD_NUMBER",
	Short: "Move card to column",
	Long:  "Moves a card to a specific column.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if cardColumnColumn == "" {
			exitWithError(newRequiredFlagError("column"))
		}

		body := map[string]interface{}{
			"column_id": cardColumnColumn,
		}

		client := getClient()
		resp, err := client.Post("/cards/"+args[0]+"/triage.json", body)
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

var cardUntriageCmd = &cobra.Command{
	Use:   "untriage CARD_NUMBER",
	Short: "Send card back to triage",
	Long:  "Removes a card from its column and sends it back to triage.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Delete("/cards/" + args[0] + "/triage.json")
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{
				"untriaged": true,
			})
		}
	},
}

// Card assign flags
var cardAssignUser string

var cardAssignCmd = &cobra.Command{
	Use:   "assign CARD_NUMBER",
	Short: "Toggle assignment on a card",
	Long:  "Toggles a user's assignment on a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if cardAssignUser == "" {
			exitWithError(newRequiredFlagError("user"))
		}

		body := map[string]interface{}{
			"assignee_id": cardAssignUser,
		}

		client := getClient()
		resp, err := client.Post("/cards/"+args[0]+"/assignments.json", body)
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

// Card tag flags
var cardTagTag string

var cardTagCmd = &cobra.Command{
	Use:   "tag CARD_NUMBER",
	Short: "Toggle tag on a card",
	Long:  "Toggles a tag on a card. Creates the tag if it doesn't exist.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if cardTagTag == "" {
			exitWithError(newRequiredFlagError("tag"))
		}

		body := map[string]interface{}{
			"tag_title": cardTagTag,
		}

		client := getClient()
		resp, err := client.Post("/cards/"+args[0]+"/taggings.json", body)
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

var cardWatchCmd = &cobra.Command{
	Use:   "watch CARD_NUMBER",
	Short: "Watch a card",
	Long:  "Subscribes to notifications for a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Post("/cards/"+args[0]+"/watch.json", nil)
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

var cardUnwatchCmd = &cobra.Command{
	Use:   "unwatch CARD_NUMBER",
	Short: "Unwatch a card",
	Long:  "Unsubscribes from notifications for a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Delete("/cards/" + args[0] + "/watch.json")
		if err != nil {
			exitWithError(err)
		}

		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

func init() {
	rootCmd.AddCommand(cardCmd)

	// List
	cardListCmd.Flags().StringVar(&cardListBoard, "board", "", "Filter by board ID")
	cardListCmd.Flags().StringVar(&cardListTag, "tag", "", "Filter by tag ID")
	cardListCmd.Flags().StringVar(&cardListStatus, "status", "", "Filter by status")
	cardListCmd.Flags().StringVar(&cardListAssignee, "assignee", "", "Filter by assignee ID")
	cardListCmd.Flags().IntVar(&cardListPage, "page", 0, "Page number")
	cardListCmd.Flags().BoolVar(&cardListAll, "all", false, "Fetch all pages")
	cardCmd.AddCommand(cardListCmd)

	// Show
	cardCmd.AddCommand(cardShowCmd)

	// Create
	cardCreateCmd.Flags().StringVar(&cardCreateBoard, "board", "", "Board ID (required)")
	cardCreateCmd.Flags().StringVar(&cardCreateTitle, "title", "", "Card title (required)")
	cardCreateCmd.Flags().StringVar(&cardCreateDescription, "description", "", "Card description (HTML)")
	cardCreateCmd.Flags().StringVar(&cardCreateDescriptionFile, "description_file", "", "Read description from file")
	cardCreateCmd.Flags().StringVar(&cardCreateTagIDs, "tag-ids", "", "Comma-separated tag IDs")
	cardCreateCmd.Flags().StringVar(&cardCreateImage, "image", "", "Header image signed ID")
	cardCreateCmd.Flags().StringVar(&cardCreateCreatedAt, "created-at", "", "Custom created_at timestamp")
	cardCmd.AddCommand(cardCreateCmd)

	// Update
	cardUpdateCmd.Flags().StringVar(&cardUpdateTitle, "title", "", "Card title")
	cardUpdateCmd.Flags().StringVar(&cardUpdateDescription, "description", "", "Card description (HTML)")
	cardUpdateCmd.Flags().StringVar(&cardUpdateDescriptionFile, "description_file", "", "Read description from file")
	cardUpdateCmd.Flags().StringVar(&cardUpdateCreatedAt, "created-at", "", "Custom created_at timestamp")
	cardCmd.AddCommand(cardUpdateCmd)

	// Delete
	cardCmd.AddCommand(cardDeleteCmd)

	// Actions
	cardCmd.AddCommand(cardCloseCmd)
	cardCmd.AddCommand(cardReopenCmd)
	cardCmd.AddCommand(cardPostponeCmd)

	// Column
	cardColumnCmd.Flags().StringVar(&cardColumnColumn, "column", "", "Column ID (required)")
	cardCmd.AddCommand(cardColumnCmd)

	// Untriage
	cardCmd.AddCommand(cardUntriageCmd)

	// Assign
	cardAssignCmd.Flags().StringVar(&cardAssignUser, "user", "", "User ID (required)")
	cardCmd.AddCommand(cardAssignCmd)

	// Tag
	cardTagCmd.Flags().StringVar(&cardTagTag, "tag", "", "Tag name (required)")
	cardCmd.AddCommand(cardTagCmd)

	// Watch/Unwatch
	cardCmd.AddCommand(cardWatchCmd)
	cardCmd.AddCommand(cardUnwatchCmd)
}
