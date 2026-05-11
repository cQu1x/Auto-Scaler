# AutoScaler

A Go web application demonstrating Kubernetes Horizontal Pod Autoscaling (HPA) with Prometheus metrics.

## Endpoints

| Method | Path      | Description                              |
|--------|-----------|------------------------------------------|
| GET    | /health   | Health check                             |
| GET    | /metrics  | Prometheus metrics                       |
| POST   | /load     | Generate CPU load to trigger HPA scaling |

### POST /load

```json
{ "amount": 4, "duration": 30 }
```

- `amount` — number of CPU-bound goroutines (1–64, default: number of CPUs)
- `duration` — how long to run in seconds (1–300, default: 10)

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [minikube](https://minikube.sigs.k8s.io/docs/start/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Helm](https://helm.sh/docs/intro/install/)
- [k6](https://k6.io/docs/get-started/installation/)

## Quick Start

### 1. Start minikube

```bash
minikube start
minikube addons enable metrics-server
minikube addons enable ingress
```

### 2. Build and load the image

```bash
make load-image
```

### 3. Deploy the application

```bash
make deploy
```

Verify the app is running via port-forward:

```bash
# Terminal 1 — keep open
kubectl port-forward svc/auto-scaler-service 8080:80

# Terminal 2
curl http://localhost:8080/health
```

> **Note (WSL2 / Docker driver):** Ingress is not directly reachable on WSL2 with the Docker driver because NodePorts are not exposed from the minikube container. Use `kubectl port-forward` as shown above, or run `minikube tunnel` in a separate terminal to enable Ingress access via `http://auto-scaler.local`.

### 4. Install Prometheus + Grafana

```bash
make setup-monitoring
```

Wait until all pods are running:

```bash
kubectl get pods -w
```

Access Grafana (credentials: `admin` / `prom-operator`):

```bash
minikube service kube-prometheus-grafana
```

> **Note (WSL2):** If the browser doesn't open automatically, run `minikube service kube-prometheus-grafana --url` to get the URL and open it manually.

### 5. Run load test

Open three terminals:

```bash
# Terminal 1 — port-forward (keep open)
kubectl port-forward svc/auto-scaler-service 8080:80

# Terminal 2 — watch HPA react in real time
make watch-hpa

# Terminal 3 — run the load test
BASE_URL=http://localhost:8080 make load-test
```

Expected behavior: CPU usage rises above 50% → HPA scales pods up to max 5 → load ends → pods scale back down after ~60s.

## Metrics

The app exposes the following Prometheus metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total requests by method, path, status |
| `http_request_duration_seconds` | Histogram | Request latency |
| `http_requests_dropped_total` | Counter | Requests canceled or timed out |

Useful PromQL queries in Grafana (Explore → Prometheus):

```
# Replica count over time
kube_deployment_status_replicas{deployment="auto-scaler"}

# Requests per second
rate(http_requests_total[1m])

# Average response time
rate(http_request_duration_seconds_sum[1m]) / rate(http_request_duration_seconds_count[1m])
```

## Project Structure

```
.
├── cmd/
│   └── main.go
├── internal/
│   ├── handlers/
│   │   ├── dto.go
│   │   ├── health.go
│   │   ├── load.go
│   │   └── metrics.go
│   └── util/
│       └── json.go
├── deploy/k8s/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── hpa.yaml
│   └── servicemonitor.yaml
├── load-test/
│   └── load-test.js
├── Dockerfile
└── Makefile
```
