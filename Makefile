NAME=csi-lvm
OS ?= linux
ifeq ($(strip $(shell git status --porcelain 2>/dev/null)),)
  GIT_TREE_STATE=clean
else
  GIT_TREE_STATE=dirty
endif
COMMIT ?= $(shell git rev-parse HEAD)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
LDFLAGS ?= -X github.com/metal-pod/csi-lvm/driver.version=${VERSION} -X github.com/metal-pod/csi-lvm/driver.commit=${COMMIT} -X github.com/metal-pod/csi-lvm/driver.gitTreeState=${GIT_TREE_STATE}
PKG ?= github.com/metal-pod/csi-lvm/cmd/do-csi-plugin

VERSION ?= $(shell cat VERSION)
DOCKER_REPO ?= metalpod/csi-lvm

all: test csi-lvm

publish: compile build push clean

.PHONY: bump-version
bump-version:
	@[ "${NEW_VERSION}" ] || ( echo "NEW_VERSION must be set (ex. make NEW_VERSION=v1.x.x bump-version)"; exit 1 )
	@(echo ${NEW_VERSION} | grep -E "^v") || ( echo "NEW_VERSION must be a semver ('v' prefix is required)"; exit 1 )
	@echo "Bumping VERSION from $(VERSION) to $(NEW_VERSION)"
	@echo $(NEW_VERSION) > VERSION
	@cp deploy/kubernetes/releases/csi-lvm-${VERSION}.yaml deploy/kubernetes/releases/csi-lvm-${NEW_VERSION}.yaml
	@sed -i'' -e 's#metalpod/csi-lvm:${VERSION}#metalpod/csi-lvm:${NEW_VERSION}#g' deploy/kubernetes/releases/csi-lvm-${NEW_VERSION}.yaml
	@sed -i'' -e 's/${VERSION}/${NEW_VERSION}/g' README.md
	$(eval NEW_DATE = $(shell date +%Y.%m.%d))
	@sed -i'' -e 's/## unreleased/## ${NEW_VERSION} - ${NEW_DATE}/g' CHANGELOG.md
	@ echo '## unreleased\n' | cat - CHANGELOG.md > temp && mv temp CHANGELOG.md

.PHONY: test
test:
	@echo "==> Testing all packages"
	@go test -v ./...

.PHONY: test-integration
test-integration:
	@echo "==> Started integration tests"
	@env go test -count 1 -v -tags integration ./test/...

.PHONY: csi-lvm
csi-lvm:
	go build -tags netgo -o bin/csi-lvm cmd/csi-lvm/*.go
	strip bin/csi-lvm

.PHONY: docker-image
docker-image:
	@echo "==> Building the docker image"
	@docker build -t $(DOCKER_REPO):$(VERSION) . -f cmd/csi-lvm/Dockerfile

.PHONY: push
push:
ifneq ($(BRANCH),master)
  ifneq ($(VERSION),dev)
	$(error "Only the `dev` tag can be published from non-master branches")
  endif
endif
	@echo "==> Publishing $(DOCKER_REPO):$(VERSION)"
	@docker push $(DOCKER_REPO):$(VERSION)
	@echo "==> Your image is now available at $(DOCKER_REPO):$(VERSION)"

.PHONY: clean
clean:
	@echo "==> Cleaning releases"
	@GOOS=${OS} go clean -i -x ./...
