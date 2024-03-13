# RumbleRSS

Simple webservice that takes a Rumble.com channel URL and returns an RSS feed containing a list of videos from the channel.

## Example Usage

```
docker pull ghcr.io/porjo/rumblerss:latest

docker run -d  -p 8080:8080 ghcr.io/porjo/rumblerss

curl localhost:8080?link=https://rumble.com/c/mychannel/videos
```

## CORS

[Cross-Origin Resource Sharing (CORS)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)  is supported via the `cors-origins` command line flag. Supply a comma separated list of CORS origin domains or '*' for all domains e.g.
```
docker run -d ghcr.io/porjo/rumblerss -cors-origins '*'
```