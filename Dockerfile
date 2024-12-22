ARG BUILD_IMAGE=golang:1.23-alpine

FROM $BUILD_IMAGE as build
ARG TARGETOS
ARG TARGETARCH
ARG LDFLAGS

WORKDIR /workspace
ENV GOPROXY=https://goproxy.cn
COPY go.mod go.sum /workspace/
RUN go mod download
COPY main.go /workspace/
COPY pkg /workspace/pkg
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -ldflags "${LDFLAGS}" -o queryexporter main.go

FROM alpine:3
COPY --from=build /workspace/queryexporter /usr/local/bin/queryexporter
ENTRYPOINT [ "/usr/local/bin/queryexporter" ]
