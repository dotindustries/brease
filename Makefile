
############################# Main targets #############################
ci-build: install openapi

# Install dependencies.
install: grpc-install api-linter-install buf-install

# Run all linters and compile proto files.
openapi: grpc
########################################################################

##### Variables ######
ifndef GOPATH
GOPATH := $(shell go env GOPATH)
endif

PROTO_ROOT := "./proto"
GOBIN := $(if $(shell go env GOBIN),$(shell go env GOBIN),$(GOPATH)/bin)
PATH := $(GOBIN):$(PATH)

COLOR := "\e[1;36m%s\e[0m\n"

##### Compile openapi specification #####
grpc: buf-lint api-linter buf-breaking

##### Plugins & tools #####
grpc-install:
	printf $(COLOR) "Install/update gRPC plugins..."
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest

buf-install:
	printf $(COLOR) "Install/update buf..."
	go install github.com/bufbuild/buf/cmd/buf@v1.6.0

api-linter-install:
	printf $(COLOR) "Install/update api-linter..."
	go install github.com/googleapis/api-linter/cmd/api-linter@v1.32.3

##### Linters #####
api-linter:
	printf $(COLOR) "Run api-linter..."
	api-linter --set-exit-status $(PROTO_IMPORTS) --config $(PROTO_ROOT)/api-linter.yaml $(PROTO_FILES)

buf-lint:
	printf $(COLOR) "Run buf linter..."
	(cd $(PROTO_ROOT) && buf lint)

buf-breaking:
	@printf $(COLOR) "Run buf breaking changes check against master branch..."
	@(cd $(PROTO_ROOT) && buf breaking --against '.git#branch=master')