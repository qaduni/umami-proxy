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
eyJjb21wb3NlIjogInNlcnZpY2VzOlxuICB1bWFtaS1wcm94eTpcbiAgICBpbWFnZTogZ2hjci5pby9xYWR1bmkvdW1hbWktcHJveHk6bGF0ZXN0XG4gICAgcHVsbF9wb2xpY3k6IGFsd2F5c1xuICAgIHJlc3RhcnQ6IGFsd2F5c1xuICAgIGV4cG9zZTpcbiAgICAgIC0gXCI4MDgwXCJcbiAgICBlbnZpcm9ubWVudDpcbiAgICAgIC0gVU1BTUlfVVJMPSR7VU1BTUlfVVJMfVxuICAgICAgLSBVTUFNSV9BUElfS0VZPSR7VU1BTUlfQVBJX0tFWX0iLCAiY29uZmlnIjogIlt2YXJpYWJsZXNdXG51bWFtaV91cmwgPSBcInN0YXRzLmV4YW1wbGUuY29tXCJcbnVtYW1pX2FwaV9rZXkgPSBcIlJFUExBQ0VfTUVfV0lUSF9BUElfS0VZXCJcbnVtYW1pX3Byb3h5X2RvbWFpbiA9IFwiJHtkb21haW59XCJcblxuW2NvbmZpZ11cbmVudiA9IFtcIlVNQU1JX1VSTD0ke3VtYW1pX3VybH1cIiwgXCJVTUFNSV9BUElfS0VZPSR7dW1hbWlfYXBpX2tleX1cIl1cblxuW1tjb25maWcuZG9tYWluc11dXG5zZXJ2aWNlTmFtZSA9IFwidW1hbWktcHJveHlcIlxucG9ydCA9IDgwODBcbmhvc3QgPSBcIiR7dW1hbWlfcHJveHlfZG9tYWlufVwiXG5odHRwcyA9IHRydWVcbiJ9
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