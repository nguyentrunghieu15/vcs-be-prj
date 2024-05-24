FROM golang:1.22 as build

WORKDIR /go/app

COPY ./cmd/mail_sender ./cmd/mail_sender
COPY ./pkg/mail_sender ./pkg/mail_sender
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
RUN mkdir log/mail_sender
RUN go build -o /go/app/mail_sender ./cmd/mail_sender

FROM alpine:3.19
COPY --from=build /go/app/mail_sender /bin/mail_sender
RUN apk add libc6-compat
RUN mkdir /bin/log
RUN mkdir /bin/log/mail_sender
CMD [ "/bin/mail_sender" ]