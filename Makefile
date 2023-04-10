#

SHELL := /bin/bash
INTERACTIVE := $(shell [ -t 0 ] && echo 1)

root_mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
export REPO_ROOT_DIR := $(realpath $(dir $(root_mkfile_path)))

export PROJECT_NAME ?= lazygpt
export DOCKER_REPOSITORY ?= ghcr.io/lazygpt/lazygpt
export DOCKER_DEVKIT_IMG ?= $(DOCKER_REPOSITORY):latest-devkit
export DOCKER_DEVKIT_PHONY_FILE ?= .docker-$(shell echo '$(DOCKER_DEVKIT_IMG)' | tr '/:' '.')

export DOCKER_DEVKIT_GITHUB_ARGS ?= \
	--env CI \
	--env-file <(env | grep GITHUB_)

export DOCKER_DEVKIT_GO_ARGS ?= \
	--env GOCACHE=/code/.cache/go/go-build \
	--env GOMODCACHE=/code/.cache/go/pkg/mod \
	--env GOLANGCI_LINT_CACHE=/code/.cache/golangci-lint \
	--env GOPRIVATE=github.com/floatme-corp

export DOCKER_DEVKIT_ARGS ?= \
	--rm \
	$(if $(INTERACTIVE),--tty) \
	--interactive \
	--env DEVKIT=true \
	--volume $(REPO_ROOT_DIR):/code:Z \
	--workdir /code \
	$(if $(shell docker --version | grep -v podman),--user "$(shell id -u):$(shell id -g)") \
	$(if $(shell docker --version | grep -v podman),--env HOME=/code) \
	$(DOCKER_DEVKIT_GITHUB_ARGS) \
	$(DOCKER_DEVKIT_GO_ARGS)

# NOTE(jkoelker) Comma must not appear in a funciton call, use a variable
#                as suggested by the documentation.
#  https://www.gnu.org/software/make/manual/html_node/Syntax-of-Functions.html
comma := ,
export DOCKER_DEVKIT_BUILDX_ARGS ?= \
	$(if $(GITHUB_ACTIONS),--cache-from type=gha) \
	$(if $(GITHUB_ACTIONS),--cache-to type=gha$(comma)mode=max)

DIST_DIR ?= $(REPO_ROOT_DIR)/dist
SOURCES := $(shell find $(REPO_ROOT_DIR) -name 'main.go')
COMMANDS := $(patsubst cmd/%/main.go,%,$(SOURCES))
DIST_BINARIES := $(addprefix $(DIST_DIR)/, $(COMMANDS))
DIST_ZIP_BINARIES := $(addsuffix .zip,$(DIST_BINARIES))

GO_MODULE_PATH := $(shell go list -m)

export GO_LDFLAGS := \
	-v \
	-s \
	-w \

# NOTE(jkoelker) Abuse ifeq and the junk variable to proxy docker image state
#                to the target file
ifneq ($(shell command -v docker),)
    ifeq ($(shell docker image ls --quiet "$(DOCKER_DEVKIT_IMG)"),)
        export junk := $(shell rm -rf $(DOCKER_DEVKIT_PHONY_FILE))
    endif
endif

$(DIST_DIR):
	mkdir -p $(DIST_DIR)


$(DIST_DIR)/lazygpt: generate.host
$(DIST_DIR)/lazygpt: $(DIST_DIR)
$(DIST_DIR)/lazygpt: $(shell find $(REPO_ROOT_DIR)/cmd -type f -name '*'.go)
$(DIST_DIR)/lazygpt: $(shell find $(REPO_ROOT_DIR)/pkg -type f -name '*'.go)
$(DIST_DIR)/lazygpt: $(shell find $(REPO_ROOT_DIR)/plugin/api -type f -name '*'.go)
	@echo "go building $(basename $@)"
	go build -ldflags="$(GO_LDFLAGS)" -o $@ cmd/lazygpt/lazygpt.go
	@echo

$(DIST_DIR)/lazygpt-plugin-openai: $(DIST_DIR)
$(DIST_DIR)/lazygpt-plugin-openai: $(shell find $(REPO_ROOT_DIR)/plugin -type f -name '*'.go)
	@echo "go building $(basename $@)"
	cd plugin/openai && go build -ldflags="$(GO_LDFLAGS)" -o $@ cmd/main.go
	@echo

$(DOCKER_DEVKIT_PHONY_FILE): Dockerfile.devkit
	docker buildx build \
		$(if $(shell docker --version | grep -v podman),--output=type=docker) \
		$(DOCKER_DEVKIT_BUILDX_ARGS) \
		--file $(REPO_ROOT_DIR)/Dockerfile.devkit \
		--tag "$(DOCKER_DEVKIT_IMG)" \
		$(REPO_ROOT_DIR) \
	&& touch $(DOCKER_DEVKIT_PHONY_FILE)

.cache/go/pkg/mod:
	mkdir -p $(REPO_ROOT_DIR)/.cache/go/pkg/mod

.cache/go/go-build:
	mkdir $(REPO_ROOT_DIR)/.cache/go/go-build

.cache/golangci-lint:
	mkdir $(REPO_ROOT_DIR)/.cache/golangci-lint

.cache/terraform/plugins:
	mkdir -p $(REPO_ROOT_DIR)/.cache/terraform/plugins

.PHONY: go-cache
go-cache: .cache/go/pkg/mod
go-cache: .cache/go/go-build
go-cache: .cache/golangci-lint

.PHONY: repo-cache
repo-cache: go-cache

.PHONY: devkit
devkit: $(DOCKER_DEVKIT_PHONY_FILE)
devkit: repo-cache

WHAT ?= /bin/bash
.PHONY: devkit.run
devkit.run: devkit
	docker run \
		$(DOCKER_DEVKIT_ARGS) \
		"$(DOCKER_DEVKIT_IMG)" \
		/bin/bash -c 'git config --global safe.directory /code && $(WHAT)'

.PHONY: dev
dev: devkit.run

.PHONY: shell
shell: devkit.run

.PHONY: build.host
build.host: generate.host
build.host: $(DIST_DIR)/lazygpt
build.host: $(DIST_DIR)/lazygpt-plugin-openai

.PHONY: build
build: WHAT=make build.host
build: devkit.run

.PHONY: lint.host-docker
lint.host-docker:
	@echo "Linting Dockerfiles"
	hadolint Dockerfile*
	@echo

.PHONY: lint.host-go
lint.host-go: generate.host
	@echo "Linting Go files"
	golangci-lint run --verbose
	@echo

.PHONY: lint.host
lint.host: lint.host-docker
lint.host: lint.host-go

.PHONY: lint
lint: WHAT=make lint.host
lint: devkit.run

.PHONY: clean-plugin
clean-plugin:
	rm -f plugin/api/*.pb.go

.PHONY: clean-coverage
clean-coverage:
	rm -f coverage.out coverage.xml coverage.html

.PHONY: clean-dist
clean-dist:
	rm -rf $(DIST_DIR)

.PHONY: clean
clean: clean-plugin
clean: clean-coverage
clean: clean-dist

plugin/api/interfaces.pb.go: plugin/api/interfaces.proto
	go generate plugin/api/api.go

plugin/api/interfaces_grpc.pb.go: plugin/api/interfaces.proto
	go generate plugin/api/api.go

.PHONY: generate-plugin.host
generate-plugin.host: plugin/api/interfaces.pb.go
generate-plugin.host: plugin/api/interfaces_grpc.pb.go

.PHONY: generate.host
generate.host: generate-plugin.host

.PHONY: generate
generate: WHAT=make generate.host
generate: devkit.run

coverage.out: $(shell find $(REPO_ROOT_DIR) -type f -name '*'.go)
coverage.out: generate.host
coverage.out:
	gotestsum \
		--junitfile-testsuite-name=relative \
		--junitfile-testcase-classname=short \
		-- \
		-covermode=atomic \
		-coverprofile=coverage.out \
		-race \
		-short \
		-v \
		./...

coverage.html: coverage.out
coverage.html:
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test.host
test.host: generate.host
test.host: coverage.html

.PHONY: test
test: WHAT=make test.host
test: devkit.run
