FROM golang:1.22 as build

WORKDIR /go/app

COPY ./cmd/file_server ./cmd/file_server
COPY ./pkg/file_server ./pkg/file_server
COPY ./pkg/env ./pkg/env
COPY ./go.mod .

RUN apt-get update
RUN go mod tidy
RUN mkdir log
RUN mkdir log/file_server
RUN mkdir file
RUN mkdir file/storage
RUN go build -o /go/app/file_server ./cmd/file_server

FROM alpine:3.19
COPY --from=build /go/app/file_server /bin/file_server
RUN apk add libc6-compat
CMD [ "/bin/file_server"]