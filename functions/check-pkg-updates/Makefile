check-pkg-updates.zip:
	go build
	go mod vendor
	zip -r $@ . -x 'go.*'
