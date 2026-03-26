FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY cmd ./cmd
COPY config ./config
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -o /pack-calculator ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /pack-calculator /app/pack-calculator
COPY config /app/config
EXPOSE 8080
ENV PORT=8080
ENV PACK_CONFIG_PATH=/app/config/packs.json
CMD ["/app/pack-calculator"]
