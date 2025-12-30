package commands

import (
	"os"

	"github.com/robzolkos/fizzy-cli/internal/errors"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload files",
	Long:  "Commands for uploading files for use in rich text fields.",
}

var uploadFileCmd = &cobra.Command{
	Use:   "file PATH",
	Short: "Upload a file",
	Long:  "Uploads a file and returns a signed_id for use in rich text fields.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		filePath := args[0]

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			exitWithError(errors.NewError("File not found: " + filePath))
		}

		client := getClient()
		resp, err := client.UploadFile(filePath)
		if err != nil {
			exitWithError(err)
		}

		printSuccess(resp.Data)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.AddCommand(uploadFileCmd)
}
