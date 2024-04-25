FROM golang:1.22

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
RUN go build ./cmd/file_server

ENTRYPOINT [ "./file_server"]