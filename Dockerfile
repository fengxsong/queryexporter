ARG BUILD_IMAGE=golang:1.20-alpine

FROM $BUILD_IMAGE as build
ARG TARGETOS
ARG TARGETARCH
ARG LDFLAGS

WORKDIR /workspace
ENV GOPROXY=https://goproxy.cn
COPY go.mod go.sum /workspace/
RUN go mod download
COPY cmd /workspace/cmd
COPY pkg /workspace/pkg
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -ldflags "${LDFLAGS}" -o queryexporter cmd/queryexporter/main.go

FROM alpine:3.17
COPY --from=build /workspace/queryexporter /usr/local/bin/queryexporter
ENTRYPOINT [ "/usr/local/bin/queryexporter" ]