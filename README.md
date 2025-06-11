# mqtt2http

MQTT broker with hooks to send authentication and publish events as HTTP requests.

This MQTT broker receives MQTT messages with a topic on port 1883 and forwards them as POST requests to an HTTP service (see `MQTT2HTTP_PUBLISH_URL`). 
The topic is either included in the URL or forwarded as the HTTP header X-Topic (see `MQTT2HTTP_TOPIC_HEADER` for customization).
When a client connects to the broker, it will send an HTTP POST request to `MQTT2HTTP_AUTHORIZE_URL`  with the client credentials as HTTP Basic Auth to validate the username and password.

# http2mqtt

There is also an API endpoint available to be used as http2mqtt at `http://localhost:8080/publish?topic=...` (see `MQTT2HTTP_HTTP_LISTEN_ADDRESS` to customize the port). 
It will publish the body of a POST request to the topic passed as query parameter. The service requires HTTP Basic authentication.

    curl --user username:password -X POST -d '{"test": true}' http://mqtt2http:8080/publish?topic=test

This example will publish the message `{"test": true}` to the topic `test`.

# Run with Docker

This tool is published as `amm0nite/mqtt2http` on Docker Hub. You can run it, for example, using Docker Compose:

    mqtt2http:
      image: "docker.io/amm0nite/mqtt2http:latest"
      restart: "unless-stopped"
      ports:
        - 1883:1883
        - 8088:8080
        - 9090:9090
      environment:
        MQTT2HTTP_AUTHORIZE_URL: http://auth.service/
        MQTT2HTTP_PUBLISH_URL: http://backend.service/api/{topic}

This will expose the MQTT port (1883), the HTTP publish interface (on port 8080), and a Prometheus metrics endpoint (on port 9090).

# Configuration

The mqtt2http tool provides several configuration options:

- `MQTT2HTTP_MQTT_LISTEN_ADDRESS`

  Default: :1883

  The TCP address (host:port) on which the embedded MQTT broker will listen for client connections.

- `MQTT2HTTP_HTTP_LISTEN_ADDRESS`

  Default: :8080
  
  The HTTP address (host:port) for the built-in REST API (/publish endpoint).

- `MQTT2HTTP_AUTHORIZE_URL`
  
  Default: http://example.com
  
  The URL to which mqtt2http will POST (with Basic Auth headers) whenever a client attempts to CONNECT. A 200 or 201 response means “allow”; any other response is treated as a failure.

- `MQTT2HTTP_PUBLISH_URL`
  
  Default: http://example.com/{topic}
  
  The URL template used when forwarding published messages. The {topic} placeholder is replaced with the actual MQTT topic name (optional).

- `MQTT2HTTP_CONTENT_TYPE`
  
  Default: application/octet-stream
  
  The Content-Type header used in the HTTP POST when forwarding the raw payload of a PUBLISH packet.
  Example: if the message is a JSON string, set this to application/json.

- `MQTT2HTTP_TOPIC_HEADER`
  
  Default: X-Topic
  
  The name of the custom HTTP header in which the MQTT topic name will be sent during a PUBLISH POST.

- `MQTT2HTTP_METRICS_HTTP_LISTEN_ADDRESS`
  
  Default: :9090
  
  The HTTP address (host:port) on which the Prometheus metrics endpoint (/metrics) will be exposed.

## Minimum Required Configuration

To use mqtt2http in a meaningful way, at least the following environment variables must be set:

    MQTT2HTTP_AUTHORIZE_URL=http://...
    MQTT2HTTP_PUBLISH_URL=http://...
