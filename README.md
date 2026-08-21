# Umami Analytics Stats Proxy

A lightweight, concurrent Go HTTP proxy service that aggregates website traffic statistics from a self-hosted [Umami Analytics](https://umami.is) instance. 

It queries Umami’s REST API concurrently using Go goroutines and returns merged **Today**, **Week**, **Month**, and **Total** pageview/visitor metrics in a single JSON payload. Designed for static site generators (Hugo, Astro, Jekyll) and serverless frontend integrations.

---

## Features

- **High Performance Pooling:** Uses a global, shared HTTP client with connection pooling for scaling across dozens of websites effortlessly.
- **Zero Hardcoded Configs:** Fully driven by environment variables at runtime.
- **Concurrent API Fetching:** Fetches all 4 time intervals in parallel via Go goroutines.
- **CORS Enabled:** Built-in `Access-Control-Allow-Origin: *` for direct client-side web calls.
- **Minimal Docker Footprint:** Multi-stage build on Alpine Linux with zero external dependencies.
- **Dokploy Ready:** Native configuration for container deployments on Dokploy.

---

## Dokploy One-Click Import

Copy the Base64 payload below and paste it directly into Dokploy under **Templates -> Import**:

<!-- DOKPLOY_BASE64_START -->
```text
eyJjb21wb3NlIjogInNlcnZpY2VzOlxuICB1bWFtaS1wcm94eTpcbiAgICBpbWFnZTogZ2hjci5pby9xYWR1bmkvdW1hbWktcHJveHk6bGF0ZXN0XG4gICAgcmVzdGFydDogYWx3YXlzXG4gICAgZXhwb3NlOlxuICAgICAgLSBcIjgwODBcIlxuICAgIGVudmlyb25tZW50OlxuICAgICAgLSBVTUFNSV9VUkw9JHtVTUFNSV9VUkx9XG4gICAgICAtIFVNQU1JX0FQSV9LRVk9JHtVTUFNSV9BUElfS0VZfSIsICJjb25maWciOiAiW3ZhcmlhYmxlc11cbnVtYW1pX3VybCA9IFwic3RhdHMuZXhhbXBsZS5jb21cIlxudW1hbWlfYXBpX2tleSA9IFwiUkVQTEFDRV9NRV9XSVRIX0FQSV9LRVlcIlxudW1hbWlfcHJveHlfZG9tYWluID0gXCIke2RvbWFpbn1cIlxuXG5bY29uZmlnXVxuZW52ID0gW1wiVU1BTUlfVVJMPSR7dW1hbWlfdXJsfVwiLCBcIlVNQU1JX0FQSV9LRVk9JHt1bWFtaV9hcGlfa2V5fVwiXVxuXG5bW2NvbmZpZy5kb21haW5zXV1cbnNlcnZpY2VOYW1lID0gXCJ1bWFtaS1wcm94eVwiXG5wb3J0ID0gODA4MFxuaG9zdCA9IFwiJHt1bWFtaV9wcm94eV9kb21haW59XCJcbmh0dHBzID0gdHJ1ZVxuIn0=
```
<!-- DOKPLOY_BASE64_END -->

## API Usage

### Endpoint

```http
GET /stats?id={UMAMI_WEBSITE_UUID}
```

### Example Response

```json
{
  "website_id": "3b2e3230-22c1-4ba2-b2d2-8b65e90d2341",
  "today": {
    "pageviews": 142,
    "visitors": 98
  },
  "week": {
    "pageviews": 1250,
    "visitors": 810
  },
  "month": {
    "pageviews": 5400,
    "visitors": 3100
  },
  "total": {
    "pageviews": 45000,
    "visitors": 22000
  }
}
```

---

## Deployment with Dokploy

1. **Create Compose Service**: Add a new Docker Compose application in Dokploy.
2. **Set Environment Variables**: In your Dokploy application settings under **Environment**, define:
   ```env
   UMAMI_URL=https://analytics.qu.edu.iq
   ```
3. **Configure Domain**:
   - Assign your desired domain name (e.g., `stats-proxy.qu.edu.iq`).
   - Set **Target Port / Container Port** to `8080`.
4. **Deploy**: Trigger a manual deployment in Dokploy.

---

## Local Development & Docker Run

### Running Locally

```bash
export UMAMI_URL="https://analytics.qu.edu.iq"
go run main.go
```

### Running with Docker

```bash
docker build -t umami-proxy .
docker run -d \
  -p 8080:8080 \
  -e UMAMI_URL="https://analytics.qu.edu.iq" \
  --name umami-proxy \
  umami-proxy
```

---

## Project Structure

```text
.
├── main.go            # Go proxy application server with shared HTTP pool
├── Dockerfile         # Multi-stage lightweight Docker build
├── docker-compose.yml # Docker Compose specification for Dokploy
└── README.md          # Project documentation
```

---

## Environment Variables

| Variable | Required | Description | Default |
| :--- | :--- | :--- | :--- |
| `UMAMI_URL` | **Yes** | Base URL of your self-hosted Umami instance | *None* |

---

## License

[MIT](LICENSE)