# YankaDevLab

A personal micro-services suite — self-hosted tools for everyday life.

## Infrastructure

| Component       | Details                                                        |
|-----------------|----------------------------------------------------------------|
| Server          | VPS Hostinger — Ubuntu 24.04.4 LTS                            |
| Resources       | 1 vCPU, 3.8 GB RAM, 48 GB disk, 2 GB swap                    |
| Access          | SSH (root)                                                     |
| Existing        | nginx + Hugo static site (personal CV at yannicktirand.xyz)   |
| Orchestration   | Docker 29.3 + Docker Compose v5.1                             |
| Reverse proxy   | Traefik (TBD)                                                  |
| SSL             | Let's Encrypt via certbot (auto-renewal)                       |
| Node.js         | v22.22 LTS                                                     |

## Tech Stack

| Layer    | Tech                   | Why                                                            |
|----------|------------------------|----------------------------------------------------------------|
| Backend  | **Go**                 | Lightweight, performant, ideal for single-core, high learning value |
| Frontend | **React + TypeScript** | Already mastered, no need to fight on every front              |
| Database | **PostgreSQL**         | Solid relational DB with JSON support for flexibility          |
| Containers | **Docker Compose**   | Simple, well-suited for single-node setups                     |
| AI/LLM   | **External APIs**     | Claude/OpenAI/Mistral — VPS lacks resources for local inference |

## Roadmap

### Phase 0 — Foundations

> The base layer everything else builds on.

- [x] Install Docker + Docker Compose on the VPS
- [x] Clean up legacy projects and configs
- [x] Upgrade OS to Ubuntu 24.04 LTS
- [x] Restore CV site (Hugo + nginx + HTTPS)
- [ ] Set up reverse proxy (Traefik) in front of nginx
- [ ] Structure the monorepo
- [ ] Shared auth system (JWT or session-based, TBD)
- [ ] Basic CI/CD pipeline (push → build → deploy)

### Phase 1 — Dashboard + Monitoring

> The central hub + first eyes on the server.

- [ ] **Dashboard** — Auth-protected web UI, single entry point for all services
- [ ] **Server monitoring** — System metrics (CPU, RAM, disk, network) with history
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

---

## Target Architecture (simplified)

```text
                    ┌─────────────┐
                    │  Internet   │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  Reverse    │
                    │  Proxy      │
                    │  (Traefik)  │
                    └──────┬──────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
   ┌──────▼──────┐ ┌──────▼──────┐ ┌───────▼─────┐
   │  Hugo       │ │  Dashboard  │ │  Micro-     │
   │  (CV site)  │ │  (React)    │ │  services   │
   │             │ │             │ │  (Go APIs)  │
   └─────────────┘ └──────┬──────┘ └───────┬─────┘
                          │                │
                   ┌──────▼────────────────▼──────┐
                   │         PostgreSQL            │
                   └──────────────────────────────┘
```

## Future Ideas (backlog)

- URL shortener
- Private paste bin
- Bookmarks / notes
- Automations (cron jobs with UI)
- Local LLM (if/when server gets a GPU upgrade)
