# SSE to JSON Courier

This is a simple server to forward Server-Sent Events (SSE) to a target URL as JSON.

## CLI

The CLI (mostly used for development purposes) starts the server and forwards the SSE to the target URL.

```bash
go run ./cli/cmd/main.go start --source https://dbname-username.turso.io/beta/listen\?table\=users --target http://localhost:3000 --auth your-auth-token
```

## Server

The server listens for incoming SSE and forwards them to the target URL.

```bash
go run main.go
```
