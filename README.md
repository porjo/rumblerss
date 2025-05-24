# RumbleRSS

Simple webservice that takes a Rumble.com channel URL and returns an RSS feed containing a list of videos from the channel.

## Usage

```
$ ./rumblerss -h
Usage of ./rumblerss:
  -debug
        debug log output
  -maxItemCount int
        limit the maximum number of feed items returned (zero is unlimited)
  -maxTextLength int
        limit each field to maximum number of characters (zero is unlimited)
  -port int
        listen on this port (default 8080)
```


## Docker

Docker image available at: `ghcr.io/porjo/rumblerss:latest`

**Example docker-compose Usage**

```
docker-compose up -d

curl localhost:8080?link=https://rumble.com/mychannel
```

**Example Docker Usage**

```
docker pull ghcr.io/porjo/rumblerss:latest

docker run -d -p 8080:8080 ghcr.io/porjo/rumblerss

curl localhost:8080?link=https://rumble.com/mychannel
```

## CORS

[Cross-Origin Resource Sharing (CORS)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS) should be handled by putting rumblerss behind a reverse proxy that can add the necessary header e.g. Caddy2
```
        route /rumblerss/* {
                header ?Access-Control-Allow-Origin https://porjo.github.io
                uri strip_prefix /rumblerss
                reverse_proxy 127.0.0.1:8080
        }
```
