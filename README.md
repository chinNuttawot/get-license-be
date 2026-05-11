# Get License Backend

Go Fiber backend for importing encrypted `.aglic` bundles, storing decoded license metadata, and verifying signed license tokens.

## Requirements

- Go 1.23+
- PostgreSQL
- Environment variables from `.env.example`

## Run

```bash
go run ./cmd/server
```

Default port: `4001`

## Build

```bash
go build -o bin/get-license-be ./cmd/server
```

## Test

```bash
go test ./...
```

## API

### `GET /health`

Health check.

### `POST /api/client-license/import`

Imports an encrypted `.aglic` bundle.

Request:

```json
{
  "fileContent": "base64-aglic-content"
}
```

### `GET /api/client-license`

Returns the latest imported license with verified tokens.

### `GET /api/client-license/all`

Returns all non-deleted imported licenses.

### `GET /api/client-license/:id`

Returns one imported license with verified tokens.

### `PATCH /api/client-license/:id/status`

Toggles `isActive`.

### `DELETE /api/client-license/:id`

Soft deletes a license by setting `isDeleted = true`.

## Project Structure

```text
cmd/server              application entrypoint
internal/config         environment and database config
internal/http           Fiber routes, handlers, response errors
internal/httperr        shared HTTP error type
internal/license        bundle decryption and token verification
internal/store          PostgreSQL connection, schema, license persistence
```

## Environment

```bash
PUBLIC_KEY=
BUNDLE_KEY=
PORT=4001
DB_HOST=
DB_PORT=5432
DB_USER=
DB_PASSWORD=
DB_NAME=
DB_SSL=true
```
