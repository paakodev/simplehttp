# Chirpy

A simple REST API backend for a Twitter-like microblogging platform, built in Go. This project is the result of a guided backend development course on [boot.dev](https://boot.dev).

## What it does

Chirpy lets users register, log in, and post short messages called "chirps" (up to 140 characters). It exposes a JSON API with JWT-based authentication and PostgreSQL for persistence.

### Features

- User registration and login with Argon2id password hashing
- JWT access tokens + refresh token rotation (with revocation)
- Create, list, and delete chirps with bad-word filtering
- Chirpy Red premium user status via Polka payment webhooks
- File server for static assets
- Hit metrics and admin reset endpoint (dev mode)

## API Overview

For full endpoint details, parameters, and response examples, see [API.md](API.md).

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/healthz` | Health check |
| `POST` | `/api/users` | Register a new user |
| `PUT` | `/api/users` | Update user email/password |
| `POST` | `/api/login` | Log in, receive tokens |
| `POST` | `/api/refresh` | Refresh access token |
| `POST` | `/api/revoke` | Revoke refresh token |
| `POST` | `/api/chirps` | Create a chirp |
| `GET` | `/api/chirps` | List chirps |
| `GET` | `/api/chirps/{id}` | Get a single chirp |
| `DELETE` | `/api/chirps/{chirpID}` | Delete a chirp |
| `POST` | `/api/polka/webhooks` | Polka payment webhook |
| `GET` | `/admin/metrics` | View file server hit count |
| `POST` | `/admin/reset` | Reset metrics and database |

## Tech Stack

- **Go** standard library HTTP server
- **PostgreSQL** with [sqlc](https://sqlc.dev) for type-safe queries
- **JWT** (`golang-jwt/jwt`) for authentication
- **Argon2id** (`alexedwards/argon2id`) for password hashing
- **godotenv** for environment configuration

## Running locally

1. Copy `.env` from `dotenv.sample` and fill in your database URL and secrets.
2. Run database migrations from `sql/schema/`.
3. Start the server:

```sh
go run .
```

The server listens on port **8080**.
