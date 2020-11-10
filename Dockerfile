FROM golang:1.10-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /XDCchain
RUN cd /XDCchain && make XDC

FROM alpine:latest

LABEL maintainer="anil@xinfin.org"

RUN addgroup -g 1000 XDC && \
    adduser -h /root -D -u 1000 -G XDC XDC && \
    chown XDC:XDC /root

USER XDC

EXPOSE 8545 8546 30303 30303/udp 30304/udp
ENTRYPOINT ["XDC"]
