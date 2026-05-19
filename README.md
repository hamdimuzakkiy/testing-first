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

## OCR

The service also exposes a local OCR endpoint:

```sh
curl -X POST http://localhost:8080/ocr \
  -F "image=@/path/to/image.png"
```

Expected JSON response:

```json
{"text":"extracted text"}
```

By default, `/ocr` runs:

```sh
tesseract <uploaded-image> stdout
```

You can point it at another local OCR model or wrapper:

```sh
OCR_COMMAND=/path/to/local-ocr go run .
```

For extra arguments after `stdout`:

```sh
OCR_COMMAND=tesseract OCR_ARGS="-l eng" go run .
```

## Test

```sh
go test ./...
```
