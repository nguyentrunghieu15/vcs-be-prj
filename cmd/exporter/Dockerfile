FROM golang:1.22

WORKDIR /go/app

COPY ./cmd/exporter ./cmd/exporter
COPY ./pkg/exporter ./pkg/exporter
COPY ./pkg/env ./pkg/env
COPY ./pkg/logger ./pkg/logger
COPY ./pkg/server ./pkg/server
COPY ./pkg/redis ./pkg/redis
COPY ./pkg/auth ./pkg/auth
COPY ./pkg/cache ./pkg/cache
COPY ./go.mod .

RUN apt-get update
RUN go mod tidy
RUN mkdir log
RUN mkdir log/exporter
RUN mkdir file
RUN mkdir file/exporter
RUN go build ./cmd/exporter

ENTRYPOINT [ "./exporter"]