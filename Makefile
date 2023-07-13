VERSION ?= latest
REPO ?= coredgeio

CONFIGMAP_REACTOR_IMG ?= $(REPO)/configmap-reactor:$(VERSION)

.PHONY: all

all: build-configmap-reactor

build-configmap-reactor: go-format go-vet
	sudo docker build -t ${CONFIGMAP_REACTOR_IMG} -f build/Dockerfile .

push-images:
	sudo docker push $(CONFIGMAP_REACTOR_IMG)

go-format:
	go fmt ./...

go-vet:
	go vet ./...
