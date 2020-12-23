# Autoupdate

## -no-update
If the flag is set, the autoupdater will not commit or push to git or write to Workers KV. This is used for local testing purposes.

## -no-pull
If the flag is set, the autoupdater will not pull from git.

## -package
Run the autoupdate for a specific package.
Usage: `autoupdate -package=h/hi-sven.json`

The path of the file must be in the `packages/packages` repo.
