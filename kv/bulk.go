package kv

// FileMetadata represents metadata for a
// particular KV.
type FileMetadata struct {
	ETag         string `json:"etag,omitempty"`
	LastModified string `json:"last_modified,omitempty"`
	SRI          string `json:"sri,omitempty"`
}

// Represents a KV write request, consisting of
// a string key, a []byte value, and file metadata.
// The name field is used to identify this write request
// with a human-readable friendly name.
type WriteRequest interface {
	GetKey() string
	GetName() string
	GetValue() []byte
	GetMeta() *FileMetadata

	// notify that we consumed the write request
	Consumed()
}

type InMemoryWriteRequest struct {
	Key   string
	Name  string
	Value []byte
	Meta  *FileMetadata
}

func (r InMemoryWriteRequest) GetKey() string  { return r.Key }
func (r InMemoryWriteRequest) GetName() string { return r.Name }
func (r InMemoryWriteRequest) GetValue() []byte {
	if r.Value == nil {
		panic("write request has already been consumed")
	}
	return r.Value
}
func (r InMemoryWriteRequest) GetMeta() *FileMetadata { return r.Meta }
func (r InMemoryWriteRequest) Consumed() {
	r.Value = nil
}
