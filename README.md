# tools

Repository that contains various tools for maintaining cdnjs.

## Tools

- [algolia](./cmd/algolia)
- [checker](./cmd/checker)
- [packages](./cmd/packages)
- [autoupdate](./cmd/autoupdate)

## Configuration

- `DEBUG`: pass true to run in debug mode
- `BOT_BASE_PATH`: cdnjs home
- `SENTRY_DSN` sentry data source name (DSN)
- `WORKERS_KV_NAMESPACE_ID` workers kv namespace ID
- `WORKERS_KV_ACCOUNT_ID` workers kv account ID
- `WORKERS_KV_API_KEY` workers kv api key
- `WORKERS_KV_EMAIL` user@cloudflare.com

## Setup a local environment

All the tools uses `BOT_BASE_PATH` to define a "cdnjs home".

We are going to create the home at `/tmp/cdnjs` and do the following in the directory:
- `git clone https://github.com/cdnjs/packages.git`
- `git clone https://github.com/cdnjs/glob.git`
- `mkdir -p /tmp/cdnjs/cdnjs/ajax/libs` (fake the cdnjs/cdnjs repo)

In glob run `npm install`.

Finally pass the `BOT_BASE_PATH` to the tool, for example: `BOT_BASE_PATH=/tmp/cdnjs bin/autoupdate -no-update`.

