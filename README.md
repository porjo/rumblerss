# RumbleRSS

Simple webservice that takes a Rumble.com channel URL and returns an RSS feed containing a list of videos from the channel.

## Example Usage

```
docker pull ghcr.io/porjo/rumblerss:latest

docker run -d ghcr.io/porjo/rumblerss

curl localhost:8080?link=https://rumble.com/c/mychannel/videos
```

## CORS

Cross-Origin Resource Sharing (CORS)  is supported via the `cors-origins` command line flag e.g.
```
docker run -d ghcr.io/porjo/rumblerss -cors-origins '*'
```