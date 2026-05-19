# testing-first

A tiny Go HTTP service with a `/ping` endpoint.

## Run

```sh
go run .
```

Then call:

```sh
curl http://localhost:8080/ping
```

Expected response:

```text
pong
```

## Test

```sh
go test ./...
```
