Sure — here’s the updated stack/plan with **Next.js** as the frontend (and we’ll drop React Router entirely).

---

## Frontend Tech Stack (Web) — Updated to Next.js

### Core

* **Next.js (App Router)** + **React 18**
* **TypeScript**
* **Node 20+** (local + CI)

### UI

* **Mantine** (`@mantine/core`, `@mantine/hooks`, `@mantine/form`, `@mantine/dates`, `@mantine/notifications`)
* **Tabler Icons** (pairs nicely with Mantine)

### Styling / Theming

* Mantine theme provider in `app/layout.tsx`
* Global styles via Mantine + optional `postcss` for resets

### Data Fetching & API Client

* **TanStack Query** (React Query) for server state:

  * caching, retries, request dedupe, invalidation after mutations
* A small typed `fetch` wrapper:

  * base URL from env (`NEXT_PUBLIC_API_BASE_URL`)
  * inject auth token
  * attach `X-Request-Id` / correlation header
* Optional (recommended): **Zod** for runtime validation of API responses

### Auth (pick one)

* **NextAuth.js** (Auth.js) if you want a standard auth solution quickly
* OR simple JWT flow (login → store token) if you want to keep the frontend thin
  *(For this project, simple JWT is totally fine.)*

### Form handling

* Mantine `useForm`
* Zod resolver (optional but clean)

### Testing

* **Vitest** + **React Testing Library** (component + page smoke tests)
* Optional: Playwright for e2e (nice to have)

---

## Backend Tech Stack (Go) — unchanged

* **Go 1.22+**
* **chi** router
* **PostgreSQL** + `pgx/pgxpool`
* Migrations: **goose** or **atlas**
* Auth: JWT (`golang-jwt/jwt/v5`)
* Logging: **zap**
* Metrics: Prometheus `/metrics`
* Tracing: OpenTelemetry
* Background: separate **worker** service (expiry jobs, waitlist, reconciliation)
* Redis: idempotency + rate limiting + short TTL data

---

## Infra / Supporting Tech — unchanged

### Local dev

* **Docker + docker-compose**

  * Postgres, Redis
  * optional: Prometheus, Grafana, Jaeger/Tempo

### CI/CD

* **GitHub Actions**

  * Frontend: install → lint → test → build
  * Backend: golangci-lint → tests → integration tests → build
  * Optional: k6 smoke load test (PR nightly)

### Deployment (options)

* **Fastest:** Fly.io
* **Simple:** Render
* **Max resume signal:** AWS (ECS Fargate + RDS Postgres + ElastiCache Redis + ALB)

---

## What changes in the “plan” because of Next.js?

### 1) Routing is file-based

Instead of React Router pages, you’ll have:

* `app/login/page.tsx`
* `app/availability/page.tsx`
* `app/bookings/page.tsx`
* `app/admin/page.tsx` (optional)

### 2) Environment variables

* Frontend uses `NEXT_PUBLIC_API_BASE_URL` to call your Go backend.
* Server-only secrets (if any) stay unprefixed.

### 3) Optional “BFF” pattern (only if you want)

You *can* have Next.js route handlers (`app/api/...`) proxy to Go.
But I’d recommend **calling Go directly** for simplicity + clearer system boundaries.

---
