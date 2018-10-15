# bl

## Build and Run

Build requires Go `v1.11` or later

```sh
git clone https://github.com/joshuarubin/bl
cd bl
```

### Local

```sh
GO111MODULE=on go build -v
./bl
```

### Docker

```
docker build -t joshuarubin/bl:latest .
docker run --rm -it -p 443:443 joshuarubin/bl:latest
```

### Runtime Flags

```
  -addr string
        address:port to listen for requests on (default ":https")
  -cert string
        tls certificate file (optional)
  -key string
        tls key file (optional)
  -workers int
        number of worker connections to maintain to the bitly api (default 16)
```

## Querying the Server

```
curl \
  --insecure \
  --header "Authorization: Bearer ${BITLY_ACCESS_TOKEN}" \
  https://localhost/v1/clicks/country | jq
```

### Request Parameters

* `unit`: a unit of time, enum: `”minute”`, `”hour”`, `”day”`, `”week”`, `”month”` (default “day”)
* `units`: an integer representing the time units to query data for (default “30”)
* `size`: the quantity of items to be returned (default “10”)
* `page`: integer specifying the numbered result at which to start (default “1”)
