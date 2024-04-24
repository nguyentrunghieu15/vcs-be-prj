FROM golang:1.22

WORKDIR /go/app

COPY ./cmd/user ./cmd/user
COPY ./pkg/env ./pkg/env
COPY ./pkg/user ./pkg/user
COPY ./pkg/auth ./pkg/auth
COPY ./pkg/logger ./pkg/logger
COPY go.mod .

RUN apt-get update
RUN go mod tidy
RUN mkdir log
RUN mkdir log/user
RUN go build ./cmd/user

ENTRYPOINT [ "./user"]