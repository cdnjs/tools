<h1 align="center">
    <a href="https://cdnjs.com"><img src="https://raw.githubusercontent.com/cdnjs/brand/master/logo/standard/dark-512.png" width="175px" alt="< cdnjs >"></a>
</h1>

<h3 align="center">The #1 free and open source CDN built to make life easier for developers.</h3>

---

## Introduction

This repository contains various tools that we use to help with the process of maintaining cdnjs.

## Tools

- [algolia](./cmd/algolia)
- [checker](./cmd/checker)
- [packages](./cmd/packages)
- [autoupdate](./cmd/autoupdate)
- [kv](./cmd/kv)

## Configuration

- `DEBUG`: pass true to run in debug mode
- `BOT_BASE_PATH`: cdnjs home
- `SENTRY_DSN` sentry data source name (DSN)
- `WORKERS_KV_FILES_NAMESPACE_ID` workers kv namespace ID for files
- `WORKERS_KV_SRIS_NAMESPACE_ID` workers kv namespace ID for file SRIs
- `WORKERS_KV_VERSIONS_NAMESPACE_ID` workers kv namespace ID containing metadata for versions
- `WORKERS_KV_PACKAGES_NAMESPACE_ID` workers kv namespace ID containing metadata for packages
- `WORKERS_KV_AGGREGATED_METADATA_NAMESPACE_ID` workers kv namespace ID containing aggregated metadata for packages
- `WORKERS_KV_ACCOUNT_ID` workers kv account ID
- `WORKERS_KV_API_TOKEN` workers kv api token

## Dependencies

In `tools/` run `npm install`.

- [jpegoptim](https://www.kokkonen.net/tjko/projects.html)
- [zopflipng](https://github.com/google/zopfli)
- [brotli](https://github.com/google/brotli) (Linux)

## Local environment

```
$ make dev
$ autoupdate -no-pull -package=h/hi-sven.json
$ ls /cdnjs/cdnjs/ajax/libs/hi-sven
```

## License

Each library hosted on cdnjs is released under its own license. This cdnjs repository is published under [MIT license](LICENSE).
