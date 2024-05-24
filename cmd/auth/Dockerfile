FROM golang:1.22 as build

WORKDIR /go/app

COPY ./cmd/auth ./cmd/auth
COPY ./pkg/auth ./pkg/auth
COPY ./pkg/env ./pkg/env
COPY ./pkg/logger ./pkg/logger
COPY ./go.mod .

RUN apt-get update
RUN mkdir log
RUN mkdir log/auth
RUN go mod tidy
RUN go build -o /go/app/auth ./cmd/auth

FROM alpine:3.19
COPY --from=build /go/app/auth /bin/auth
RUN apk add libc6-compat
RUN mkdir /log
RUN mkdir /log/auth
CMD [ "/bin/auth" ]