SHELL := /bin/bash

COMMIT_SHA := $(shell git rev-parse HEAD)

tools:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/mitchellh/gox
	go get -u github.com/tcnksm/ghr

release:
	gox -arch="amd64" -os="linux windows darwin" \
	    -output "dist/aws-reverse-proxy_{{.OS}}_{{.Arch}}" ./aws-reverse-proxy/
	gox -arch="amd64" -os="linux windows darwin" \
	    -output "dist/aws-curl_{{.OS}}_{{.Arch}}" ./aws-curl/
	ghr -t $$GITHUB_TOKEN -u BSick7 -r aws-signing -c $(COMMIT_SHA) --replace `cat VERSION` dist/
