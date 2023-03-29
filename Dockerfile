ARG BUILD_IMAGE=golang:1.20-alpine

FROM $BUILD_IMAGE as build
WORKDIR /workspace
ENV GOPROXY=https://goproxy.cn
COPY go.mod go.sum /workspace/
RUN go mod download
COPY cmd /workspace/cmd
COPY pkg /workspace/pkg
ARG GOFLAGS
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${GOFLAGS} -o queryexporter cmd/queryexporter/main.go

FROM alpine:3.17
COPY --from=build /workspace/queryexporter /usr/local/bin/queryexporter
ENTRYPOINT [ "/usr/local/bin/queryexporter" ]