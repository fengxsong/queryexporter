APP_NAME ?= queryexporter

BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_USER = $(shell whoami)
REVISION = $(shell git describe --dirty --tags --always)
BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
VERSION ?= unknown

BUILD_PATH = cmd/queryexporter/main.go
OUTPUT_PATH = build/_output/bin

LDFLAGS := -s -X github.com/prometheus/common/version.Version=${VERSION} \
	-X github.com/prometheus/common/version.Revision=${REVISION} \
	-X github.com/prometheus/common/version.Branch=${BRANCH} \
	-X github.com/prometheus/common/version.BuildUser=${BUILD_DATE} \
	-X github.com/prometheus/common/version.BuildDate=${BUILD_USER}

IMAGE_REPO ?= fengxsong/${APP_NAME}
IMAGE_TAG ?= ${REVISION}
IMAGE := ${IMAGE_REPO}:${IMAGE_TAG}

tidy:
	go mod tidy

vendor: tidy
	go mod vendor

mkdir:
	mkdir -p ${OUTPUT_PATH}

bin: mkdir
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags "${LDFLAGS}" -o ${OUTPUT_PATH}/${APP_NAME} ${BUILD_PATH} || exit 1

linux-bin: mkdir
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags "${LDFLAGS}" -o ${OUTPUT_PATH}/${APP_NAME}-linux-amd64 ${BUILD_PATH} || exit 1

upx: bin
	upx ${OUTPUT_PATH}/${APP_NAME}

# Build the docker image
docker-build:
	docker build --rm --build-arg=LDFLAGS="${LDFLAGS}" -t ${IMAGE} -t ${IMAGE_REPO}:latest -f Dockerfile .

# Push the docker image
docker-push:
	docker push ${IMAGE}