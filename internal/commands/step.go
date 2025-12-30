package commands

import (
	"github.com/spf13/cobra"
)

var stepCmd = &cobra.Command{
	Use:   "step",
	Short: "Manage steps (to-do items)",
	Long:  "Commands for managing card steps (to-do items).",
}

// Step show flags
var stepShowCard string

var stepShowCmd = &cobra.Command{
	Use:   "show STEP_ID",
	Short: "Show a step",
	Long:  "Shows details of a specific step.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if stepShowCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		client := getClient()
		resp, err := client.Get("/cards/" + stepShowCard + "/steps/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Step create flags
var stepCreateCard string
var stepCreateContent string
var stepCreateCompleted bool

var stepCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a step",
	Long:  "Creates a new step (to-do item) on a card.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if stepCreateCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}
		if stepCreateContent == "" {
			exitWithError(newRequiredFlagError("content"))
		}

		body := map[string]interface{}{
			"content": stepCreateContent,
		}
		if stepCreateCompleted {
			body["completed"] = true
		}

		client := getClient()
		resp, err := client.Post("/cards/"+stepCreateCard+"/steps.json", body)
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

// Step update flags
var stepUpdateCard string
var stepUpdateContent string
var stepUpdateCompleted bool
var stepUpdateNotCompleted bool

var stepUpdateCmd = &cobra.Command{
	Use:   "update STEP_ID",
	Short: "Update a step",
	Long:  "Updates an existing step.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if stepUpdateCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		body := make(map[string]interface{})

		if stepUpdateContent != "" {
			body["content"] = stepUpdateContent
		}
		if stepUpdateCompleted {
			body["completed"] = true
		}
		if stepUpdateNotCompleted {
			body["completed"] = false
		}

		client := getClient()
		resp, err := client.Patch("/cards/"+stepUpdateCard+"/steps/"+args[0]+".json", body)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

// Step delete flags
var stepDeleteCard string

var stepDeleteCmd = &cobra.Command{
	Use:   "delete STEP_ID",
	Short: "Delete a step",
	Long:  "Deletes a step from a card.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		if stepDeleteCard == "" {
			exitWithError(newRequiredFlagError("card"))
		}

		client := getClient()
		_, err := client.Delete("/cards/" + stepDeleteCard + "/steps/" + args[0] + ".json")
		if err != nil {
			exitWithError(err)
		}

		printSuccess(map[string]interface{}{
			"deleted": true,
		})
	},
}

func init() {
	rootCmd.AddCommand(stepCmd)

	// Show
	stepShowCmd.Flags().StringVar(&stepShowCard, "card", "", "Card number (required)")
	stepCmd.AddCommand(stepShowCmd)

	// Create
	stepCreateCmd.Flags().StringVar(&stepCreateCard, "card", "", "Card number (required)")
	stepCreateCmd.Flags().StringVar(&stepCreateContent, "content", "", "Step content (required)")
	stepCreateCmd.Flags().BoolVar(&stepCreateCompleted, "completed", false, "Mark as completed")
	stepCmd.AddCommand(stepCreateCmd)

	// Update
	stepUpdateCmd.Flags().StringVar(&stepUpdateCard, "card", "", "Card number (required)")
	stepUpdateCmd.Flags().StringVar(&stepUpdateContent, "content", "", "Step content")
	stepUpdateCmd.Flags().BoolVar(&stepUpdateCompleted, "completed", false, "Mark as completed")
	stepUpdateCmd.Flags().BoolVar(&stepUpdateNotCompleted, "not_completed", false, "Mark as not completed")
	stepCmd.AddCommand(stepUpdateCmd)

	// Delete
	stepDeleteCmd.Flags().StringVar(&stepDeleteCard, "card", "", "Card number (required)")
	stepCmd.AddCommand(stepDeleteCmd)
}
