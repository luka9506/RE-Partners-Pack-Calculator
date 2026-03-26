# RE Partners Pack Calculator

This project implements the pack-calculation assessment as a single Go web application. It exposes an HTTP API and serves a browser UI from the same service.

## Features

- Go HTTP API for pack calculation
- Server-rendered UI for manual testing
- Config-driven pack sizes loaded from `config/packs.json`
- Unit tests for the calculator and HTTP layer
- Dockerfile for deployment

## How it works

The calculator applies the rules in the required order:

1. Ship only whole packs
2. Minimize the total number of shipped items while still meeting the order quantity
3. Among those minimal-shipment solutions, minimize the number of packs

The implementation uses dynamic programming rather than a greedy algorithm, which ensures correctness across configurable pack sizes.

## Configuration

Pack sizes are defined in [config/packs.json](config/packs.json):

```json
{
  "pack_sizes": [250, 500, 1000, 2000, 5000]
}
```

You can add, remove, or change pack sizes without changing the application code.

## Run locally

Prerequisite: Go 1.22 or newer.

```bash
go run ./cmd/server
```

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

## Tests

```bash
go test ./...
```

## Docker

Build and run:

```bash
docker build -t pack-calculator .
docker run -p 8080:8080 pack-calculator
```

## Deployment recommendation

A simple target such as Railway, Render, or Fly.io is sufficient for this project. The included Dockerfile is enough for a container-based deployment.
