package commands

import (
	"regexp"
	"strconv"

	"github.com/robzolkos/fizzy-cli/internal/errors"
	"github.com/spf13/cobra"
)

// Attachment represents a parsed attachment from description_html
type Attachment struct {
	Index       int    `json:"index"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Filesize    int64  `json:"filesize"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	DownloadURL string `json:"download_url"`
	SGID        string `json:"sgid"`
}

var attachmentsCmd = &cobra.Command{
	Use:   "attachments",
	Short: "Manage card attachments",
	Long:  "Commands for viewing and downloading card attachments.",
}

var attachmentsShowCmd = &cobra.Command{
	Use:   "show CARD_NUMBER",
	Short: "List attachments on a card",
	Long:  "Lists all attachments embedded in a card's description.",
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

		cardData, ok := resp.Data.(map[string]interface{})
		if !ok {
			exitWithError(errors.NewError("Invalid card response"))
		}

		descriptionHTML, _ := cardData["description_html"].(string)
		attachments := parseAttachments(descriptionHTML)

		printSuccess(attachments)
	},
}

// Attachment download flags
var attachmentDownloadOutput string

var attachmentsDownloadCmd = &cobra.Command{
	Use:   "download CARD_NUMBER [ATTACHMENT_INDEX]",
	Short: "Download attachments from a card",
	Long: `Downloads attachments from a card.

If ATTACHMENT_INDEX is provided, downloads only that attachment (1-based index).
If ATTACHMENT_INDEX is omitted, downloads all attachments.

Use 'fizzy card attachments show CARD_NUMBER' to see available attachments and their indices.`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := requireAuthAndAccount(); err != nil {
			exitWithError(err)
		}

		cardNumber := args[0]

		client := getClient()
		resp, err := client.Get("/cards/" + cardNumber + ".json")
		if err != nil {
			exitWithError(err)
		}

		cardData, ok := resp.Data.(map[string]interface{})
		if !ok {
			exitWithError(errors.NewError("Invalid card response"))
		}

		descriptionHTML, _ := cardData["description_html"].(string)
		attachments := parseAttachments(descriptionHTML)

		if len(attachments) == 0 {
			exitWithError(errors.NewNotFoundError("No attachments found on this card"))
		}

		// Determine which attachments to download
		var toDownload []Attachment
		if len(args) == 2 {
			// Download specific attachment
			attachmentIndex, err := strconv.Atoi(args[1])
			if err != nil {
				exitWithError(errors.NewInvalidArgsError("attachment index must be a number"))
			}
			if attachmentIndex < 1 || attachmentIndex > len(attachments) {
				exitWithError(errors.NewInvalidArgsError("attachment index must be between 1 and " + strconv.Itoa(len(attachments))))
			}
			toDownload = []Attachment{attachments[attachmentIndex-1]}
		} else {
			// Download all attachments
			toDownload = attachments
		}

		// Download the files
		var results []map[string]interface{}
		for _, attachment := range toDownload {
			outputPath := attachment.Filename
			// If downloading single file with custom output name
			if len(toDownload) == 1 && attachmentDownloadOutput != "" {
				outputPath = attachmentDownloadOutput
			}

			if err := client.DownloadFile(attachment.DownloadURL, outputPath); err != nil {
				exitWithError(err)
			}

			results = append(results, map[string]interface{}{
				"filename": attachment.Filename,
				"saved_to": outputPath,
				"filesize": attachment.Filesize,
			})
		}

		printSuccess(map[string]interface{}{
			"downloaded": len(results),
			"files":      results,
		})
	},
}

// parseAttachments extracts attachment information from description_html
func parseAttachments(html string) []Attachment {
	var attachments []Attachment

	// Match action-text-attachment elements with their content
	// <action-text-attachment sgid="..." content-type="..." filename="..." filesize="..." ...>...</action-text-attachment>
	attachmentRegex := regexp.MustCompile(`(?s)<action-text-attachment\s+([^>]+)>(.*?)</action-text-attachment>`)
	matches := attachmentRegex.FindAllStringSubmatch(html, -1)

	for i, match := range matches {
		if len(match) < 3 {
			continue
		}

		attrs := match[1]
		content := match[2]
		attachment := Attachment{
			Index: i + 1,
		}

		// Parse attributes
		attachment.SGID = extractAttr(attrs, "sgid")
		attachment.ContentType = extractAttr(attrs, "content-type")
		attachment.Filename = extractAttr(attrs, "filename")

		if filesize := extractAttr(attrs, "filesize"); filesize != "" {
			if size, err := strconv.ParseInt(filesize, 10, 64); err == nil {
				attachment.Filesize = size
			}
		}

		if width := extractAttr(attrs, "width"); width != "" {
			if w, err := strconv.Atoi(width); err == nil {
				attachment.Width = w
			}
		}

		if height := extractAttr(attrs, "height"); height != "" {
			if h, err := strconv.Atoi(height); err == nil {
				attachment.Height = h
			}
		}

		// Extract download URL from within this attachment's content
		downloadURLRegex := regexp.MustCompile(`href="([^"]+\?disposition=attachment)"`)
		if downloadMatch := downloadURLRegex.FindStringSubmatch(content); len(downloadMatch) > 1 {
			attachment.DownloadURL = downloadMatch[1]
		}

		// If no download URL with disposition found, try blob URL pattern within content
		if attachment.DownloadURL == "" {
			blobURLRegex := regexp.MustCompile(`href="(/[^"]+/rails/active_storage/blobs/redirect/[^"]+)"`)
			if blobMatch := blobURLRegex.FindStringSubmatch(content); len(blobMatch) > 1 {
				url := blobMatch[1]
				// Add disposition=attachment if not present
				if !regexp.MustCompile(`\?`).MatchString(url) {
					attachment.DownloadURL = url + "?disposition=attachment"
				} else {
					attachment.DownloadURL = url
				}
			}
		}

		attachments = append(attachments, attachment)
	}

	return attachments
}

// extractAttr extracts an attribute value from an HTML attribute string
func extractAttr(attrs, name string) string {
	re := regexp.MustCompile(name + `="([^"]*)"`)
	match := re.FindStringSubmatch(attrs)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func init() {
	cardCmd.AddCommand(attachmentsCmd)

	attachmentsCmd.AddCommand(attachmentsShowCmd)

	attachmentsDownloadCmd.Flags().StringVarP(&attachmentDownloadOutput, "output", "o", "", "Output filename (default: original filename, only applies when downloading single attachment)")
	attachmentsCmd.AddCommand(attachmentsDownloadCmd)
}
