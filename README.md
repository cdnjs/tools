# tools

Repository that contains various tools for maintaining cdnjs.

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
- `DISABLE_KV` set to disable writes to workers KV
- `WORKERS_KV_NAMESPACE_ID` workers kv namespace ID
- `WORKERS_KV_ACCOUNT_ID` workers kv account ID
- `WORKERS_KV_API_TOKEN` workers kv api token

## Dependencies

In `tools/` run `npm install`.

- [jpegoptim](https://www.kokkonen.net/tjko/projects.html)
- [zopflipng](https://github.com/google/zopfli)

## Setup a local environment

All the tools uses `BOT_BASE_PATH` to define a "cdnjs home".

We are going to create the home at `/tmp/cdnjs` and do the following in the directory:
- `git clone https://github.com/cdnjs/packages.git`
- `git clone https://github.com/cdnjs/glob.git`
- `mkdir -p /tmp/cdnjs/cdnjs/ajax/libs` (fake the cdnjs/cdnjs repo)

In glob run `npm install`.

Finally pass the `BOT_BASE_PATH` to the tool, for example: `BOT_BASE_PATH=/tmp/cdnjs bin/autoupdate -no-update`.
