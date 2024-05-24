FROM golang:1.22 as build

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
RUN go build -o /go/app/exporter ./cmd/exporter

FROM alpine:3.19
COPY --from=build /go/app/exporter /bin/exporter
RUN apk add libc6-compat
CMD [ "/bin/exporter"]