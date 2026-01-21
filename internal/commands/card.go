package commands

import (
	"os"
	"strconv"
	"strings"

	"github.com/robzolkos/fizzy-cli/internal/errors"
	"github.com/spf13/cobra"
)

var cardCmd = &cobra.Command{
	Use:   "card",
	Short: "Manage cards",
	Long:  "Commands for managing Fizzy cards.",
}

// Card list flags
var cardListBoard string
var cardListColumn string
var cardListTag string
var cardListIndexedBy string
var cardListAssignee string
var cardListSort string
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
		columnFilter := strings.TrimSpace(cardListColumn)
		indexedByFilter := strings.TrimSpace(cardListIndexedBy)
		effectiveIndexedBy := indexedByFilter

		client := getClient()
		path := "/cards.json"

		var params []string
		if boardID != "" {
			params = append(params, "board_ids[]="+boardID)
		}

		clientSideColumnFilter := ""
		clientSideTriage := false
		if columnFilter != "" {
			if pseudo, ok := parsePseudoColumnID(columnFilter); ok {
				switch pseudo.Kind {
				case "not_now":
					if effectiveIndexedBy != "" && effectiveIndexedBy != "not_now" {
						exitWithError(errors.NewInvalidArgsError("cannot combine --indexed-by with --column maybe"))
					}
					effectiveIndexedBy = "not_now"
				case "closed":
					if effectiveIndexedBy != "" && effectiveIndexedBy != "closed" {
						exitWithError(errors.NewInvalidArgsError("cannot combine --indexed-by with --column done"))
					}
					effectiveIndexedBy = "closed"
				case "triage":
					if effectiveIndexedBy != "" {
						exitWithError(errors.NewInvalidArgsError("cannot combine --indexed-by with --column not-yet"))
					}
					clientSideTriage = true
				default:
					clientSideColumnFilter = columnFilter
				}
			} else {
				if effectiveIndexedBy != "" {
					exitWithError(errors.NewInvalidArgsError("cannot combine --indexed-by with --column"))
				}
				clientSideColumnFilter = columnFilter
			}
		}

		if effectiveIndexedBy != "" {
			params = append(params, "indexed_by="+effectiveIndexedBy)
		}

		if cardListTag != "" {
			params = append(params, "tag_ids[]="+cardListTag)
		}
		if cardListAssignee != "" {
			params = append(params, "assignee_ids[]="+cardListAssignee)
		}
		if cardListSort != "" {
			params = append(params, "sorted_by="+cardListSort)
		}
		if cardListPage > 0 {
			params = append(params, "page="+strconv.Itoa(cardListPage))
		}
		if len(params) > 0 {
			path += "?" + strings.Join(params, "&")
		}

		if (clientSideTriage || clientSideColumnFilter != "") && !cardListAll && cardListPage == 0 {
			exitWithError(errors.NewInvalidArgsError("Filtering by column requires --all (or --page) because it is applied client-side"))
		}

		resp, err := client.GetWithPagination(path, cardListAll)
		if err != nil {
			exitWithError(err)
		}

		if clientSideTriage || clientSideColumnFilter != "" {
			arr, ok := resp.Data.([]interface{})
			if !ok {
				exitWithError(errors.NewError("Unexpected cards list response"))
			}

			filtered := make([]interface{}, 0, len(arr))
			for _, item := range arr {
				card, ok := item.(map[string]interface{})
				if !ok {
					continue
				}

				columnID := ""
				if v, ok := card["column_id"].(string); ok {
					columnID = v
				}
				if columnID == "" {
					if col, ok := card["column"].(map[string]interface{}); ok {
						if id, ok := col["id"].(string); ok {
							columnID = id
						}
					}
				}

				if clientSideTriage {
					if columnID == "" {
						filtered = append(filtered, item)
					}
					continue
				}

				if clientSideColumnFilter != "" && columnID == clientSideColumnFilter {
					filtered = append(filtered, item)
				}
			}

			resp.Data = filtered
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

		cardParams := map[string]interface{}{
			"title": cardCreateTitle,
		}

		// Handle description
		if cardCreateDescriptionFile != "" {
			content, err := os.ReadFile(cardCreateDescriptionFile)
			if err != nil {
				exitWithError(err)
			}
			cardParams["description"] = string(content)
		} else if cardCreateDescription != "" {
			cardParams["description"] = cardCreateDescription
		}

		if cardCreateTagIDs != "" {
			cardParams["tag_ids"] = cardCreateTagIDs
		}
		if cardCreateImage != "" {
			cardParams["image"] = cardCreateImage
		}
		if cardCreateCreatedAt != "" {
			cardParams["created_at"] = cardCreateCreatedAt
		}

		body := map[string]interface{}{
			"board_id": boardID,
			"card":     cardParams,
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
var cardUpdateImage string
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

		cardParams := make(map[string]interface{})

		if cardUpdateTitle != "" {
			cardParams["title"] = cardUpdateTitle
		}
		if cardUpdateDescriptionFile != "" {
			content, err := os.ReadFile(cardUpdateDescriptionFile)
			if err != nil {
				exitWithError(err)
			}
			cardParams["description"] = string(content)
		} else if cardUpdateDescription != "" {
			cardParams["description"] = cardUpdateDescription
		}
		if cardUpdateImage != "" {
			cardParams["image"] = cardUpdateImage
		}
		if cardUpdateCreatedAt != "" {
			cardParams["created_at"] = cardUpdateCreatedAt
		}

		body := map[string]interface{}{
			"card": cardParams,
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

		client := getClient()
		if pseudo, ok := parsePseudoColumnID(cardColumnColumn); ok {
			switch pseudo.Kind {
			case "triage":
				resp, err := client.Delete("/cards/" + args[0] + "/triage.json")
				if err != nil {
					exitWithError(err)
				}
				if resp != nil && resp.Data != nil {
					printSuccess(resp.Data)
				} else {
					printSuccess(map[string]interface{}{})
				}
				return
			case "not_now":
				resp, err := client.Post("/cards/"+args[0]+"/not_now.json", nil)
				if err != nil {
					exitWithError(err)
				}
				if resp != nil && resp.Data != nil {
					printSuccess(resp.Data)
				} else {
					printSuccess(map[string]interface{}{})
				}
				return
			case "closed":
				resp, err := client.Post("/cards/"+args[0]+"/closure.json", nil)
				if err != nil {
					exitWithError(err)
				}
				if resp != nil && resp.Data != nil {
					printSuccess(resp.Data)
				} else {
					printSuccess(map[string]interface{}{})
				}
				return
			}
		}

		body := map[string]interface{}{
			"column_id": cardColumnColumn,
		}

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

var cardImageRemoveCmd = &cobra.Command{
	Use:   "image-remove CARD_NUMBER",
	Short: "Remove card header image",
	Long:  "Removes the header image from a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Delete("/cards/" + args[0] + "/image.json")
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

var cardGoldenCmd = &cobra.Command{
	Use:   "golden CARD_NUMBER",
	Short: "Mark card as golden",
	Long:  "Marks a card as golden (starred/important).",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Post("/cards/"+args[0]+"/goldness.json", nil)
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

var cardUngoldenCmd = &cobra.Command{
	Use:   "ungolden CARD_NUMBER",
	Short: "Remove golden status from card",
	Long:  "Removes the golden (starred/important) status from a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		client := getClient()
		resp, err := client.Delete("/cards/" + args[0] + "/goldness.json")
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
	cardListCmd.Flags().StringVar(&cardListColumn, "column", "", "Filter by column ID or pseudo column (not-yet, maybe, done)")
	cardListCmd.Flags().StringVar(&cardListTag, "tag", "", "Filter by tag ID")
	cardListCmd.Flags().StringVar(&cardListIndexedBy, "indexed-by", "", "Filter by lane/index (all, closed, not_now, stalled, postponing_soon, golden)")
	cardListCmd.Flags().StringVar(&cardListIndexedBy, "status", "", "Alias for --indexed-by")
	_ = cardListCmd.Flags().MarkDeprecated("status", "use --indexed-by")
	cardListCmd.Flags().StringVar(&cardListAssignee, "assignee", "", "Filter by assignee ID")
	cardListCmd.Flags().StringVar(&cardListSort, "sort", "", "Sort order: newest, oldest, or latest (default)")
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
	cardUpdateCmd.Flags().StringVar(&cardUpdateImage, "image", "", "Header image signed ID")
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

	// Image removal
	cardCmd.AddCommand(cardImageRemoveCmd)

	// Golden
	cardCmd.AddCommand(cardGoldenCmd)
	cardCmd.AddCommand(cardUngoldenCmd)
}
