# KV

Tools to test our Workers KV namespace.

## `upload`

Inserts packages from disk to KV. Package files and version metadata will be pushed to KV.
If the flag `-meta-only` is set, only version metadata will be pushed to KV.

```
make kv && ./bin/kv upload jquery mathjax fontawesome
```

## `upload-aggregate`

Inserts aggregate metadata to KV from scratch by scraping KV entries for package-level and version-specific metadata.

```
make kv && ./bin/kv upload-aggregate jquery mathjax fontawesome
```

## `files`

Gets the file names stored in KV for a package.
Note that currently if there are more than 1000 files, it will only note that 1000 exist.

```
make kv && ./bin/kv files jquery
```

## `meta`

Gets all metadata associated with a package in KV.
Note that currently if there are more than 1000 versions for a package, it will only process the first 1000.

```
make kv && ./bin/kv meta jquery
```
