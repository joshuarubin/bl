FROM golang:alpine as build
RUN apk add --no-cache git gcc musl-dev
COPY . /go/src/jrubin.io/bl
WORKDIR /go/src/jrubin.io/bl
ENV GO111MODULE on
RUN go build -v

FROM alpine:latest
MAINTAINER Joshua Rubin <joshua@rubixconsulting.com>
ENTRYPOINT ["bl"]
RUN apk add --no-cache ca-certificates
COPY --from=build /go/src/jrubin.io/bl/bl /usr/local/bin/bl
