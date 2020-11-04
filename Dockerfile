# Build Geth in a stock Go builder container
FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ADD . /XDCchain
RUN cd /XDCchain && make XDC

FROM alpine:latest

LABEL maintainer="anil@xinfin.org"

EXPOSE 8545 8546 8547 30303 30303/udp
ENTRYPOINT ["XDC"]
