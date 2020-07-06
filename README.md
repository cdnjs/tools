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

## Dependencies

- [brotli](https://github.com/google/brotli)
- [gzip](https://www.gzip.org/)
- [zopflipng](https://github.com/google/zopfli)

## Setup a local environment

All the tools uses `BOT_BASE_PATH` to define a "cdnjs home".

We are going to create the home at `/tmp/cdnjs` and do the following in the directory:
- `git clone https://github.com/cdnjs/packages.git`
- `git clone https://github.com/cdnjs/glob.git`
- `mkdir -p /tmp/cdnjs/cdnjs/ajax/libs` (fake the cdnjs/cdnjs repo)

In glob run `npm install`.

Finally pass the `BOT_BASE_PATH` to the tool, for example: `BOT_BASE_PATH=/tmp/cdnjs bin/autoupdate -no-update`.

