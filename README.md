# sing-box-exporter

A robust Prometheus exporter for Sing-box / V2Ray, written in Go.

## Features
- **Zero Config Secrets**: Pass sensitive data via flags.
- **Native gRPC**: Connects directly to Sing-box's V2Ray Stats API.
- **Multi-metric**: Tracks Per-User and Per-Inbound traffic.
- **Low Overhead**: ~10MB binary, minimal CPU/RAM usage.

## Usage

### 1. Enable V2Ray API in Sing-box
Ensure your `config.json` has:
```json
"experimental": {
  "v2ray_api": {
    "listen": "127.0.0.1:19998",
    "stats": {
      "enabled": true,
      "inbounds": ["main"],
      "users": ["user1", "user2"]
    }
  }
}
```

### 2. Run Exporter
```bash
./sing-box-exporter \
  -singbox 127.0.0.1:19998 \
  -users "user1,user2" \
  -inbounds "inbound-1,inbound-2"
```

### 3. Check Metrics
```bash
curl localhost:9091/metrics
```

## Prometheus Configuration

Add this to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'sing-box'
    scrape_interval: 15s
    static_configs:
      - targets: ['localhost:9091']

## Multi-Target Configuration

If you have exporters running on multiple servers:

```yaml
scrape_configs:
  - job_name: 'sing-box-cluster'
    scrape_interval: 15s
    static_configs:
      - targets:
        - 'server-us.example.com:9091'
        - 'server-jp.example.com:9091'
        - '192.168.1.100:9091'
```

---
> This project was written autonomously by **Antigravity's Gemini 3 Pro (High)**. 🤖✨
```
