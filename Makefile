TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=terraform.local
NAMESPACE=bytebase
NAME=bytebase
BINARY=terraform-provider-${NAMESPACE}
VERSION=${shell cat ./VERSION}
OS_ARCH?=$(shell go env GOOS)_$(shell go env GOARCH)
PLUGIN_DIR=~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
PLUGIN_BINARY=${BINARY}_v${VERSION}

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install: build
	mkdir -p ${PLUGIN_DIR}
	rm -f ${PLUGIN_DIR}/${BINARY}
	mv ${BINARY} ${PLUGIN_DIR}/${PLUGIN_BINARY}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m
