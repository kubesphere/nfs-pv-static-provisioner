REPO ?= kubesphere
TAG ?= latest

build-local: ; $(info $(M)...Begin to build  binary.)  @ ## Build .
	CGO_ENABLED=0 go build -ldflags \
	"-X 'main.goVersion=$(shell go version|sed 's/go version //g')' \
	-X 'main.gitHash=$(shell git describe --dirty --always --tags)' \
	-X 'main.buildTime=$(shell TZ=UTC-8 date +%Y-%m-%d" "%H:%M:%S)'" \
	-o bin/manager cmd/main.go

build-image: ; $(info $(M)...Begin to build  image.)  @ ## Build  image.
	docker build -f Dockerfile -t ${REPO}/nfs-pv-static-provisioner:${TAG}  .
	docker push ${REPO}/nfs-pv-static-provisioner:${TAG}

build-cross-image: ; $(info $(M)...Begin to build  image.)  @ ## Build  image.
	docker buildx build -f Dockerfile -t ${REPO}/nfs-pv-static-provisioner:${TAG} --push --platform linux/amd64,linux/arm64 .