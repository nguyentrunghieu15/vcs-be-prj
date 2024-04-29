FROM golang:1.22

WORKDIR /go/app

COPY ./cmd/gateway ./cmd/gateway
COPY ./pkg/gateway ./pkg/gateway
COPY ./pkg/env ./pkg/env
COPY ./pkg/auth ./pkg/auth
COPY ./pkg/logger ./pkg/logger
COPY ./go.mod .

RUN apt-get update
RUN go mod tidy
RUN mkdir log
RUN mkdir log/gateway
RUN mkdir file
RUN mkdir file/gateway
RUN go build ./cmd/gateway

ENTRYPOINT [ "./gateway"]