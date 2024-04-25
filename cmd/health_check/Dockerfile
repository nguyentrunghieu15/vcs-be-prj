FROM golang:1.22

WORKDIR /go/app

COPY ./cmd/health_check ./cmd/health_check
COPY ./pkg/health_check ./pkg/health_check
COPY ./pkg/env ./pkg/env
COPY ./pkg/server ./pkg/server
COPY ./pkg/auth ./pkg/auth
COPY ./pkg/redis ./pkg/redis
COPY ./pkg/cache ./pkg/cache
COPY ./pkg/logger ./pkg/logger
COPY ./go.mod .

RUN apt-get update
RUN go mod tidy
RUN mkdir log
RUN mkdir log/health_check
RUN go build ./cmd/health_check

ENTRYPOINT [ "./health_check"]