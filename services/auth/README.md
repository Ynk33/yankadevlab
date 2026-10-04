# Auth

Single sign-on for YankaDevLab. Serves the login page and is called by Traefik (ForwardAuth) to protect private
services. Public.

## Stack

- Go + chi, PostgreSQL (`users` table, migrations run on startup)
- HS256 JWT stored in a `session` cookie on `COOKIE_DOMAIN` (7 days)

## Routes

| Method | Path      | Description                                                                |
| ------ | --------- | -------------------------------------------------------------------------- |
| GET    | `/health` | Health check                                                               |
| GET    | `/login`  | Login page (`rd` query param: where to go after login)                     |
| POST   | `/login`  | Check credentials, set the session cookie (5 req/min per client IP)        |
| POST   | `/logout` | Clear the session cookie and redirect to the login page                    |
| GET    | `/verify` | ForwardAuth: 200 with `X-User-Id` / `X-User-Email`, renews sessions > 24 h |

`/verify` redirects HTML requests without a valid session to the login page and answers 401 to the others. `rd` is only
followed for https URLs on `COOKIE_DOMAIN` or one of its subdomains. A session expires after 7 days without a visit to a
protected service.

## Configuration

| Variable               | Required | Default | Description                                  |
| ---------------------- | -------- | ------- | -------------------------------------------- |
| `DATABASE_URL`         | Yes      | —       | PostgreSQL connection string                 |
| `JWT_SECRET`           | Yes      | —       | Secret used to sign session tokens           |
| `COOKIE_DOMAIN`        | Yes      | —       | Session cookie domain (`.yankadevlab.tech`)  |
| `LOGIN_URL`            | Yes      | —       | Public URL of the login page                 |
| `DEFAULT_REDIRECT_URL` | Yes      | —       | Where to go after login when `rd` is invalid |
| `SERVER_PORT`          | No       | `8080`  | HTTP port                                    |

## Development

No local run script: the session cookie and redirects only work on `.yankadevlab.tech` over https.

```bash
go test ./...
```

## Production

Deployed at `auth.yankadevlab.tech` behind Traefik, by CI on push to `main`. A service becomes private when
`auth-verify` is in its Traefik router's middlewares (see the root README, Access).

Create a user with:

```bash
docker compose exec auth /seed <email> <password>
```

An existing email is left unchanged (its password is not updated), although the command still reports success. Emails
are case-sensitive, at seed and at login: sign in with the exact spelling used here.
