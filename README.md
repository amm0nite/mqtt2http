MQTT2HTTP
[![GitHub Release](https://img.shields.io/github/v/release/amm0nite/mqtt2http)](https://github.com/amm0nite/mqtt2http/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/amm0nite/mqtt2http)](https://hub.docker.com/r/amm0nite/mqtt2http)
=========

MQTT broker that forwards topics to HTTP endpoints with configurable routing.

## Table of Contents

* [Features](#features)
* [Missing Features](#missing-features)
* [Quick Start](#quick-start)
* [Docker](#docker)
* [Configuration](#configuration)
* [Routing](#routing)
* [Metrics](#metrics)

## Features

* **Authentication**: Validates MQTT `CONNECT` requests using HTTP Basic Auth
* **MQTT to HTTP**: Forwards `PUBLISH` messages as HTTP `POST` requests
* **HTTP to MQTT**: Accepts HTTP `POST` requests to publish MQTT messages
* **Metrics**: Exposes Prometheus-compatible metrics

## Missing features

* **ACL**: Limits which topic can be published or subscribed to. Everything is allowed for now.

## Quick Start

Set the URLs for your HTTP services using environment variables:

```env
MQTT2HTTP_AUTHORIZE_URL=http://...
MQTT2HTTP_PUBLISH_URL=http://...
MQTT2HTTP_API_PASSWORD=somesecret
```

### MQTT to HTTP

* When a client connects, its username and password are sent to your authorization endpoint.
* All MQTT `PUBLISH` messages are forwarded as HTTP `POST` requests to the specified URL.

### HTTP to MQTT

* Publish messages into MQTT topics using the built-in REST API.
* Requests to `/publish` must include HTTP Basic Auth with the password set in `MQTT2HTTP_API_PASSWORD` (the username is ignored).
* If `MQTT2HTTP_API_PASSWORD` is not set, a random password is generated at start-up and is not exposed by the API, so set it explicitly if you want to use the authenticated endpoints.

`/publish` lets you inject payloads into MQTT topics:

```bash
curl --user user:somesecret -X POST -d '{"test": true}' http://mqtt2http:8080/publish?topic=hello
```

### Information endpoints

`/` returns the application version and broker runtime information:

```bash
curl http://mqtt2http:8080/
```

`/clients` dumps the active MQTT sessions, including their username, subscriptions, publication counters, and timestamps:

```bash
curl --user user:somesecret http://mqtt2http:8080/clients
```

Requests to `/clients` must include HTTP Basic Auth with the password set in `MQTT2HTTP_API_PASSWORD` (the username is ignored).

The endpoint responds with a JSON array of objects matching the structure of `lib.Client` (fields: `id`, `username`, `subscriptions`, `publications`, `connected_at`, `last_activity_at`).

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
    MQTT2HTTP_API_PASSWORD: somesecret
```

To use a specific version:

```yaml
image: docker.io/amm0nite/mqtt2http:1.0.0
```

## Configuration

| Variable                                | Default                            | Description |
| --------------------------------------- | ---------------------------------- | ----------- |
| `MQTT2HTTP_MQTT_LISTEN_ADDRESS`         | `:1883`                            | Address where the MQTT broker listens (host\:port). |
| `MQTT2HTTP_HTTP_LISTEN_ADDRESS`         | `:8080`                            | Address for the HTTP REST API (hosts `/`, `/publish`, and `/clients`). |
| `MQTT2HTTP_AUTHORIZE_URL`               | `http://127.0.0.1/authorize`       | HTTP endpoint used to authorize MQTT `CONNECT` requests. The broker sends an HTTP `POST` with Basic Auth credentials and no body. A `200` or `201` response allows access. |
| `MQTT2HTTP_PUBLISH_URL`                 | `http://127.0.0.1/publish/{topic}` | Template URL for forwarding `PUBLISH` messages. `{topic}` is replaced once in the outgoing URL. When no routes file is loaded, this URL is used for a catch-all default route. |
| `MQTT2HTTP_CONTENT_TYPE`                | `application/octet-stream`         | `Content-Type` header used in forwarded HTTP `POST` requests. E.g. `application/json`. |
| `MQTT2HTTP_TOPIC_HEADER`                | `X-Topic`                          | Name of the HTTP header that carries the MQTT topic on forwarded publish requests. |
| `MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS` | `:9090`                            | Address for serving Prometheus metrics at the `/metrics` endpoint. |
| `MQTT2HTTP_ROUTES_FILE_PATH`            | `routes.yaml`                      | Path to the YAML file that defines routing rules loaded at start-up. |
| `MQTT2HTTP_API_PASSWORD`                | random value                       | Password used to secure `/publish` and `/clients`. If unset, a random password is generated at start-up. |

## Routing

Define fine-grained routing rules in a YAML file that is loaded at start-up. By default the broker looks for `routes.yaml` in the working directory, or you can set `MQTT2HTTP_ROUTES_FILE_PATH` to point to a different file.

Each entry in the file is a map with three fields:

* `name`: friendly identifier used in logs when the route matches.
* `pattern`: Go regular expression tested against the MQTT topic (`^` / `$` anchors are optional).
* `url`: target HTTP endpoint to receive the forwarded payload. Leave empty to drop messages for this route after a match.

Example `routes.yaml`:

```yaml
- name: telemetry
  pattern: '^sensors/.+'
  url: https://example.com/iot/publish
- name: drop-debug
  pattern: '^debug/'
  url: ''
- name: fallback
  pattern: '.*'
  url: https://example.com/default/{topic}
```

Routes are evaluated in order and the first match wins. If no route matches, the broker logs the miss and no HTTP request is sent. When the routes file is empty (or missing) and `MQTT2HTTP_PUBLISH_URL` is configured, a default catch-all route using that URL is created automatically.

Both the authorization and publish HTTP requests use a 5 second client timeout.

## Metrics

Prometheus metrics are available at `/metrics` on the configured metrics address (`MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS`).

| Metric                        | Type   | Labels        | Description                                                                                          |
| ----------------------------- | ------ | ------------- | ---------------------------------------------------------------------------------------------------- |
| `mqtt2http_sessions`          | Gauge  | _none_        | Tracks the current number of connected MQTT sessions.                                                |
| `mqtt2http_authenticate_count`| Counter| `url`, `code` | Counts HTTP Basic Auth attempts made during MQTT `CONNECT`, labeled by authorization URL and status. |
| `mqtt2http_publish_count`     | Counter| `topic`       | Counts MQTT `PUBLISH` packets received per topic.                                                    |
| `mqtt2http_forward_count`     | Counter| `url`, `code` | Counts HTTP requests sent while forwarding MQTT payloads, labeled by the resolved URL and status.    |
| `mqtt2http_subscribe_count`   | Counter| `topic`       | Counts subscription requests per topic.                                                              |
| `mqtt2http_no_match_count`    | Counter| `topic`       | Counts messages for which no route was found.                                                        |
