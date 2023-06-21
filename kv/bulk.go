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

type ConsumableWriteRequest struct {
	Key   string
	Name  string
	Value []byte
	Meta  *FileMetadata
}

func (r ConsumableWriteRequest) GetKey() string  { return r.Key }
func (r ConsumableWriteRequest) GetName() string { return r.Name }
func (r ConsumableWriteRequest) GetValue() []byte {
	if r.Value == nil {
		panic(r.GetName() + ": write request has already been consumed")
	}
	return r.Value
}
func (r ConsumableWriteRequest) GetMeta() *FileMetadata { return r.Meta }
func (r ConsumableWriteRequest) Consumed() {
	r.Value = nil //nolint:all
}

type MetaWriteRequest struct {
	Key  string
	Name string
	Meta *FileMetadata
}

func (r MetaWriteRequest) GetKey() string         { return r.Key }
func (r MetaWriteRequest) GetName() string        { return r.Name }
func (r MetaWriteRequest) GetValue() []byte       { return nil }
func (r MetaWriteRequest) GetMeta() *FileMetadata { return r.Meta }
func (r MetaWriteRequest) Consumed()              {}
