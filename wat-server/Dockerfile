#
# Note this image needs to be built from the main wat-app dir with the command
#   docker build -f wat-server/Dockerfile .
#
# The reason for this is that it needs access to the vendor directory.
# and you cant use COPY with paths like ../vendor
#

FROM golang:1.8

RUN go get -u github.com/kardianos/govendor

RUN mkdir -p /go/src/github.com/joeledstrom/wat-app/vendor

WORKDIR /go/src/github.com/joeledstrom/wat-app

COPY vendor/vendor.json vendor

RUN govendor sync

COPY wat-api wat-api
COPY wat-server wat-server

WORKDIR wat-server

RUN go build

ENTRYPOINT ["./wat-server"]
