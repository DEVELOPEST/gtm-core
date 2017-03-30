BINARY=gtm
VERSION=v1.2.8-beta

LDFLAGS=-ldflags "-X main.Version=${VERSION}"

build:
	go build ${LDFLAGS} -o ${BINARY}

test:
	go test $$(go list ./... | grep -v vendor)

vet:
	go vet $$(go list ./... | grep -v vendor)

fmt:
	go fmt $$(go list ./... | grep -v vendor)

install:
	go install ${LDFLAGS}

clean:
	go clean

.PHONY: test vet install clean fmt todo note
