FROM golang:1.22

WORKDIR /go/app

COPY ./cmd/server ./cmd/server
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
RUN mkdir log/server
RUN mkdir file
RUN mkdir file/server
RUN go build ./cmd/server

ENTRYPOINT [ "./server"]