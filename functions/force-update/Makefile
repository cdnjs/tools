force-update.zip:
	go build
	go mod vendor
	zip -r $@ . -x 'go.*'
