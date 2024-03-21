# RumbleRSS

Simple webservice that takes a Rumble.com channel URL and returns an RSS feed containing a list of videos from the channel.

## Example docker-compose Usage
```
git clone https://github.com/porjo/rumblerss.git
cd rumblerss
docker-compose up -d
```

## Example Docker Usage

```
docker pull ghcr.io/porjo/rumblerss:latest

docker run -d -p 8080:8080 ghcr.io/porjo/rumblerss

curl localhost:8080?link=https://rumble.com/mychannel
```

## CORS

[Cross-Origin Resource Sharing (CORS)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)  is supported via the `cors-origins` command line flag. Supply a comma separated list of CORS origins or '*' for all origins e.g.
```
docker run -d ghcr.io/porjo/rumblerss -cors-origins '*'
```