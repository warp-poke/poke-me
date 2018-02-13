BUILD_DIR=build

CC=go build
GITHASH=$(shell git rev-parse HEAD)
DFLAGS=-race
CFLAGS=-X github.com/warp-poke/poke-me/cmd.githash=$(GITHASH)
CROSS=GOOS=linux GOARCH=amd64

rwildcard=$(foreach d,$(wildcard $1*),$(call rwildcard,$d/,$2) $(filter $(subst *,%,$2),$d))
VPATH= $(BUILD_DIR)

.SECONDEXPANSION:

.PHONY: init
init:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install --update

.PHONY: dep
dep:
	glide install

build: pokeme.go $$(call rwildcard, ./cmd, *.go) $$(call rwildcard, ./core, *.go)
	$(CC) $(DFLAGS) -ldflags "$(CFLAGS)" -o $(BUILD_DIR)/poke-me poke-me.go

.PHONY: release
release: poke-me.go $$(call rwildcard, ./cmd, *.go) $$(call rwildcard, ./core, *.go) $$(call rwildcard, ./catalyser, *.go)
	$(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/poke-me poke-me.go

.PHONY: dist
dist: poke-me.go $$(call rwildcard, ./cmd, *.go) $$(call rwildcard, ./core, *.go) $$(call rwildcard, ./catalyser, *.go)
	$(CROSS) $(CC) -ldflags "-s -w $(CFLAGS)" -o $(BUILD_DIR)/poke-me poke-me.go

.PHONY: lint
lint:
	@command -v gometalinter >/dev/null 2>&1 || { echo >&2 "gometalinter is required but not available please follow instructions from https://github.com/alecthomas/gometalinter"; exit 1; }
	gometalinter --deadline=180s --disable-all --enable=gofmt ./cmd/... ./core/... ./
	gometalinter --deadline=180s --disable-all --enable=vet ./cmd/... ./core/... ./
	gometalinter --deadline=180s --disable-all --enable=golint ./cmd/... ./core/... ./
	gometalinter --deadline=180s --disable-all --enable=ineffassign ./cmd/... ./core/... ./
	gometalinter --deadline=180s --disable-all --enable=misspell ./cmd/... ./core/... ./
	gometalinter --deadline=180s --disable-all --enable=staticcheck ./cmd/... ./core/... ./

.PHONY: format
format:
	gofmt -w -s ./cmd ./core poke-me.go

.PHONY: dev
dev: format lint build

.PHONY: clean
clean:
	-rm -r build