## Certificate

`openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert-streamer.pem -days 9999
`

password cert: password

self signed: IT - Italy - Turin - MyCompany

Go does not support encrypted private key for tls.
To decrypt:
openssl rsa -in key.pem -out decrypted_key.pem

---

### How to run

`go run .`

`go run . &`

### Config

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


### API

- /alive    -> return 200 OK
- /stop     -> send sigterm kill to application
- /restart  -> kill process e recreate child process
- /stream-logs -> api that return stream of selected log file (file must be under `logsDir`)
- /version -> return http code 200 and body version and buildDate program
- /list-files -> return json with available .log file name 

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