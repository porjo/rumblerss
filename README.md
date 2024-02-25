# RumbleRSS

Simple webservice that takes a Rumble channel URL and returns an RSS feed containing a list of videos from the channel.

## Example Usage

```
docker pull ghcr.io/porjo/rumblerss:latest

docker run -d ghcr.io/porjo/rumblerss

curl localhost:8080?link=https://rumble.com/c/mychannel/videos
```

URL Parameters:
- link: Rumble channel URL
- title: (optional) title to use for the feed
- description: (optional) description to use for the feed
- publishTime: (optional) publish time for the feed

