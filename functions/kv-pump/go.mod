module github.com/cdnjs/tools/functions/kv-pump

go 1.16

require (
	cloud.google.com/go v0.81.0
	cloud.google.com/go/storage v1.15.0
	github.com/agnivade/levenshtein v1.1.0 // indirect
	github.com/cdnjs/tools v0.0.0-00010101000000-000000000000
	github.com/cloudflare/cloudflare-go v0.16.0
	github.com/dlclark/regexp2 v1.2.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/pachyderm/ohmyglob v0.0.0-20190808212558-a8e61fd76805 // indirect
	github.com/pkg/errors v0.9.1
	golang.org/x/oauth2 v0.0.0-20210413134643-5e61552d6c78 // indirect
	google.golang.org/api v0.45.0 // indirect
	google.golang.org/grpc v1.37.0 // indirect
)

replace github.com/cdnjs/tools => ../../
