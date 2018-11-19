# Request Splitter

Simple tool to split one HTTP request into multiple requests. Incoming requests are forwarded to any number of upstream services.

# Run

Create a `config.json` with your endpoints and redirects.
Run the Docker image and mount the `config.json` in the `config` directory.

```sh
docker run \
    --mount type=bind,source="$(pwd)"/config.json,target=/config/config.json,readonly \
    request-splitter:latest
```

# Configure

The following sample configuration exposes two endpoints, `/metrics` and `/forward`. Each endpoint forwards the request to two upstream services.

```js
{
    "Endpoints": [
        {
            "Endpoint": "/metrics",
            "Redirects": [
                {
                    "URL": "http://0.0.0.0/metrics"
                },
                {
                    "URL": "http://localhost:8081/m"
                }
            ]
        },
        {
            "Endpoint": "/forward",
            "Redirects": [
                {
                    "URL": "https://invaliddomain"
                },
                {
                    "URL": "https://someupstreamservice.com"
                }
            ]
        }
    ],
    "Port": 8080
}
```

# Build

```sh
go run *.go
```

```sh
go build
```