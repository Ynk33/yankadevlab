# Assiette

Recipes and shopping list app, shared within a team. Own user accounts (no SSO), invite-only signup.

## Stack

- Go + chi, PostgreSQL (`assiette` schema, migrations run on startup)
- Single-file frontend (`web/index.html`), embedded in the binary. It also runs as a Claude artifact, with its own
  storage and an "add recipes with Claude" feature; served by this service it always uses the API and hides that feature
- Session cookie (`assiette_session`, 30 days, renewed when less than 15 days are left)

## Routes

| Method | Path                        | Description                                                           |
| ------ | --------------------------- | --------------------------------------------------------------------- |
| GET    | `/health`                   | Health check                                                          |
| GET    | `/`                         | App (CSP: only its own inline script, by hash; Google Fonts allowed)  |
| GET    | `/logo.svg`                 | Logo and favicon                                                      |
| POST   | `/login`                    | Sign in (5 req/min per client IP)                                     |
| POST   | `/signup`                   | Create an account from an invite (5 req/min per client IP)            |
| POST   | `/logout`                   | Sign out                                                              |
| GET    | `/api/state`                | Team state (selected recipes, shopping list, favourites, preferences) |
| PATCH  | `/api/state`                | Apply a JSON merge patch to the team state                            |
| GET    | `/api/custom`               | List the team's custom recipes                                        |
| POST   | `/api/custom`               | Add custom recipes                                                    |
| GET    | `/api/team`                 | List team members                                                     |
| GET    | `/api/me`                   | Current user settings: `{ lang }` (`fr` or `en`)                      |
| PUT    | `/api/me`                   | Update `lang`                                                         |
| POST   | `/api/invites`              | Create a 7-day single-use invite                                      |
| POST   | `/api/invites/{token}/join` | Move the current account to the invite's team                         |

`/api/*` routes require a valid session.

- **Invites** — the app shares them as `https://assiette.yankadevlab.tech/#invite=<token>`. An invalid, expired or
  already used invite returns 410.
- **Passwords** — 8 to 72 bytes.
- **Joining a team** — the account leaves its current team. If that team is left empty, it is deleted with its state and
  custom recipes.

## Configuration

| Variable       | Required | Default | Description                  |
| -------------- | -------- | ------- | ---------------------------- |
| `DATABASE_URL` | Yes      | —       | PostgreSQL connection string |
| `SERVER_PORT`  | No       | `8080`  | HTTP port                    |

## Development

Conductor run script `assiette`: starts a per-workspace PostgreSQL container, builds and runs the service on
`$CONDUCTOR_PORT`, and seeds `dev@local.test` / `devpassword`.

```bash
go test ./...
```

## Production

Deployed at `assiette.yankadevlab.tech` behind Traefik, by CI on push to `main`. Create the first account with:

```bash
docker compose exec assiette /seed <email> <password>
```

`/seed` creates a new team on every call and fails if the email already exists. It also moves the legacy single-user
data (`assiette.state`, custom recipes without a team) into that team. Further accounts join through invite links
created in the app.
