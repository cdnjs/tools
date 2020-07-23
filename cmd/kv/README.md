# KV

Tools to test our Workers KV namespace.

## `upload`

Inserts packages from disk to KV. Package files and version metadata will be pushed to KV.

```
make kv && ./bin/kv upload jquery mathjax fontawesome
```

## `upload-meta`

Inserts respective package metadata JSON files from disk to KV. 

**Make sure the bot is not running to avoid KV write race conditions for the latest package version**.

```
make kv && ./bin/kv upload-meta jquery mathjax fontawesome
```

## `meta`

Gets all metadata associated with a package in KV.

```
make kv && ./bin/kv meta jquery
```
