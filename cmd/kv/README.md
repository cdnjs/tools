# KV

Tools to test our Workers KV namespace.

## `upload`

Inserts packages from disk to KV. Package files and version metadata will be pushed to KV.
If the flag `-meta-only` is set, only version metadata will be pushed to KV.
If the flag `-sris-only` is set, only SRIs are pushed to KV.
If the flag `-files-only` is set, only files are pushed to KV.
If the flag `-count` is set, the the count of KV keys that should be in KV will be outputted at the end of the program. This will assume all entries can fit into KV (<= 10MiB).
If the flag `-no-push` is set, nothing will be written to KV. However, theoretical keys will be counted if the `-count` flag is set.
If the flag `-panic-oversized` is set, the program will panic if any KV compressed file is oversized (> 10MiB). Note that the program will already panic for oversized entries in other namespaces.

```
make kv && ./bin/kv upload jquery mathjax font-awesome
```

## `upload-version`

Insert a specific package version from disk to KV. Package files and version metadata will be pushed to KV.
If the flag `-meta-only` is set, only version metadata will be pushed to KV.
If the flag `-sris-only` is set, only SRIs are pushed to KV.
If the flag `-files-only` is set, only files are pushed to KV.
If the flag `-count` is set, the the count of KV keys that should be in KV will be outputted at the end of the program. This will assume all entries can fit into KV (<= 10MiB).
If the flag `-no-push` is set, nothing will be written to KV. However, theoretical keys will be counted if the `-count` flag is set.

```
make kv && ./bin/kv upload-version jquery 3.5.1
```

## `upload-aggregate`

Inserts aggregate metadata to KV from scratch by scraping KV entries for package-level and version-specific metadata.

```
make kv && ./bin/kv upload-aggregate jquery mathjax font-awesome
```

## `packages`

Lists all packages in KV.

## `aggregate-packages`

Lists all packages with aggregated metadata in KV. To check each package in KV has an entry for aggregated metadata:

```
unset DEBUG && make kv && diff <(./bin/kv aggregated-packages) <(./bin/kv packages)
```

## `file`

Gets a file from KV using its KV key.
If the flag `-ungzip` is set, the content will be ungzipped.
If the flag `-unbrotli` is set, the content will be unbrotlied.
These two flags are mutually exclusive.

```
make kv && ./bin/kv -ungzip file jquery/3.5.1/jquery.min.js.gz
```

## `files`

Gets the file names stored in KV for a package.

```
make kv && ./bin/kv files jquery
```

## `meta`

Gets all metadata associated with a package in KV.

```
make kv && ./bin/kv meta jquery
```

## `aggregate`

Gets the aggregated metadata associated with a package in KV.

```
make kv && ./bin/kv aggregate jquery
```

## `sris`

Lists all SRIs for files starting with a prefix.

```
make kv && ./bin/kv sris a-happy-tyler
```

```
make kv && ./bin/kv sris a-happy-tyler/1.0.0
```

```
make kv && ./bin/kv sris a-happy-tyler/1.0.0/happy.js
```
