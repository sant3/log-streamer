The backend is a lightweight Go application that serves log files from a specified directory over an HTTP/S WebSocket API. It is designed to be run as an agent on a server where log files are generated.

## Certificate

To run the server with HTTPS (enabled by default), you need a TLS certificate. You can generate a self-signed certificate and a decrypted private key using `make`:

```sh
make generate-self-signed-cert
```

This command will create `cert-streamer.pem` and `decrypted_key.pem` in the `backend` directory.

> **Note**: Go's native TLS support does not handle encrypted private keys. The command above generates a decrypted key for this reason. If you have an existing encrypted key, you can decrypt it using: `openssl rsa -in your_key.pem -out decrypted_key.pem`.

The default password for the certificate is `password`, and the distinguished name is `/C=IT/ST=Italy/L=Turin/O=MyCompany/OU=MyDivision`.

---

### How to run

`go run .`

`go run . &`

---

### Configurations

Default port is 5005

Default https port is 8443

Priority config is cli > os env

#### Cli

Available flags:

- `-port` Listening HTTP port (default 5005)
- `-httpsPort` Listening HTTPS port (default 8443)
- `-enableHTTPS` Enable HTTPS server (default true)
- `-tlsCert` Path to TLS certificate (default `cert-streamer.pem`)
- `-tlsKey` Path to TLS private key (default `decrypted_key.pem`)
- `-logsDir` Directory containing `.log` files to stream (default `.`)
- `-corsOrigins` Comma-separated list of allowed CORS origins, or `*` (default `http://localhost:3000,https://anotherdomain.com,http://localhost:1972`)
- `-allowedIPs` Comma-separated list of allowed client IPs; empty disables whitelist
- `-jwtSecret` JWT secret for token validation

Examples:

- port  `go run . -port 5006`
- httpsPort `go run . -httpsPort 9443`
- disable HTTPS `go run . -enableHTTPS=false`
- custom TLS `go run . -tlsCert=/path/cert.pem -tlsKey=/path/key.pem`
- custom logs dir `go run . -logsDir=/var/log/streamer`
- allow all origins `go run . -corsOrigins=*`

#### OS Env

- `STREAMER_PORT`
- `STREAMER_HTTPS_PORT`
- `STREAMER_ENABLE_HTTPS` (true|false|1|0)
- `STREAMER_TLS_CERT`
- `STREAMER_TLS_KEY`
- `STREAMER_LOGS_DIR`
- `STREAMER_CORS_ORIGINS` (comma-separated or `*`)
- `STREAMER_ALLOWED_IPS` (comma-separated)
- `STREAMER_JWT_SECRET`

`STREAMER_PORT=5006 STREAMER_HTTPS_PORT=9443 STREAMER_ENABLE_HTTPS=false STREAMER_LOGS_DIR=/var/log go run .`

## Testing

The backend includes a suite of unit tests to verify the functionality of the API handlers. To run the tests, navigate to the `backend` directory and execute:

```sh
go test
```

### API

- `/alive` -> Returns `200 OK`.
- `/version` -> Returns the version and build date of the application.
- `/list-files` -> Returns a JSON array of available `.log` file names in the `logsDir`.
- `/stream-logs` -> Establishes a WebSocket connection to stream the content of a selected log file.
- `/stop` -> Sends a SIGTERM signal to gracefully shut down the application.
- `/restart` -> Restarts the application by spawning a new child process.

> **Note on `/stop` and `/restart`**: These endpoints provide a way to manage the agent's lifecycle via its API. While convenient for development, this is an unconventional pattern for production environments. It is generally recommended to manage the application process using a dedicated service manager like `systemd`, `supervisor`, or Docker.

### Logs

Logs are **streamer.log** in same path of binary app

Logs are available also on console by default

## Security

### Auth Token

https://www.scottbrady91.com/tools/jwt to generate online JWT token

Added JWT authentication (https://jwt.io/introduction), just for alive API for test.
You can set the JWT secret via `-jwtSecret` or `STREAMER_JWT_SECRET`.
Frontend should attach the header `Authorization: Bearer <token>`.

### CORS 

You can enable CORS for every client (`-corsOrigins=*`) or specify the allowed origins via flag or env.

### TODO
- Add configuration yaml file
- Add support for basic auth on api
- Add some tests
- Add structured logging configuration
- Add ability to choose initial tail offset
- Add path config validation and health check endpoint