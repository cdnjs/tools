# KV

Tools to test our Workers KV namespace.

#### `upload "<package 1>" "<package 2>" ... "<package n>"`

Inserts `n` packages from disk to KV. Package files and version metadata will be pushed to KV.

#### `upload-meta "<package 1>" "<package 2>" ... "<package n>"`

Inserts `n` respective package metadata JSON files from disk to KV. 

**Make sure the bot is not running to avoid KV write race conditions for the latest package version**.
