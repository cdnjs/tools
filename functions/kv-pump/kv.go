package kv_pump

type FileMetadata struct {
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	SRI          string `json:"sri,omitempty"`
}
