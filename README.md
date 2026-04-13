# floodgate

A lightweight rate-limiting middleware library for Go HTTP services with pluggable backends.

## Installation

```bash
go get github.com/yourusername/floodgate
```

## Usage

```go
package main

import (
    "net/http"
    "github.com/yourusername/floodgate"
)

func main() {
    limiter := floodgate.New(floodgate.Options{
        RequestsPerSecond: 10,
        Burst:             20,
        Backend:           floodgate.InMemoryBackend(),
    })

    mux := http.NewServeMux()
    mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, world!"))
    })

    http.ListenAndServe(":8080", limiter.Wrap(mux))
}
```

Floodgate wraps any `http.Handler` and automatically rejects requests that exceed
your configured rate limits with a `429 Too Many Requests` response.

## Backends

| Backend      | Description                          |
|--------------|--------------------------------------|
| `InMemory`   | Default, single-instance in-process  |
| `Redis`      | Distributed, suitable for clusters   |

## Configuration

| Option              | Default | Description                          |
|---------------------|---------|--------------------------------------|
| `RequestsPerSecond` | `10`    | Sustained request rate per client    |
| `Burst`             | `20`    | Maximum burst size                   |
| `KeyFunc`           | IP-based| Function to identify clients         |

## License

MIT © [yourusername](https://github.com/yourusername)