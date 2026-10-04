# Monitoring

Server metrics API for the dashboard: live and historical CPU, RAM, disk and network usage, read from Prometheus.
Private, behind SSO.

## Stack

- Go + chi
- Prometheus (5 s scrape, 15-day retention) fed by node-exporter

## Routes

| Method | Path                                     | Description                                                  |
| ------ | ---------------------------------------- | ------------------------------------------------------------ |
| GET    | `/health`                                | Health check                                                 |
| GET    | `/metrics/cpu`                           | CPU usage: `{ usage_percent, timestamp }`                    |
| GET    | `/metrics/ram`                           | RAM usage: `{ usage_percent, timestamp }`                    |
| GET    | `/metrics/disk`                          | Root disk usage: `{ usage_percent, timestamp }`              |
| GET    | `/metrics/network`                       | Network: `{ rx_bytes_per_sec, tx_bytes_per_sec, timestamp }` |
| GET    | `/metrics/{cpu,ram,disk}/history?range=` | History: `[{ timestamp, value }]`                            |
| GET    | `/metrics/network/history?range=`        | History: `[{ timestamp, rx, tx }]` (bytes/s)                 |

`timestamp` is a Unix time in seconds. Prometheus errors, empty results and unparsable responses are returned as 502.

History `range` is `1h`, `24h` or `7d` (step 15 s, 5 min, 30 min); anything else returns 400. Points are sorted by
`timestamp`; a range with no data returns `[]`.

## Configuration

| Variable         | Required | Default                  | Description    |
| ---------------- | -------- | ------------------------ | -------------- |
| `PROMETHEUS_URL` | No       | `http://prometheus:9090` | Prometheus URL |
| `SERVER_PORT`    | No       | `8080`                   | HTTP port      |

## Development

No local run script: the API needs Prometheus and node-exporter.

```bash
go vet ./...
go test ./...
```

## Production

Deployed at `monitoring.yankadevlab.tech` behind Traefik, by CI on push to `main`. Protected by the `auth-verify`
middleware; a Traefik CORS middleware placed before it allows `GET` from the dashboard origin.
