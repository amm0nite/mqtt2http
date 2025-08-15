# MQTT2HTTP

A simple MQTT broker that connects to your HTTP services for login and message handling.

## Table of Contents

* [Features](#features)
* [Missing Features](#missing-features)
* [Quick Start](#quick-start)
* [Docker](#docker)
* [Configuration](#configuration)
* [Metrics](#metrics)

## Features

* **Authentication**: Validates MQTT `CONNECT` requests using HTTP Basic Auth
* **MQTT to HTTP**: Forwards `PUBLISH` messages as HTTP `POST` requests
* **HTTP to MQTT**: Accepts HTTP `POST` requests to publish MQTT messages
* **Metrics**: Exposes Prometheus-compatible metrics

## Missing features

* **ACL**: Limits wich topic can be published or subscribed to. Everything is allowed for now.

## Quick Start

Set the URLs for your HTTP services using environment variables:

```env
MQTT2HTTP_AUTHORIZE_URL=http://...
MQTT2HTTP_PUBLISH_URL=http://...
```

### MQTT to HTTP

* When a client connects, its username and password are sent to your authorization endpoint.
* All MQTT `PUBLISH` messages are forwarded as HTTP `POST` requests to the specified URL.

### HTTP to MQTT

* Publish messages to MQTT topics using the built-in REST API.

```bash
curl --user username:password -X POST -d '{"test": true}' http://mqtt2http:8080/publish?topic=hello
```

## Docker

Run with Docker Compose:

```yaml
mqtt2http:
  image: docker.io/amm0nite/mqtt2http:latest
  ports:
    - 1883:1883     # MQTT
    - 8088:8080     # HTTP API
    - 9090:9090     # Prometheus metrics
  environment:
    MQTT2HTTP_AUTHORIZE_URL: http://auth.service/
    MQTT2HTTP_PUBLISH_URL: http://backend.service/api/{topic}
```

To use a specific version:

```yaml
image: docker.io/amm0nite/mqtt2http:1.0.0
```

## Configuration

| Variable                                | Default                      | Description                                                                                    |
| --------------------------------------- | ---------------------------- | ---------------------------------------------------------------------------------------------- |
| `MQTT2HTTP_MQTT_LISTEN_ADDRESS`         | `:1883`                      | Address where the MQTT broker listens (host\:port).                                            |
| `MQTT2HTTP_HTTP_LISTEN_ADDRESS`         | `:8080`                      | Address for the HTTP REST API (`/publish` endpoint).                                           |
| `MQTT2HTTP_AUTHORIZE_URL`               | `http://example.com`         | HTTP Basic Auth endpoint for authorizing `CONNECT` requests. A 200/201 response allows access. |
| `MQTT2HTTP_PUBLISH_URL`                 | `http://example.com/{topic}` | Template URL for forwarding `PUBLISH` messages. `{topic}` is replaced dynamically.             |
| `MQTT2HTTP_CONTENT_TYPE`                | `application/octet-stream`   | `Content-Type` header used in forwarded HTTP `POST` requests. E.g., `application/json`.        |
| `MQTT2HTTP_TOPIC_HEADER`                | `X-Topic`                    | Name of the HTTP header that carries the MQTT topic.                                           |
| `MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS` | `:9090`                      | Address for serving Prometheus metrics at the `/metrics` endpoint.                             |
| `MQTT2HTTP_ROUTES_FILE_PATH` | `routes.yaml` | Path for the yaml file that defines all routes.
| `MQTT2HTTP_API_PASSWORD` | random value | Password used to secure the API endpoints.

## Metrics

Prometheus metrics are available at `/metrics` on the configured metrics address (`MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS`).

| Metric                         | Type    | Labels          | Description                                                                                         |
| ------------------------------ | ------- | --------------- | --------------------------------------------------------------------------------------------------- |
| `mqtt2http_publish_count`      | Counter | `topic`, `code` | Counts forwarded MQTT `PUBLISH` messages as HTTP `POST` requests, labeled by topic and HTTP status. |
| `mqtt2http_authenticate_count` | Counter | `code`          | Counts HTTP authentication attempts during MQTT `CONNECT`, labeled by HTTP status code.             |
