# Dashboard

Web UI for YankaDevLab, currently showing live server metrics. Private, behind SSO.

## Stack

- React 19 + TypeScript, Vite
- Tailwind CSS v4 + shadcn/ui (Base UI), React Router
- Served as a static SPA by nginx

## Routes

| Path | Description                                                   |
| ---- | ------------------------------------------------------------- |
| `/`  | Home: CPU, RAM, disk and network cards, polled every 5 s      |
| `*`  | 404 page (nginx falls back to `index.html` for client routes) |

## Configuration

Build-time variables, passed as Docker build args in `docker-compose.yml`.

| Variable                  | Required   | Default | Description                                    |
| ------------------------- | ---------- | ------- | ---------------------------------------------- |
| `VITE_AUTH_API_URL`       | Yes (prod) | `""`    | Auth service base URL, used by the logout form |
| `VITE_MONITORING_API_URL` | Yes (prod) | `""`    | Monitoring API base URL for the metric cards   |

## Development

Conductor run script `dashboard` (Vite dev server on `$CONDUCTOR_PORT + 2`), or:

```bash
npm ci
npm run dev
```

The SPA has no auth code, so it runs without the auth service. Metric cards need the monitoring API and fail without it,
and logout is broken locally: with an empty `VITE_AUTH_API_URL` the form posts to the Vite dev server.

In production, a 401 from the monitoring API (expired session) reloads the page so that Traefik redirects to the login
page.

## Production

Deployed at `dashboard.yankadevlab.tech` behind Traefik, protected by the `auth-verify` ForwardAuth middleware (SSO).
Deployed by CI on push to `main`.
