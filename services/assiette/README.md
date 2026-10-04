# Assiette

Recipes and shopping list app, shared within a team. Own user accounts (no SSO), invite-only signup.

## Stack

- Go + chi, PostgreSQL (`assiette` schema, migrations run on startup)
- Frontend in `web/`: React 19 + TypeScript, Vite, Tailwind CSS v4 + shadcn/ui (Base UI). The Vite build
  (`web/dist`) is embedded in the binary
- Session cookie (`assiette_session`, 30 days, renewed when less than 15 days are left)

## Routes

| Method | Path                         | Description                                                           |
| ------ | ---------------------------- | --------------------------------------------------------------------- |
| GET    | `/health`                    | Health check                                                          |
| GET    | `/*`                         | Frontend build (CSP: same origin only, fonts self-hosted)             |
| POST   | `/login`                     | Sign in (5 req/min per client IP)                                     |
| POST   | `/signup`                    | Create an account from an invite (5 req/min per client IP)            |
| POST   | `/logout`                    | Sign out                                                              |
| GET    | `/api/state`                 | Team state (selected recipes, shopping list, favourites, preferences) |
| PATCH  | `/api/state`                 | Apply a JSON merge patch to the team state                            |
| GET    | `/api/custom`                | List the team's custom recipes                                        |
| GET    | `/api/recipes/{id}/comments` | List the team's comments on a recipe (oldest first)                   |
| POST   | `/api/recipes/{id}/comments` | Add a comment (201)                                                   |
| PUT    | `/api/comments/{cid}`        | Edit one of your own comments (404 otherwise)                         |
| DELETE | `/api/comments/{cid}`        | Delete one of your own comments (404 otherwise)                       |
| GET    | `/api/team`                  | List team members                                                     |
| GET    | `/api/me`                    | Current user settings: `{ lang }` (`fr` or `en`)                      |
| PUT    | `/api/me`                    | Update `lang`                                                         |
| POST   | `/api/invites`               | Create a 7-day single-use invite                                      |
| POST   | `/api/invites/{token}/join`  | Move the current account to the invite's team                         |

`/api/*` routes require a valid session.

- **Invites** — the app shares them as `https://assiette.yankadevlab.tech/#invite=<token>`. An invalid, expired or
  already used invite returns 410.
- **Passwords** — 8 to 72 bytes.
- **Comments** — scoped to the team, 1 to 2000 characters. Each comment carries `author` (email) and `mine`; only the
  author can edit or delete it.
- **Joining a team** — the account leaves its current team. If that team is left empty, it is deleted with its state,
  custom recipes and comments.

## Configuration

| Variable       | Required | Default | Description                  |
| -------------- | -------- | ------- | ---------------------------- |
| `DATABASE_URL` | Yes      | —       | PostgreSQL connection string |
| `SERVER_PORT`  | No       | `8080`  | HTTP port                    |

## Development

Conductor run script `assiette`: starts a per-workspace PostgreSQL container, builds and runs the service on
`$CONDUCTOR_PORT + 3`, seeds `dev@local.test` / `devpassword`, then starts the Vite dev server on `$CONDUCTOR_PORT`. Vite
proxies `/api`, `/login`, `/signup` and `/logout` to the service (`ASSIETTE_API_URL`, default `http://localhost:8080`).

```bash
go test ./...
cd web && npm ci && npm run lint && npm test && npm run build
```

`go build` works without a frontend build (`web/dist` only holds a `.gitkeep`), but the binary then serves no app.
Static files under `/assets/` are cached for a year (hashed names); the rest is revalidated on every request.

## Production

Deployed at `assiette.yankadevlab.tech` behind Traefik, by CI on push to `main`. Create the first account with:

```bash
docker compose exec assiette /seed <email> <password>
```

`/seed` creates a new team on every call and fails if the email already exists. It also moves the legacy single-user
data (`assiette.state`, custom recipes without a team) into that team. Further accounts join through invite links
created in the app.
