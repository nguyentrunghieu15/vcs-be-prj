FROM golang:1.22 as build

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
RUN go build -o /go/app/server ./cmd/server

FROM alpine:3.19
COPY --from=build /go/app/server /bin/server
RUN apk add libc6-compat
RUN mkdir /bin/log
RUN mkdir /bin/log/server
CMD [ "/bin/server" ]