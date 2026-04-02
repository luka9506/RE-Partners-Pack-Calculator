# RE Partners Pack Calculator

This project implements the pack-calculation assessment as a single Go web application. It exposes an HTTP API and serves a browser UI from the same service.

## Features

- Go HTTP API for pack calculation
- Server-rendered UI for manual testing
- Config-driven pack sizes loaded from `config/packs.json`
- Unit tests for the calculator and HTTP layer
- Dockerfile for deployment

## Project structure

- `cmd/server`: application entrypoint and runtime wiring
- `internal/config`: configuration loading and validation
- `internal/packing`: core calculation logic
- `internal/httpapi`: HTTP handlers and server-rendered UI
- `config/packs.json`: default pack-size configuration

## How it works

The calculator applies the rules in the required order:

1. Ship only whole packs
2. Minimize the total number of shipped items while still meeting the order quantity
3. Among those minimal-shipment solutions, minimize the number of packs

The implementation uses dynamic programming rather than a greedy algorithm, which ensures correctness across configurable pack sizes.

Think of this like making change with coins: there may be many combinations that reach or exceed a target amount. The application must first choose the combination with the smallest total shipped quantity, then choose the one with the fewest packs among those tied minimal totals.

## Configuration

Pack sizes are defined in [config/packs.json](config/packs.json):

```json
{
  "pack_sizes": [250, 500, 1000, 2000, 5000]
}
```

You can add, remove, or change pack sizes without changing the application code.

Example edge-case configuration:

```json
{
  "pack_sizes": [23, 31, 53]
}
```

For an order quantity of `500000`, the expected exact result is:

```json
{
  "23": 2,
  "31": 7,
  "53": 9429
}
```

That scenario is covered by a unit test to prove that configurable pack sizes continue to produce the correct minimal solution.

## Run locally

Prerequisite: Go 1.22 or newer.

```bash
go run ./cmd/server
```

Open `http://localhost:8080` for the UI.

The app listens on port `8080` by default.

Environment variables:

- `PORT`: HTTP port, default `8080`
- `PACK_CONFIG_PATH`: path to the config file, default `config/packs.json`

## API

### Calculate packs

`POST /api/calculate`

Request:

```json
{
  "quantity": 501
}
```

Response:

```json
{
  "ordered_quantity": 501,
  "total_items": 750,
  "overfill": 249,
  "packs": [
    {
      "pack_size": 500,
      "count": 1
    },
    {
      "pack_size": 250,
      "count": 1
    }
  ]
}
```

### Get current pack configuration

`GET /api/config/packs`

Response:

```json
{
  "pack_sizes": [250, 500, 1000, 2000, 5000]
}
```

### Update pack configuration

`PUT /api/config/packs`

Request:

```json
{
  "pack_sizes": [23, 31, 53]
}
```

Response:

```json
{
  "pack_sizes": [23, 31, 53]
}
```

A successful update is validated, written back to the configured JSON file, and applied immediately by the running server. Invalid payloads return `400`, and persistence failures return `500` without changing the active calculator.

## Tests

```bash
go test ./...
```

The test suite covers:

- all examples from the assessment brief
- invalid input handling
- configurable pack sizes
- the provided `23,31,53` / `500000` edge case
- rule precedence between minimizing shipped quantity and minimizing pack count

## Docker

The application is containerized so reviewers can build and test it without a local Go setup.

Build the image:

```bash
docker build -t pack-calculator .
```

Run the container:

```bash
docker run -p 8080:8080 pack-calculator
```

Open `http://localhost:8080` for the UI.

Test the API from another terminal:

```bash
curl -X POST http://localhost:8080/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"quantity":501}'
```

Fetch the current pack sizes:

```bash
curl http://localhost:8080/api/config/packs
```

Replace the active pack sizes:

```bash
curl -X PUT http://localhost:8080/api/config/packs \
  -H "Content-Type: application/json" \
  -d '{"pack_sizes":[23,31,53]}'
```

Run the container with a custom configuration file:

```bash
docker run -p 8080:8080 \
  -e PACK_CONFIG_PATH=/app/config/custom-packs.json \
  -v $(pwd)/config:/app/config \
  pack-calculator
```

The container sets `PACK_CONFIG_PATH=/app/config/packs.json` by default so the built image works without extra flags.

## Architecture notes

- The packing algorithm is isolated from HTTP concerns so it can be tested independently.
- Configuration is loaded at startup and can also be updated through the config API; successful updates are validated, persisted, and swapped into the running server immediately.
- The browser UI and JSON API use the same calculator state, which prevents rule drift between interfaces.
- There is no persistence layer because the assessment only requires stateless calculation. If this were extended into a production order service, persistence would be a reasonable next step for storing order history, configuration versions, or audit data.

## Deployment recommendation

A simple target such as Railway, Render, or Fly.io is sufficient for this project. The included Dockerfile is enough for a container-based deployment.

## Verification note

This workspace does not currently have Go or Docker installed, so local execution could not be performed here. The project has been structured so that `go test ./...`, `go run ./cmd/server`, and the Docker commands above are the exact verification steps to run in a machine with those tools installed.
