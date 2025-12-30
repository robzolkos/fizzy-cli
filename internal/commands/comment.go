package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var commentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Manage comments",
	Long:  "Commands for managing card comments.",
}

// Comment list flags
var commentListCard string
var commentListPage int
var commentListAll bool

var commentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List comments for a card",
	Long:  "Lists all comments for a specific card.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if commentListCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		client := getClient()
		path := "/cards/" + commentListCard + "/comments.json"
		if commentListPage > 0 {
			path += "?page=" + string(rune(commentListPage+'0'))
		}

		resp, err := client.GetWithPagination(path, commentListAll)
		if err != nil {
			exitWithError(err)
		}

		hasNext := resp.LinkNext != ""
		printSuccessWithPagination(resp.Data, hasNext, resp.LinkNext)
	},
}

// Comment show flags
var commentShowCard string

var commentShowCmd = &cobra.Command{
	Use:   "show COMMENT_ID",
	Short: "Show a comment",
	Long:  "Shows details of a specific comment.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if commentShowCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		client := getClient()
		resp, err := client.Get("/cards/" + commentShowCard + "/comments/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Comment create flags
var commentCreateCard string
var commentCreateBody string
var commentCreateBodyFile string
var commentCreateCreatedAt string

var commentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a comment",
	Long:  "Creates a new comment on a card.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if commentCreateCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		// Determine body content
		var body string
		if commentCreateBodyFile != "" {
			content, err := os.ReadFile(commentCreateBodyFile)
			if err != nil {
				exitWithError(err)
			}
			body = string(content)
		} else if commentCreateBody != "" {
			body = commentCreateBody
		} else {
			exitWithError(newRequiredFlagError("body or body_file"))
		}

		commentParams := map[string]interface{}{
			"body": body,
		}
		if commentCreateCreatedAt != "" {
			commentParams["created_at"] = commentCreateCreatedAt
		}

		reqBody := map[string]interface{}{
			"comment": commentParams,
		}

		client := getClient()
		resp, err := client.Post("/cards/"+commentCreateCard+"/comments.json", reqBody)
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

// Comment update flags
var commentUpdateCard string
var commentUpdateBody string
var commentUpdateBodyFile string

var commentUpdateCmd = &cobra.Command{
	Use:   "update COMMENT_ID",
	Short: "Update a comment",
	Long:  "Updates an existing comment.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if commentUpdateCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		commentParams := make(map[string]interface{})

		if commentUpdateBodyFile != "" {
			content, err := os.ReadFile(commentUpdateBodyFile)
			if err != nil {
				exitWithError(err)
			}
			commentParams["body"] = string(content)
		} else if commentUpdateBody != "" {
			commentParams["body"] = commentUpdateBody
		}

		reqBody := map[string]interface{}{
			"comment": commentParams,
		}

		client := getClient()
		resp, err := client.Patch("/cards/"+commentUpdateCard+"/comments/"+args[0]+".json", reqBody)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Comment delete flags
var commentDeleteCard string

var commentDeleteCmd = &cobra.Command{
	Use:   "delete COMMENT_ID",
	Short: "Delete a comment",
	Long:  "Deletes a comment from a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if commentDeleteCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		client := getClient()
		_, err := client.Delete("/cards/" + commentDeleteCard + "/comments/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"deleted": true,
		})
	},
}

func init() {
	rootCmd.AddCommand(commentCmd)

	// List
	commentListCmd.Flags().StringVar(&commentListCard, "card", "", "Card number (required)")
	commentListCmd.Flags().IntVar(&commentListPage, "page", 0, "Page number")
	commentListCmd.Flags().BoolVar(&commentListAll, "all", false, "Fetch all pages")
	commentCmd.AddCommand(commentListCmd)

	// Show
	commentShowCmd.Flags().StringVar(&commentShowCard, "card", "", "Card number (required)")
	commentCmd.AddCommand(commentShowCmd)

	// Create
	commentCreateCmd.Flags().StringVar(&commentCreateCard, "card", "", "Card number (required)")
	commentCreateCmd.Flags().StringVar(&commentCreateBody, "body", "", "Comment body (HTML)")
	commentCreateCmd.Flags().StringVar(&commentCreateBodyFile, "body_file", "", "Read body from file")
	commentCreateCmd.Flags().StringVar(&commentCreateCreatedAt, "created-at", "", "Custom created_at timestamp")
	commentCmd.AddCommand(commentCreateCmd)

	// Update
	commentUpdateCmd.Flags().StringVar(&commentUpdateCard, "card", "", "Card number (required)")
	commentUpdateCmd.Flags().StringVar(&commentUpdateBody, "body", "", "Comment body (HTML)")
	commentUpdateCmd.Flags().StringVar(&commentUpdateBodyFile, "body_file", "", "Read body from file")
	commentCmd.AddCommand(commentUpdateCmd)

	// Delete
	commentDeleteCmd.Flags().StringVar(&commentDeleteCard, "card", "", "Card number (required)")
	commentCmd.AddCommand(commentDeleteCmd)
}
