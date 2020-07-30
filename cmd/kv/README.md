# KV

Tools to test our Workers KV namespace.

## `upload`

Inserts packages from disk to KV. Package files and version metadata will be pushed to KV.

```
make kv && ./bin/kv upload jquery mathjax fontawesome
```

## `files`

Gets the file names stored in KV for a package.
Note that currently if there are more than 1000 files, it will only note that 1000 exist.

```
make kv && ./bin/kv files jquery
```

## `meta`

Gets all metadata associated with a package in KV.
Note that currently if there are more than 1000 assets for a version, it will only note that 1000 exist.

```
make kv && ./bin/kv meta jquery
```
