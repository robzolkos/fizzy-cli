package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reactionCmd = &cobra.Command{
	Use:   "reaction",
	Short: "Manage reactions",
	Long:  "Commands for managing comment reactions.",
}

// Reaction list flags
var reactionListCard string
var reactionListComment string

var reactionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List reactions for a comment",
	Long:  "Lists all reactions for a specific comment.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if reactionListCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}
		if reactionListComment == "" {
			exitWithError(newRequiredFlagError("comment"))
		}

		client := getClient()
		resp, err := client.Get("/cards/" + reactionListCard + "/comments/" + reactionListComment + "/reactions.json")
		if err != nil {
			exitWithError(err)
		}

		// Build summary
		count := 0
		if arr, ok := resp.Data.([]interface{}); ok {
			count = len(arr)
		}
		summary := fmt.Sprintf("%d reactions on comment", count)

		printSuccessWithSummary(resp.Data, summary)
	},
}

// Reaction create flags
var reactionCreateCard string
var reactionCreateComment string
var reactionCreateContent string

var reactionCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a reaction to a comment",
	Long:  "Adds an emoji reaction to a comment.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if reactionCreateCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}
		if reactionCreateComment == "" {
			exitWithError(newRequiredFlagError("comment"))
		}
		if reactionCreateContent == "" {
			exitWithError(newRequiredFlagError("content"))
		}

		body := map[string]interface{}{
			"content": reactionCreateContent,
		}

		client := getClient()
		resp, err := client.Post("/cards/"+reactionCreateCard+"/comments/"+reactionCreateComment+"/reactions.json", body)
		if err != nil {
			exitWithError(err)
		}

		// Reaction create returns just success, no location or data
		if resp.Data != nil {
			printSuccess(resp.Data)
		} else {
			printSuccess(map[string]interface{}{})
		}
	},
}

// Reaction delete flags
var reactionDeleteCard string
var reactionDeleteComment string

var reactionDeleteCmd = &cobra.Command{
	Use:   "delete REACTION_ID",
	Short: "Remove a reaction",
	Long:  "Removes a reaction from a comment.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if reactionDeleteCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}
		if reactionDeleteComment == "" {
			exitWithError(newRequiredFlagError("comment"))
		}

		client := getClient()
		_, err := client.Delete("/cards/" + reactionDeleteCard + "/comments/" + reactionDeleteComment + "/reactions/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"deleted": true,
		})
	},
}

func init() {
	rootCmd.AddCommand(reactionCmd)

	// List
	reactionListCmd.Flags().StringVar(&reactionListCard, "card", "", "Card number (required)")
	reactionListCmd.Flags().StringVar(&reactionListComment, "comment", "", "Comment ID (required)")
	reactionCmd.AddCommand(reactionListCmd)

	// Create
	reactionCreateCmd.Flags().StringVar(&reactionCreateCard, "card", "", "Card number (required)")
	reactionCreateCmd.Flags().StringVar(&reactionCreateComment, "comment", "", "Comment ID (required)")
	reactionCreateCmd.Flags().StringVar(&reactionCreateContent, "content", "", "Emoji content (required)")
	reactionCmd.AddCommand(reactionCreateCmd)

	// Delete
	reactionDeleteCmd.Flags().StringVar(&reactionDeleteCard, "card", "", "Card number (required)")
	reactionDeleteCmd.Flags().StringVar(&reactionDeleteComment, "comment", "", "Comment ID (required)")
	reactionCmd.AddCommand(reactionDeleteCmd)
}
