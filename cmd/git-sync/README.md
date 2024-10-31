## Build

Build git-sync:
```
cd ./cmd/git-sync
go build
```

## Usage

Write last update marker file (can be found at https://github.com/cdnjs/cdnjs/blob/master/last-sync):
```
echo "2024-09-08T01:33:46.508Z" > /tmp/last-sync
```

Run git-sync:
```
DEBUG=1 PUSH_FREQ=0 ./git-sync /tmp/last-sync cdnjs-outgoing-prod
```
