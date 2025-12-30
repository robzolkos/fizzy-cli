package client

// API defines the interface for API operations.
// This allows for mocking in tests.
type API interface {
	Get(path string) (*APIResponse, error)
	Post(path string, body interface{}) (*APIResponse, error)
	Patch(path string, body interface{}) (*APIResponse, error)
	Put(path string, body interface{}) (*APIResponse, error)
	Delete(path string) (*APIResponse, error)
	GetWithPagination(path string, fetchAll bool) (*APIResponse, error)
	FollowLocation(location string) (*APIResponse, error)
	UploadFile(filePath string) (*APIResponse, error)
}

// Ensure Client implements API interface
var _ API = (*Client)(nil)
