package commands

import (
	"testing"
)

func TestParseAttachments(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []Attachment
	}{
		{
			name: "single attachment with all attributes",
			html: `<div class="action-text-content">
  <action-text-attachment sgid="eyJfcmFpbHMiOnt9fQ==" content-type="image/png" filename="screenshot.png" filesize="332363" width="794" height="2312" previewable="true">
    <figure class="attachment">
      <a href="/123/rails/active_storage/blobs/redirect/abc123/screenshot.png?disposition=attachment">Download</a>
    </figure>
  </action-text-attachment>
</div>`,
			expected: []Attachment{
				{
					Index:       1,
					Filename:    "screenshot.png",
					ContentType: "image/png",
					Filesize:    332363,
					Width:       794,
					Height:      2312,
					SGID:        "eyJfcmFpbHMiOnt9fQ==",
					DownloadURL: "/123/rails/active_storage/blobs/redirect/abc123/screenshot.png?disposition=attachment",
				},
			},
		},
		{
			name: "multiple attachments",
			html: `<div>
  <action-text-attachment sgid="sgid1" content-type="image/png" filename="image1.png" filesize="1000" width="100" height="100">
    <a href="/rails/active_storage/blobs/redirect/blob1/image1.png?disposition=attachment">Download</a>
  </action-text-attachment>
  <action-text-attachment sgid="sgid2" content-type="application/pdf" filename="document.pdf" filesize="2000">
    <a href="/rails/active_storage/blobs/redirect/blob2/document.pdf?disposition=attachment">Download</a>
  </action-text-attachment>
</div>`,
			expected: []Attachment{
				{
					Index:       1,
					Filename:    "image1.png",
					ContentType: "image/png",
					Filesize:    1000,
					Width:       100,
					Height:      100,
					SGID:        "sgid1",
					DownloadURL: "/rails/active_storage/blobs/redirect/blob1/image1.png?disposition=attachment",
				},
				{
					Index:       2,
					Filename:    "document.pdf",
					ContentType: "application/pdf",
					Filesize:    2000,
					SGID:        "sgid2",
					DownloadURL: "/rails/active_storage/blobs/redirect/blob2/document.pdf?disposition=attachment",
				},
			},
		},
		{
			name:     "no attachments",
			html:     `<div class="action-text-content"><p>Just some text</p></div>`,
			expected: []Attachment{},
		},
		{
			name:     "empty html",
			html:     "",
			expected: []Attachment{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAttachments(tt.html)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d attachments, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				actual := result[i]

				if actual.Index != expected.Index {
					t.Errorf("attachment %d: expected index %d, got %d", i, expected.Index, actual.Index)
				}
				if actual.Filename != expected.Filename {
					t.Errorf("attachment %d: expected filename %s, got %s", i, expected.Filename, actual.Filename)
				}
				if actual.ContentType != expected.ContentType {
					t.Errorf("attachment %d: expected content-type %s, got %s", i, expected.ContentType, actual.ContentType)
				}
				if actual.Filesize != expected.Filesize {
					t.Errorf("attachment %d: expected filesize %d, got %d", i, expected.Filesize, actual.Filesize)
				}
				if actual.Width != expected.Width {
					t.Errorf("attachment %d: expected width %d, got %d", i, expected.Width, actual.Width)
				}
				if actual.Height != expected.Height {
					t.Errorf("attachment %d: expected height %d, got %d", i, expected.Height, actual.Height)
				}
				if actual.SGID != expected.SGID {
					t.Errorf("attachment %d: expected sgid %s, got %s", i, expected.SGID, actual.SGID)
				}
				if actual.DownloadURL != expected.DownloadURL {
					t.Errorf("attachment %d: expected download_url %s, got %s", i, expected.DownloadURL, actual.DownloadURL)
				}
			}
		})
	}
}

func TestCardAttachmentsCommand(t *testing.T) {
	tests := []struct {
		name          string
		cardNumber    string
		cardData      map[string]interface{}
		expectSuccess bool
		expectError   string
		expectedCount int
	}{
		{
			name:       "card with attachments",
			cardNumber: "241",
			cardData: map[string]interface{}{
				"id":     "card-id",
				"number": 241,
				"description_html": `<action-text-attachment sgid="test-sgid" content-type="image/png" filename="test.png" filesize="1000">
					<a href="/rails/active_storage/blobs/redirect/blob/test.png?disposition=attachment">Download</a>
				</action-text-attachment>`,
			},
			expectSuccess: true,
			expectedCount: 1,
		},
		{
			name:       "card without attachments",
			cardNumber: "100",
			cardData: map[string]interface{}{
				"id":               "card-id",
				"number":           100,
				"description_html": "<p>No attachments here</p>",
			},
			expectSuccess: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient().WithGetData(tt.cardData)
			result := SetTestMode(mock)
			SetTestConfig("test-token", "test-account", "https://api.test.com")
			defer ResetTestMode()

			rootCmd.SetArgs([]string{"card", "attachments", "show", tt.cardNumber})

			RunTestCommand(func() {
				_ = rootCmd.Execute()
			})

			if result.Response == nil {
				t.Fatal("expected response, got nil")
			}

			if tt.expectSuccess {
				if !result.Response.Success {
					t.Errorf("expected success, got error: %v", result.Response)
				}

				// Check that attachments were parsed
				if attachments, ok := result.Response.Data.([]Attachment); ok {
					if len(attachments) != tt.expectedCount {
						t.Errorf("expected %d attachments, got %d", tt.expectedCount, len(attachments))
					}
				}
			} else {
				if result.Response.Success {
					t.Errorf("expected error containing %q, got success", tt.expectError)
				}
			}
		})
	}
}

func TestExtractAttr(t *testing.T) {
	tests := []struct {
		attrs    string
		name     string
		expected string
	}{
		{
			attrs:    `sgid="abc123" content-type="image/png"`,
			name:     "sgid",
			expected: "abc123",
		},
		{
			attrs:    `sgid="abc123" content-type="image/png"`,
			name:     "content-type",
			expected: "image/png",
		},
		{
			attrs:    `filename="test file.png"`,
			name:     "filename",
			expected: "test file.png",
		},
		{
			attrs:    `width="100" height="200"`,
			name:     "missing",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractAttr(tt.attrs, tt.name)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestCardAttachmentsDownloadCommand(t *testing.T) {
	tests := []struct {
		name                string
		args                []string
		cardData            map[string]interface{}
		downloadError       error
		expectSuccess       bool
		expectError         string
		expectedDownloads   int
		expectedDownloadURL string
	}{
		{
			name: "download single attachment by index",
			args: []string{"card", "attachments", "download", "241", "1"},
			cardData: map[string]interface{}{
				"id":     "card-id",
				"number": 241,
				"description_html": `<action-text-attachment sgid="test-sgid" content-type="image/png" filename="test.png" filesize="1000">
					<a href="/rails/active_storage/blobs/redirect/blob/test.png?disposition=attachment">Download</a>
				</action-text-attachment>`,
			},
			expectSuccess:       true,
			expectedDownloads:   1,
			expectedDownloadURL: "/rails/active_storage/blobs/redirect/blob/test.png?disposition=attachment",
		},
		{
			name: "download all attachments",
			args: []string{"card", "attachments", "download", "241"},
			cardData: map[string]interface{}{
				"id":     "card-id",
				"number": 241,
				"description_html": `<action-text-attachment sgid="sgid1" content-type="image/png" filename="image1.png" filesize="1000">
					<a href="/blobs/blob1/image1.png?disposition=attachment">Download</a>
				</action-text-attachment>
				<action-text-attachment sgid="sgid2" content-type="application/pdf" filename="doc.pdf" filesize="2000">
					<a href="/blobs/blob2/doc.pdf?disposition=attachment">Download</a>
				</action-text-attachment>`,
			},
			expectSuccess:     true,
			expectedDownloads: 2,
		},
		{
			name: "no attachments on card",
			args: []string{"card", "attachments", "download", "100"},
			cardData: map[string]interface{}{
				"id":               "card-id",
				"number":           100,
				"description_html": "<p>No attachments here</p>",
			},
			expectSuccess: false,
			expectError:   "No attachments found",
		},
		{
			name: "invalid attachment index - not a number",
			args: []string{"card", "attachments", "download", "241", "abc"},
			cardData: map[string]interface{}{
				"id":     "card-id",
				"number": 241,
				"description_html": `<action-text-attachment sgid="test-sgid" content-type="image/png" filename="test.png" filesize="1000">
					<a href="/blobs/blob/test.png?disposition=attachment">Download</a>
				</action-text-attachment>`,
			},
			expectSuccess: false,
			expectError:   "attachment index must be a number",
		},
		{
			name: "attachment index out of range - too high",
			args: []string{"card", "attachments", "download", "241", "5"},
			cardData: map[string]interface{}{
				"id":     "card-id",
				"number": 241,
				"description_html": `<action-text-attachment sgid="test-sgid" content-type="image/png" filename="test.png" filesize="1000">
					<a href="/blobs/blob/test.png?disposition=attachment">Download</a>
				</action-text-attachment>`,
			},
			expectSuccess: false,
			expectError:   "attachment index must be between 1 and 1",
		},
		{
			name: "attachment index out of range - zero",
			args: []string{"card", "attachments", "download", "241", "0"},
			cardData: map[string]interface{}{
				"id":     "card-id",
				"number": 241,
				"description_html": `<action-text-attachment sgid="test-sgid" content-type="image/png" filename="test.png" filesize="1000">
					<a href="/blobs/blob/test.png?disposition=attachment">Download</a>
				</action-text-attachment>`,
			},
			expectSuccess: false,
			expectError:   "attachment index must be between 1 and 1",
		},
		{
			name: "download error",
			args: []string{"card", "attachments", "download", "241", "1"},
			cardData: map[string]interface{}{
				"id":     "card-id",
				"number": 241,
				"description_html": `<action-text-attachment sgid="test-sgid" content-type="image/png" filename="test.png" filesize="1000">
					<a href="/blobs/blob/test.png?disposition=attachment">Download</a>
				</action-text-attachment>`,
			},
			downloadError: &MockError{message: "connection refused"},
			expectSuccess: false,
			expectError:   "connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient().WithGetData(tt.cardData)
			if tt.downloadError != nil {
				mock.DownloadFileError = tt.downloadError
			}
			result := SetTestMode(mock)
			SetTestConfig("test-token", "test-account", "https://api.test.com")
			defer ResetTestMode()

			rootCmd.SetArgs(tt.args)

			RunTestCommand(func() {
				_ = rootCmd.Execute()
			})

			if result.Response == nil {
				t.Fatal("expected response, got nil")
			}

			if tt.expectSuccess {
				if !result.Response.Success {
					t.Errorf("expected success, got error: %v", result.Response)
				}

				// Verify downloads were called
				if len(mock.DownloadFileCalls) != tt.expectedDownloads {
					t.Errorf("expected %d downloads, got %d", tt.expectedDownloads, len(mock.DownloadFileCalls))
				}

				// Verify download URL if specified
				if tt.expectedDownloadURL != "" && len(mock.DownloadFileCalls) > 0 {
					if mock.DownloadFileCalls[0].URLPath != tt.expectedDownloadURL {
						t.Errorf("expected download URL %q, got %q", tt.expectedDownloadURL, mock.DownloadFileCalls[0].URLPath)
					}
				}

				// Verify response data
				if data, ok := result.Response.Data.(map[string]interface{}); ok {
					if downloaded, ok := data["downloaded"].(int); ok {
						if downloaded != tt.expectedDownloads {
							t.Errorf("expected downloaded count %d, got %d", tt.expectedDownloads, downloaded)
						}
					}
				}
			} else {
				if result.Response.Success {
					t.Errorf("expected error containing %q, got success", tt.expectError)
				}
				if tt.expectError != "" && result.Response.Error != nil {
					if !containsString(result.Response.Error.Message, tt.expectError) {
						t.Errorf("expected error containing %q, got %q", tt.expectError, result.Response.Error.Message)
					}
				}
			}
		})
	}
}

// MockError implements the error interface for testing
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}

// containsString checks if s contains substr
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
