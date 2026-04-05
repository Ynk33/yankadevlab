# Dashboard

Auth-protected web UI for YankaDevLab — single entry point for all services.

## Stack

- React 19 + TypeScript
- Vite (build + dev server)
- Tailwind CSS v4 + shadcn/ui (Base UI)
- React Router

## Dev

```bash
npm install
npm run dev
```

Login is bypassed in dev mode (no auth API needed).

## Production

Built as a static SPA and served by nginx via Docker.
Deployed at `dashboard.yankadevlab.tech` behind Traefik (HTTPS).

```bash
docker compose up -d --build dashboard
```
