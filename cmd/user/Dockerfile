FROM golang:1.22 as build

WORKDIR /go/app

COPY ./cmd/user ./cmd/user
COPY ./pkg/env ./pkg/env
COPY ./pkg/user ./pkg/user
COPY ./pkg/auth ./pkg/auth
COPY ./pkg/logger ./pkg/logger
COPY ./go.mod .

RUN apt-get update
RUN go mod tidy
RUN mkdir log
RUN mkdir log/user
RUN go build -o /go/app/user ./cmd/user

FROM alpine:3.19
COPY --from=build /go/app/user /bin/user
RUN apk add libc6-compat
RUN mkdir /log
RUN mkdir /log/user
CMD [ "/bin/user" ]