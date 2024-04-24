FROM golang:1.22

WORKDIR /go/app

COPY ./cmd/auth ./cmd/auth
COPY ./pkg/auth ./pkg/auth
COPY ./pkg/env ./pkg/env
COPY ./pkg/logger ./pkg/logger
COPY go.mod .

RUN apt-get update
RUN mkdir log
RUN mkdir log/auth
RUN go mod tidy
RUN go build ./cmd/auth

ENTRYPOINT [ "./auth"]