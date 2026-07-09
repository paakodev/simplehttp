# Chirpy API Reference

Base path: `/api`

---

## Health

### `GET /api/healthz`
Returns `200 OK` with plain text `"OK"` if the server is running.

**Response `200`**
```
OK
```

---

## Users

### `POST /api/users`
Create a new user.

**Body (JSON)**
| Field | Required | Description |
|---|---|---|
| `email` | yes | User email address |
| `password` | yes | Plain-text password |

**Response `201`**
```json
{
  "id": "uuid",
  "created_at": "2026-07-09T12:00:00Z",
  "updated_at": "2026-07-09T12:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

---

### `PUT /api/users`
Update the authenticated user's email and/or password.

**Auth:** `Authorization: Bearer <access_token>`

**Body (JSON)**
| Field | Required | Description |
|---|---|---|
| `email` | no | New email address |
| `password` | no | New plain-text password |

**Response `200`**
```json
{
  "id": "uuid",
  "created_at": "2026-07-09T12:00:00Z",
  "updated_at": "2026-07-09T12:00:01Z",
  "email": "user@example.com",
  "is_chirpy_red": false
}
```

---

## Auth

### `POST /api/login`
Authenticate a user and receive tokens.

**Body (JSON)**
| Field | Required | Description |
|---|---|---|
| `email` | yes | User email |
| `password` | yes | User password |

**Response `200`**
```json
{
  "id": "uuid",
  "created_at": "2026-07-09T12:00:00Z",
  "updated_at": "2026-07-09T12:00:00Z",
  "email": "user@example.com",
  "is_chirpy_red": false,
  "token": "<jwt_access_token>",
  "refresh_token": "<refresh_token>"
}
```

> `token` expires in 1 hour. `refresh_token` expires in 60 days.

---

### `POST /api/refresh`
Exchange a valid refresh token for a new access token.

**Auth:** `Authorization: Bearer <refresh_token>`  
**Body:** none

**Response `200`**
```json
{ "token": "<new_access_token>" }
```

---

### `POST /api/revoke`
Revoke a refresh token.

**Auth:** `Authorization: Bearer <refresh_token>`  
**Body:** none

**Response `204`** No Content

---

## Chirps

### `POST /api/chirps`
Create a new chirp.

**Auth:** `Authorization: Bearer <access_token>`

**Body (JSON)**
| Field | Required | Description |
|---|---|---|
| `body` | yes | Chirp text (max 140 characters) |

**Response `201`**
```json
{
  "id": "uuid",
  "created_at": "2026-07-09T12:00:00Z",
  "updated_at": "2026-07-09T12:00:00Z",
  "body": "Hello world!",
  "user_id": "uuid"
}
```

> Profanity filter replaces `kerfuffle`, `sharbert`, and `fornax` with `****`.

---

### `GET /api/chirps`
List chirps.

**Query parameters**
| Param | Required | Default | Description |
|---|---|---|---|
| `author_id` | no | — | Filter by user UUID |
| `sort` | no | `asc` | Sort order: `asc` or `desc` |
| `limit` | no | `100` | Max results (hard cap 1000) |
| `offset` | no | `0` | Pagination offset |

**Response `200`**
```json
[
  {
    "id": "uuid",
    "created_at": "2026-07-09T12:00:00Z",
    "updated_at": "2026-07-09T12:00:00Z",
    "body": "Hello world!",
    "user_id": "uuid"
  }
]
```

---

### `GET /api/chirps/{id}`
Get a single chirp by UUID.

**Response `200`**
```json
{
  "id": "uuid",
  "created_at": "2026-07-09T12:00:00Z",
  "updated_at": "2026-07-09T12:00:00Z",
  "body": "Hello world!",
  "user_id": "uuid"
}
```

---

### `DELETE /api/chirps/{chirpID}`
Delete a chirp. Only the chirp's author may delete it.

**Auth:** `Authorization: Bearer <access_token>`

**Response `204`** No Content

---

## Webhooks

### `POST /api/polka/webhooks`
Polka payment webhook. Upgrades a user to Chirpy Red on the `user.upgraded` event.

**Auth:** `Authorization: ApiKey <polka_api_key>`

**Body (JSON)**
```json
{
  "event": "user.upgraded",
  "data": { "user_id": "<uuid>" }
}
```

**Response `204`** No Content (also returned for unrecognised events)

---

## Admin

### `GET /admin/metrics`
Returns an HTML page with the current fileserver hit count.

**Response `200`** (`text/html`)
```html
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited 42 times!</p>
  </body>
</html>
```

### `POST /admin/reset`
Resets the hit counter **and deletes all users** (dev platform only).

**Response `200`** (`text/plain`)
```
Hits and users reset
```

> Returns `403 Forbidden` when not running on the `dev` platform.

---

## Error Responses

All errors return JSON with an `error` field:

```json
{ "error": "<description>" }
```

| Status | Meaning |
|---|---|
| `400` | Bad request / invalid input |
| `401` | Missing or invalid credentials |
| `403` | Forbidden (wrong user or wrong platform) |
| `404` | Resource not found |
| `500` | Internal server error |
