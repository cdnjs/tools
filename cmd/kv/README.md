# KV

Tools to test our Workers KV namespace.

## `upload`

Inserts packages from disk to KV. Package files and version metadata will be pushed to KV.

## `upload-meta`

Inserts respective package metadata JSON files from disk to KV. 

**Make sure the bot is not running to avoid KV write race conditions for the latest package version**.
