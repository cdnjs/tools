# KV

Tools to test our Workers KV namespace.

## `upload`

Inserts packages from disk to KV. Package files and version metadata will be pushed to KV.

```
make kv && ./bin/kv upload jquery mathjax fontawesome
```

## `meta`

Gets all metadata associated with a package in KV.

```
make kv && ./bin/kv meta jquery
```
