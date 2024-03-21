# Build stage
FROM golang:alpine AS build-env

COPY . /etc
WORKDIR /etc

RUN apk update && \
    apk upgrade && \
	apk add git

RUN go build -o rumblerss

# Final stage
FROM alpine

RUN apk update && apk upgrade

WORKDIR /etc
COPY --from=build-env /etc/rumblerss /etc/

ENTRYPOINT ["/etc/rumblerss"]