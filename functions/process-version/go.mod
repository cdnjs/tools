module github.com/cdnjs/tools/functions/process-version

go 1.16

require (
	cloud.google.com/go v0.81.0
	cloud.google.com/go/pubsub v1.10.3
	cloud.google.com/go/storage v1.15.0
	github.com/cdnjs/tools v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	google.golang.org/genproto v0.0.0-20210423144448-3a41ef94ed2b
)

replace github.com/cdnjs/tools => ../../
