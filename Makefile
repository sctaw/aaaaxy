# System properties.
EXE = $(shell go env GOEXE)
SUFFIX = -$(shell go env GOOS)-$(shell go env GOARCH)$(EXE)

# Internal variables.
PACKAGE = github.com/divVerent/aaaaaa/cmd/aaaaaa
DUMPCPS = github.com/divVerent/aaaaaa/cmd/dumpcps
DEBUG = aaaaaa-debug$(SUFFIX)
DEBUG_GOFLAGS =
RELEASE = aaaaaa$(SUFFIX)
RELEASE_GOFLAGS = -ldflags="-s -w" -gcflags="-B -dwarf=false"
UPXFLAGS = -9
SOURCES = $(shell find . -name \*.go)
GENERATED_ASSETS = assets/maps/level.cp.json
STATIK_ASSETS_ROOT = internal/assets
STATIK_ASSETS = $(STATIK_ASSETS_ROOT)/statik/statik.go

.PHONY: default
default: debug

.PHONY: all
all: debug release

.PHONY: debug
debug: $(DEBUG)

.PHONY: release
release: $(RELEASE)

.PHONY: clean
clean:
	$(RM) -r $(DEBUG) $(RELEASE) $(STATIK_ASSETS) $(GENERATED_ASSETS)

.PHONY: vet
vet:
	go vet `find ./cmd ./internal -name \*.go -print | sed -e 's,/[^/]*$$,,' | sort -u`

.PHONY: $(STATIK_ASSETS)
$(STATIK_ASSETS): $(GENERATED_ASSETS)
	GOOS= GOARCH= ./statik-vfs.sh $(STATIK_ASSETS_ROOT)

$(DEBUG): $(GENERATED_ASSETS) $(SOURCES)
	go build -o $(DEBUG) $(DEBUG_GOFLAGS) $(PACKAGE)

$(RELEASE): $(STATIK_ASSETS) $(SOURCES)
	go build -tags statik -o $(RELEASE) $(RELEASE_GOFLAGS) $(PACKAGE)

%.cp.json: %.cp.dot
	neato -Tjson $< > $@

%.cp.pdf: %.cp.dot
	neato -Tpdf $< > $@

%.cp.dot: %.tmx cmd/dumpcps/main.go
	GOOS= GOARCH= go run $(DUMPCPS) $< > $@

# Building of release zip files starts here.
ZIPFILE = aaaaaa.zip

.PHONY: addrelease
addrelease: $(RELEASE)
	zip -9r $(ZIPFILE) $(RELEASE)
	$(MAKE) clean

.PHONY: allrelease
allrelease: allreleaseclean
	$(RM) $(ZIPFILE)
	GOOS=linux GOARCH=amd64 $(MAKE) addrelease
	# Disabled due to Windows Defender FP:
	# GOOS=windows GOARCH=386 $(MAKE) release
	GOOS=windows GOARCH=amd64 $(MAKE) addrelease

.PHONY: allreleaseclean
allreleaseclean:
	GOOS=linux GOARCH=amd64 $(MAKE) clean
	GOOS=windows GOARCH=amd64 $(MAKE) clean
	$(RM) $(ZIPFILE)

# Helper targets.
.PHONY: run
run: $(DEBUG)
	./$(DEBUG) $(ARGS)

.PHONY: setup-git
setup-git:
	git config filter.git-clean-tmx.clean "$$PWD"/git-clean-tmx.sh

