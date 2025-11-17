# syntax=docker/dockerfile:1

FROM golang:1.25-alpine

ENV USER user
ENV HOME /home/${USER}

RUN apk add --update sudo
RUN adduser -D $USER \
        && echo "$USER ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/$USER \
        && chmod 0440 /etc/sudoers.d/$USER
USER $USER

WORKDIR /proto-build

RUN sudo apk add build-base
RUN sudo apk add protobuf-dev
RUN sudo apk add protoc

RUN go install github.com/golang/protobuf/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
RUN go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
