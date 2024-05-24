FROM golang:1.22 as build

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
RUN go build -o /go/app/health_check ./cmd/health_check


FROM alpine:3.19
COPY --from=build /go/app/health_check /bin/health_check
RUN apk add libc6-compat
RUN mkdir /log
RUN mkdir /log/health_check
CMD [ "/bin/health_check"]