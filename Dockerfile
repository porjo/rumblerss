# Build stage
FROM golang:alpine AS build-env

COPY . /app
WORKDIR /app

RUN apk update && \
    apk upgrade && \
	apk add git

RUN go build -o rumblerss

# Final stage
FROM alpine

RUN apk update && apk upgrade

WORKDIR /app
COPY --from=build-env /app/rumblerss /app/

CMD ["/app/rumblerss"]