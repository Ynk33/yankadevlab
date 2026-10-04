# YankaDevLab

A personal micro-services suite — self-hosted tools for everyday life.

## Infrastructure

| Component     | Details                                                                                                     |
| ------------- | ----------------------------------------------------------------------------------------------------------- |
| Server        | VPS Hostinger — Ubuntu 24.04.4 LTS                                                                          |
| Resources     | 1 vCPU, 3.8 GB RAM, 48 GB disk, 2 GB swap                                                                   |
| Access        | SSH (key-based)                                                                                             |
| Existing      | Hugo static site (personal CV), built outside this repo and served from `/var/www/yannicktirand.xyz/public` |
| Orchestration | Docker 29.3 + Docker Compose v5.1                                                                           |
| Reverse proxy | Traefik v3.6                                                                                                |
| SSL           | Let's Encrypt via Traefik ACME (auto-renewal)                                                               |
| CI/CD         | GitHub Actions (push to main → auto-deploy; base images of built services pulled on each deploy)            |
| Database      | PostgreSQL 16 (shared instance, internal network)                                                           |
| Build images  | Go 1.26, Node 22 (multi-stage Docker builds, no toolchain needed on the VPS)                                |

## Tech Stack

| Layer      | Tech                   | Why                                                                 |
| ---------- | ---------------------- | ------------------------------------------------------------------- |
| Backend    | **Go**                 | Lightweight, performant, ideal for single-core, high learning value |
| Frontend   | **React + TypeScript** | Already mastered, no need to fight on every front                   |
| Database   | **PostgreSQL**         | Solid relational DB with JSON support for flexibility               |
| Containers | **Docker Compose**     | Simple, well-suited for single-node setups                          |
| AI/LLM     | **External APIs**      | Claude/OpenAI/Mistral — VPS lacks resources for local inference     |

## Roadmap

### Phase 0 — Foundations

> The base layer everything else builds on.

- [x] Install Docker + Docker Compose on the VPS
- [x] Clean up legacy projects and configs
- [x] Upgrade OS to Ubuntu 24.04 LTS
- [x] Restore CV site (Hugo + HTTPS)
- [x] Set up reverse proxy (Traefik) — replaces nginx
- [x] Structure the monorepo
- [x] CI/CD pipeline (GitHub Actions: push to main → auto-deploy)

### Phase 1 — Dashboard + Monitoring

> The central hub + first eyes on the server.

- [x] **Shared auth system** — JWT session cookie, single sign-on for all services via Traefik ForwardAuth
  - [x] Add PostgreSQL to Docker Compose (shared instance, internal network)
  - [x] Auth service scaffold (Go + chi, health endpoint, multi-stage Dockerfile)
  - [x] Wire auth service into Docker Compose + Traefik routing
  - [x] Database connection + migrations (`users` table)
  - [x] Seed users (`/seed` CLI in the image — see the [auth README](services/auth/README.md))
  - [x] Config management (env-based)
  - [x] GET/POST /login (server-rendered login page, sets a 7-day `session` JWT cookie on `.yankadevlab.tech`)
  - [x] Structured logging (slog)
  - [x] POST /logout (clears the session cookie)
  - [x] GET /verify (Traefik ForwardAuth endpoint, renews the session cookie once it is older than 24 h)
  - [x] Traefik ForwardAuth middleware config
  - [x] Rate limiting on /login (per client IP behind Traefik)
  - [x] Drop the refresh-token flow (replaced by the sliding session cookie)
- [x] **Dashboard** — Web UI behind SSO, currently showing live server metrics
  - [x] Scaffold (Vite + React 19 + TypeScript, Tailwind CSS v4, shadcn/ui)
  - [x] Protected by the Traefik `auth-verify` middleware (no auth code in the SPA)
  - [x] Layout (collapsible sidebar, header, logout through the auth service)
  - [x] Docker setup (multi-stage build → nginx, Traefik routing on dashboard.yankadevlab.tech)
  - [x] Dark mode toggle
  - [x] 404 page
  - [ ] Links to the other services (single entry point)
- [x] **Server monitoring** — System metrics (CPU, RAM, disk, network) with history
  - [x] Prometheus + node-exporter (5 s scrape, 15-day retention)
  - [x] Monitoring API (Go) exposing live and historical CPU, RAM, disk and network metrics, behind SSO
  - [x] Live metric cards on the dashboard (5 s polling)
  - [x] History (charts over time, 1h / 24h / 7d)
- [ ] **Homemade analytics** — Lightweight visit tracking for the public-facing site (simplified Plausible/Umami)

### Phase 2 — Subscription Tracker

> No more surprise charges on the bank account.

- [ ] Email inbox connection (IMAP or provider API)
- [ ] Parse confirmation/billing emails (LLM via external API)
- [ ] Active subscriptions dashboard with amounts, renewal dates
- [ ] Alerts before renewal / end of free trial

### Phase 3 — Curated News & Watch

> A personalized, sourced, AI-sorted news feed.

- [ ] Configure interests (topics, keywords, sources)
- [ ] Scraping / aggregation from multiple sources (RSS, websites, APIs)
- [ ] Relevance scoring and sorting via LLM (external API)
- [ ] Reading interface with sources, summaries, and filters

### Standalone Apps

- [x] **Assiette** — Recipes and shopping list shared within a team, own accounts
      ([README](services/assiette/README.md))

---

## Architecture

```text
                         ┌─────────────┐
                         │  Internet   │
                         └──────┬──────┘
                                │
                         ┌──────▼──────┐   ForwardAuth   ┌─────────────┐
                         │   Traefik   │────────────────▶│    auth     │
                         └──────┬──────┘   (/verify)     │    (Go)     │
                                │                        └──────┬──────┘
       ┌────────────┬───────────┼─────────────┐                 │
       │            │           │             │                 │
┌──────▼─────┐ ┌────▼──────┐ ┌──▼─────────┐ ┌─▼──────────┐      │
│  cv-site   │ │ dashboard │ │ monitoring │ │  assiette  │      │
│  (Hugo,    │ │ (React,   │ │   (Go)     │ │   (Go +    │      │
│   nginx)   │ │  nginx)   │ │            │ │   React)   │      │
│            │ │           │ │            │ │            │      │
└────────────┘ └───────────┘ └──────┬─────┘ └──────┬─────┘      │
                                    │              │            │
                             ┌──────▼─────┐ ┌──────▼────────────▼──┐
                             │ Prometheus │ │      PostgreSQL      │
                             └──────┬─────┘ └──────────────────────┘
                             ┌──────▼────────┐
                             │ node-exporter │
                             └───────────────┘
```

The dashboard calls the monitoring API from the browser; it never talks to PostgreSQL.

## Repository Layout

| Path           | Content                                                          |
| -------------- | ---------------------------------------------------------------- |
| `services/`    | One directory per service (own Dockerfile, Go module or npm app) |
| `infra/`       | Traefik static config and Prometheus scrape config               |
| `shared/`      | Reserved for shared code (empty for now)                         |
| `.conductor/`  | Conductor setup and run scripts for local development            |
| `.github/`     | Deploy workflow (GitHub Actions)                                 |
| `.env.example` | Template of the production `.env`                                |

Each service has its own README with the same sections: Stack, Routes, Configuration, Development, Production —
[auth](services/auth/README.md), [dashboard](services/dashboard/README.md), [monitoring](services/monitoring/README.md),
[assiette](services/assiette/README.md).

## Local Development

Only Assiette and the dashboard run locally, through Conductor run scripts — see their READMEs. The full stack depends
on Traefik, Let's Encrypt and the `.yankadevlab.tech` SSO cookie. Prerequisites: Docker, Go 1.26, Node 22.

## Production Setup

- **Server** — the repository is cloned in `/opt/yankadevlab`, with a `.env` file next to `docker-compose.yml`
  (template: `.env.example`): `ACME_EMAIL`, `TRAEFIK_DASHBOARD_AUTH`, `POSTGRES_USER`, `POSTGRES_PASSWORD`,
  `POSTGRES_DB`, `JWT_SECRET`.
- **DNS** — every host in the Access table needs a record pointing to the VPS (Let's Encrypt HTTP challenge).
- **Deploy** — on push to `main`, `.github/workflows/deploy.yml` connects over SSH and runs `git pull origin main`,
  `docker compose build --pull` and `docker compose up -d`. GitHub secrets: `SERVER_HOST`, `SERVER_USER`,
  `SSH_PRIVATE_KEY`.
- **Image updates** — `build --pull` only refreshes the base images of built services. `traefik`, `postgres`,
  `prometheus`, `node-exporter` and `cv-site` (`nginx:alpine`) keep their cached image, including the `:latest` tags,
  until `docker compose pull` is run on the server.

## Access

A service is private when `auth-verify` is in its Traefik router's middlewares in `docker-compose.yml`. Private services
redirect to the login page served by the auth service; the resulting `session` cookie on `.yankadevlab.tech` is shared
by every subdomain (SSO). For an API called from another origin, the CORS middleware must come before `auth-verify`
(e.g. `middlewares=monitoring-cors,auth-verify`) so that preflight requests are answered.

| Service    | Host                                   | Access                |
| ---------- | -------------------------------------- | --------------------- |
| auth       | auth.yankadevlab.tech                  | Public (login page)   |
| dashboard  | dashboard.yankadevlab.tech             | Private (SSO)         |
| monitoring | monitoring.yankadevlab.tech            | Private (SSO)         |
| assiette   | assiette.yankadevlab.tech              | Own accounts (no SSO) |
| cv-site    | yannicktirand.xyz / .fr / .com (+ www) | Public                |
| traefik    | traefik.yankadevlab.tech               | Basic auth            |

## Future Ideas (backlog)

- URL shortener
- Private paste bin
- Bookmarks / notes
- Automations (cron jobs with UI)
- Local LLM (if/when server gets a GPU upgrade)
