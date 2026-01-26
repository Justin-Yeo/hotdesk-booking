# Hotdesk Booking System - Complete Implementation Plan

**Version**: 1.0
**Date**: 2025-01-25
**Status**: ALL DECISIONS FINAL

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Tech Stack](#tech-stack)
3. [Core Data Model & Booking Engine](#core-data-model--booking-engine)
4. [Policy Engine](#policy-engine)
5. [Authentication & Authorization](#authentication--authorization)
6. [Frontend/UX Details](#frontendux-details)
7. [Backend Architecture](#backend-architecture)
8. [Deployment & Operations](#deployment--operations)
9. [Testing Strategy](#testing-strategy)
10. [MVP Scope & Success Criteria](#mvp-scope--success-criteria)
11. [Technical Implementation Details](#technical-implementation-details)
12. [Development Workflow](#development-workflow)
13. [Phase 2 Features](#phase-2-features)
14. [Out of Scope](#out-of-scope)

---

## Project Overview

**Purpose**: Hot desk booking system for shared workspaces with conflict detection, check-in enforcement, and admin management.

**Target Scale**: 50-200 desks, ~50 concurrent users, ~200 bookings/day, ~20 peak concurrent booking attempts.

**Performance Targets**:
- Booking creation: <500ms
- Zero double-bookings (100% conflict detection accuracy)
- 80% backend test coverage

---

## Tech Stack

### Frontend
- **Framework**: Next.js 14+ (App Router)
- **Language**: TypeScript (strict mode)
- **UI Library**: shadcn/ui + Tailwind CSS
- **State Management**: Zustand (global), TanStack Query (server state)
- **Forms**: React Hook Form + Zod validation
- **Date/Time**: date-fns
- **HTTP Client**: Fetch API (via TanStack Query)
- **Testing**: Vitest + React Testing Library + Playwright (E2E)
- **Dark Mode**: Auto-detect system preference

### Backend
- **Language**: Go 1.22+
- **Router**: chi
- **Database**: PostgreSQL 16+ (via pgx/pgxpool)
- **Migrations**: goose or atlas
- **Auth**: JWT (golang-jwt/jwt/v5)
- **Redis**: Idempotency, rate limiting, session storage
- **Background Jobs**: Separate worker service (Asynq)
- **Logging**: zap (structured JSON)
- **Metrics**: Prometheus /metrics endpoint
- **Tracing**: OpenTelemetry
- **API Format**: REST with /api/v1/ versioning

### Infrastructure
- **Deployment**: Render (managed platform)
- **Database**: Render Postgres (automated daily + weekly backups for 30 days)
- **Redis**: Render Redis (ephemeral, no persistence)
- **CI/CD**: GitHub Actions
- **Environments**: Staging + Production
- **Domain**: Render subdomain (e.g., hotdesk.onrender.com)
- **SSL**: Render-managed (Let's Encrypt)
- **Error Tracking**: Sentry (free tier)
- **Monitoring**: Structured logs → Render aggregation

### Development Tools
- **Monorepo**: Single repo, /frontend and /backend directories
- **Local Dev**: Docker Compose (Postgres + Redis)
- **Package Manager**: npm (frontend), go modules (backend)
- **Code Quality**: Prettier, ESLint, pre-commit hooks (Husky + lint-staged)
- **Commit Convention**: Conventional Commits + commitlint
- **Branch Naming**: feature/*, bugfix/*, hotfix/*

---

## Core Data Model & Booking Engine

### 1.1 Resources & Capacity

**Resource Types**: Desks only (no meeting rooms in MVP)

**Desk Attributes**:
- `id` (primary key)
- `desk_number` (e.g., "E-101", "W-205")
- `wing` (enum: "East", "West")
- `status` (enum: "available", "maintenance")
- `created_at`, `updated_at`

**Capacity**: 50-200 desks

**Schema Design** (extendable for Phase 2):
```sql
CREATE TYPE wing_type AS ENUM ('East', 'West');
CREATE TYPE desk_status AS ENUM ('available', 'maintenance');

CREATE TABLE desks (
  id SERIAL PRIMARY KEY,
  desk_number VARCHAR(20) NOT NULL UNIQUE,
  wing wing_type NOT NULL,
  status desk_status NOT NULL DEFAULT 'available',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_desks_wing ON desks(wing);
CREATE INDEX idx_desks_status ON desks(status);
```

---

### 1.2 Time Granularity & Constraints

**Booking Slots**: Fixed 30-minute blocks (8:00, 8:30, 9:00, ..., 21:30, 22:00)

**Duration Constraints**:
- **Minimum**: 30 minutes (1 slot)
- **Maximum**: 12 hours (24 slots)

**Opening Hours**: 8:00 AM - 10:00 PM daily (14 hours, 28 slots/day)

**Buffer Time**: None (back-to-back bookings allowed)

**Early Release**: Users can end bookings early; desk released immediately (no credits)

**Time Zone**: Single timezone (assume all users in same timezone, store times in UTC)

---

### 1.3 Booking Window (Rolling Lookahead)

**Rule**: Users can book within **current week + next week** only.

**Week Definition**: Monday 00:00 - Sunday 23:59

**Implementation Logic**:
```go
// getCurrentBookableRange returns the earliest and latest bookable times
func getCurrentBookableRange(now time.Time) (start, end time.Time) {
    // Start: now (can't book in the past)
    start = now

    // End: Sunday 23:59:59 of next week
    currentWeekStart := startOfWeek(now) // Monday 00:00 of current week
    nextWeekEnd := endOfWeek(currentWeekStart.Add(7 * 24 * time.Hour)) // Sunday 23:59:59 of next week
    end = nextWeekEnd

    return start, end
}

// Example: Today is Thu Jan 16, 2025
// - Current week: Mon Jan 13 - Sun Jan 19
// - Next week: Mon Jan 20 - Sun Jan 26
// - Bookable: Thu Jan 16 (now) → Sun Jan 26 23:59
```

**Background Job**: Daily materialization job runs at 00:05 to expand booking window.

---

### 1.4 Recurring Bookings (Phase 2)

**Patterns**: Weekly with custom days (e.g., "Every Tue/Thu 2-4pm")

**Max Duration**: Limited by booking window (can only create instances within current + next week)

**Conflict Handling**: All-or-nothing (reject entire series if any instance conflicts)

**Storage**: Expand instances on creation (create individual booking rows)

**Cancellation**: Both options supported (cancel single instance or entire series)

**Implementation** (Phase 2):
- Store recurrence rule in `recurring_bookings` table
- On creation, expand to individual `bookings` rows
- Link via `recurring_booking_id` foreign key

---

### 1.5 Booking Modification & Cancellation

**Modification**: Full modification allowed (time, date, desk)

**Cancellation Cutoff**: None (users can cancel anytime, even 1 minute before)

**Modification Flow**:
1. User requests modification (new time/desk)
2. System validates new time/desk (conflict check)
3. If conflict → reject + suggest alternatives (same time, different desk priority)
4. If valid → update booking + create audit log entry

**Cancellation Flow**:
1. User cancels booking
2. Soft delete (set `status = 'cancelled'`, `cancelled_at = NOW()`)
3. Desk immediately available for rebooking
4. Create audit log entry

---

### 1.6 Conflict Detection Strategy

**Approach**: Hybrid (PostgreSQL exclusion constraints + app-level validation)

**Flow**:
1. **App validation**: Check availability before insert (fast fail)
2. **DB insert**: Attempt insert with exclusion constraint (DB-level guarantee)
3. **Constraint violation**: Catch error, generate suggestions (alternative desks/times)

**Exclusion Constraint**:
```sql
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE bookings (
  id SERIAL PRIMARY KEY,
  desk_id INTEGER NOT NULL REFERENCES desks(id),
  user_id INTEGER NOT NULL REFERENCES users(id),
  start_time TIMESTAMPTZ NOT NULL,
  end_time TIMESTAMPTZ NOT NULL,
  time_range TSTZRANGE GENERATED ALWAYS AS (tstzrange(start_time, end_time)) STORED,
  status VARCHAR NOT NULL DEFAULT 'confirmed', -- 'confirmed', 'cancelled', 'completed', 'no_show'
  checked_in_at TIMESTAMPTZ,
  actual_end_time TIMESTAMPTZ, -- For early release
  cancelled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT no_overlapping_bookings
    EXCLUDE USING GIST (
      desk_id WITH =,
      time_range WITH &&
    ) WHERE (status NOT IN ('cancelled', 'no_show'))
);

CREATE INDEX idx_bookings_desk_id ON bookings(desk_id);
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_time_range ON bookings USING GIST(time_range);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_start_time ON bookings(start_time); -- For no-show queries
```

**Concurrency Strategy**: Optimistic (let DB constraint catch conflicts, provide good error messages)

**Idempotency**: Postgres table storage (24hr TTL) with UUID keys in request headers

---

## Policy Engine

### 2.1 Booking Limits & Fairness

**Daily Limit**: Max 10 hours per user per day (enforced at booking creation)

**Weekly Limit**: None

**Concurrent Future Bookings**: No limit

**Enforcement**: Hard limits (booking rejected if exceeds 10hr/day)

**Recurring Bookings** (Phase 2): Each instance counts toward daily limit when it occurs

---

### 2.2 Advance Booking Window

**Earliest**: Current week + next week (rolling window, see section 1.3)

**Latest**: Same-day booking allowed as long as time hasn't passed

**Validation**:
```go
func validateBookingWindow(startTime time.Time, now time.Time) error {
    bookableStart, bookableEnd := getCurrentBookableRange(now)

    if startTime.Before(now) {
        return errors.New("cannot book in the past")
    }

    if startTime.After(bookableEnd) {
        return errors.New("booking exceeds 1-week lookahead window")
    }

    return nil
}
```

---

### 2.3 Check-In & No-Show Policy

**Check-In Window**: 15 minutes before start time → 15 minutes after start time (30min total)

**Grace Period**: 15 minutes after start time

**No-Show Detection**: Booking auto-cancelled if no check-in within grace period

**No-Show Tracking**: Metrics only (visible to admins, no penalties)

**Desk Release**: Immediately after grace period expires (desk becomes available)

**Check-In Button Behavior**:
- **Before check-in window**: Disabled, show "Check-in available at [start_time - 15min]"
- **Within window**: Enabled
- **After grace period**: Booking auto-cancelled, button hidden

**Background Job (Worker Service)**:
```go
// Runs every 1 minute
func ProcessNoShows(ctx context.Context) error {
    now := time.Now()
    gracePeriodCutoff := now.Add(-15 * time.Minute)

    // Find bookings that should be cancelled
    query := `
        SELECT id FROM bookings
        WHERE status = 'confirmed'
          AND checked_in_at IS NULL
          AND start_time <= $1
    `

    rows, err := db.Query(ctx, query, gracePeriodCutoff)
    // ... iterate and cancel in transactions (idempotent)

    for _, bookingID := range bookingIDs {
        err := db.BeginTx(ctx, func(tx) error {
            // Re-check status (idempotency)
            var currentStatus string
            tx.QueryRow("SELECT status FROM bookings WHERE id = $1 FOR UPDATE", bookingID).Scan(&currentStatus)

            if currentStatus != "confirmed" {
                return nil // Already processed
            }

            // Update status
            _, err := tx.Exec(`
                UPDATE bookings
                SET status = 'no_show', updated_at = NOW()
                WHERE id = $1
            `, bookingID)

            // Create audit log entry
            // ...

            return err
        })
    }
}
```

---

### 2.4 Admin Override Capabilities

**Force Cancel**: Admins can cancel any booking (reason field required)

**Maintenance Blocking**:
- Admins can set desk `status = 'maintenance'`
- System auto-cancels all conflicting bookings
- Notifications sent to affected users (in-app only)

**Policy Exceptions**: None (all users follow same policies)

**Approval Workflow**: None (all bookings instant, first-come first-served)

---

## Authentication & Authorization

### 3.1 Auth Implementation

**Approach**: Manual JWT implementation (full control)

**Registration**: Self-service (anyone can sign up with email + password)

**Password Requirements**: Minimum 8 characters (no complexity requirements)

**Password Hashing**: bcrypt (cost factor 10-12)

**Password Reset**: Admin-managed (no email flow for MVP)

---

### 3.2 Roles & Permissions

**Roles**:
- **Member**: Regular user (book/cancel own bookings, view availability)
- **Admin**: Full control (manage desks, view all bookings, cancel any booking, manage users)

**Permission Model**: Simple role-based (no resource-based permissions)

**Other Users' Bookings**: Members see "Occupied" only (no user details)

**Super Admin**: No (all admins have equal privileges)

---

### 3.3 Session Management

**Access Token**:
- Lifetime: 15 minutes
- Storage: Client-side (localStorage or httpOnly cookie)
- Claims: `user_id`, `role`, `exp`, `iat`

**Refresh Token**:
- Lifetime: 7 days
- Storage: Postgres table `sessions` (not Redis, survives restarts)
- Rotation: Issue new refresh token on each refresh

**Concurrent Sessions**: Unlimited (users can login on multiple devices)

**Remember Me**: No (always use standard token lifetime)

**Logout Everywhere**: Yes (invalidate all refresh tokens for user)

**Schema**:
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'member', -- 'member', 'admin'
  status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'disabled'
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at); -- For cleanup
```

---

## Frontend/UX Details

### 4.1 Availability Search & Display

**Primary Flow**: Desk-first view

1. User browses desks (list or grid)
2. Filters: Wing (East/West), desk number search
3. User clicks desk → see calendar view for that desk
4. User picks time slot → create booking

**Availability Visualization**: Calendar grid (like Google Calendar)
- Time slots (30min blocks) down the left
- Days across the top (current week + next week)
- Color-coded cells: Green (available), Gray (occupied)
- Click available slot → booking wizard

**Search/Filter Options**:
- Filter by wing (East/West dropdown)
- Search by desk number (text input, e.g., "E-101")
- Future: Filter by equipment (standing desk, monitor) - schema designed for this

---

### 4.2 Booking Creation Flow

**Multi-Step Wizard**:

**Step 1: Select Date & Time**
- Date picker (limited to bookable range)
- Start time dropdown (30min slots, 8:00-21:30)
- Duration dropdown (30min to 12hrs)
- "Next" button

**Step 2: Select Desk**
- Show available desks for selected time
- List view with wing labels
- Highlight if approaching 10hr daily limit (warning, not blocking)
- "Next" button

**Step 3: Review & Confirm**
- Summary: Date, time, duration, desk
- Policy warnings (if close to 10hr limit)
- "Confirm Booking" button

**On Success**: Redirect to My Bookings, show success notification

**On Conflict**: Show error + suggestions (alternative desks for same time prioritized, then alternative times)

**Recurring Bookings**: Phase 2 (not in MVP)

---

### 4.3 My Bookings View

**View Type**: Calendar view (default)
- Weekly calendar showing user's bookings
- Color-coded by status (upcoming: blue, past: gray, cancelled: red)
- Click event → detail modal

**Actions Available**:
- **Check In**: If within check-in window (15min before to 15min after start)
- **Cancel Booking**: Anytime
- **Modify Booking**: Opens edit form (same as creation wizard, pre-filled)
- **End Booking Early**: If currently active (start time passed, end time not reached)
- **View Details**: Show full booking info + audit log (who created, modified, etc.)

**Quick Rebook**: "Book Again" button on past bookings (copies time/desk, opens creation wizard)

---

### 4.4 Admin Dashboard

**Priority 1 (MVP Must-Have)**:

1. **Desk Management**
   - CRUD interface (table view with edit/delete actions)
   - Create desk form: desk number, wing
   - Mark for maintenance (checkbox → sets `status = 'maintenance'`, auto-cancels conflicts)

2. **System Settings**
   - Configure opening hours (start/end time)
   - Configure daily hour limit (currently 10hrs)
   - Configure check-in grace period (currently 15min)

3. **User Management**
   - View all users (table: email, role, status, created_at)
   - Disable accounts (set `status = 'disabled'`)
   - Reset passwords (manual, admin sets new password)

**Priority 2 (Nice-to-Have MVP)**:

4. **View All Bookings**
   - Table view with filters: date range, user (search), desk, status
   - Pagination (20 per page, offset/limit)
   - Click row → detail view

5. **Force Cancel Bookings**
   - Cancel button on booking detail view
   - Reason field (required, stored in audit log)
   - Notification sent to user (in-app)

6. **Audit Log**
   - Table view: timestamp, user, action, entity_type, entity_id, changes_json
   - Filters: date range, user, action type
   - Read-only view

**Phase 2 (Post-MVP)**:
- Utilization metrics (desk usage %, peak times)
- No-show report (users with most no-shows)

---

### 4.5 Notifications

**Delivery**: In-app only (notification center in UI)

**Types**:
- Booking confirmed (on creation)
- Booking cancelled (by user or admin)
- Booking modified (time/desk changed)
- Booking auto-cancelled (no-show after grace period)

**Email Notifications**: Phase 2

**Implementation**: Simple notifications table + polling or WebSocket

```sql
CREATE TABLE notifications (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type VARCHAR(50) NOT NULL, -- 'booking_confirmed', 'booking_cancelled', etc.
  message TEXT NOT NULL,
  read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_read ON notifications(read);
```

---

### 4.6 Accessibility & Responsiveness

**Accessibility**: Best-effort (semantic HTML, keyboard navigation, ARIA labels)
- No formal WCAG testing for MVP
- Use shadcn components (built with accessibility in mind)

**Responsiveness**: Desktop-first, mobile-responsive
- All key flows work on mobile:
  - Browse availability ✅
  - Create booking ✅
  - View my bookings ✅
  - Cancel booking ✅
  - Check in ✅
  - Modify booking ✅

**Dark Mode**: Auto-detect system preference (Next.js + Tailwind dark mode)

**Mobile Optimization**: Phase 2 (mobile-first redesign)

---

## Backend Architecture

### 5.1 API Design

**Style**: REST

**Versioning**: `/api/v1/...` from day 1

**Response Format**: Detailed envelope
```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "timestamp": "2025-01-25T10:00:00Z",
    "request_id": "uuid"
  }
}
```

**Pagination**: Offset/limit (`?page=1&limit=20`)

**Error Response Format**:
```json
{
  "success": false,
  "error": "BOOKING_CONFLICT",
  "message": "Desk E-101 is already booked 2:00pm-4:00pm",
  "details": {
    "conflicting_booking_id": 123
  },
  "meta": { ... }
}
```

**Conflict Error with Suggestions**:
```json
{
  "success": false,
  "error": "BOOKING_CONFLICT",
  "message": "Desk E-101 is unavailable",
  "suggestions": {
    "alternative_desks": [
      { "desk_id": 105, "desk_number": "E-105", "wing": "East" },
      { "desk_id": 203, "desk_number": "W-203", "wing": "West" }
    ],
    "alternative_times": [
      { "start_time": "16:00", "end_time": "18:00" }
    ]
  }
}
```

---

### 5.2 API Endpoints (Key Routes)

**Auth**:
```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/logout-all
```

**Bookings**:
```
GET    /api/v1/bookings          # My bookings (user) or all bookings (admin)
POST   /api/v1/bookings          # Create booking
GET    /api/v1/bookings/:id
PATCH  /api/v1/bookings/:id      # Modify booking
DELETE /api/v1/bookings/:id      # Cancel booking
POST   /api/v1/bookings/:id/check-in
POST   /api/v1/bookings/:id/end-early
```

**Desks**:
```
GET    /api/v1/desks             # List all desks (with availability status)
GET    /api/v1/desks/:id
GET    /api/v1/desks/:id/availability?start=...&end=...
POST   /api/v1/desks             # Admin only
PATCH  /api/v1/desks/:id         # Admin only
DELETE /api/v1/desks/:id         # Admin only
```

**Admin**:
```
GET    /api/v1/admin/users
PATCH  /api/v1/admin/users/:id   # Disable, reset password
GET    /api/v1/admin/bookings    # All bookings with filters
DELETE /api/v1/admin/bookings/:id # Force cancel
GET    /api/v1/admin/audit-logs
GET    /api/v1/admin/settings
PATCH  /api/v1/admin/settings
```

**Health**:
```
GET    /api/health               # Health check endpoint
```

**Metrics** (OpenTelemetry):
```
GET    /metrics                  # Prometheus metrics
```

---

### 5.3 Database Schema Decisions

**Recurring Bookings** (Phase 2): Expand on creation (create individual booking rows)

**Cancelled Bookings**: Soft delete (`status = 'cancelled'`, `cancelled_at` timestamp)

**Audit Trail**: Separate audit table
```sql
CREATE TABLE audit_logs (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  entity_type VARCHAR(50) NOT NULL, -- 'booking', 'desk', 'user'
  entity_id INTEGER NOT NULL,
  action VARCHAR(50) NOT NULL, -- 'created', 'updated', 'deleted', 'cancelled'
  changes JSONB, -- Before/after values
  metadata JSONB, -- Additional context (e.g., cancel reason)
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
```

**Idempotency Keys**: Postgres table
```sql
CREATE TABLE idempotency_keys (
  key VARCHAR(255) PRIMARY KEY,
  response JSONB NOT NULL, -- Cached response
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_idempotency_keys_created_at ON idempotency_keys(created_at); -- For cleanup
```

**Cleanup**: Daily job to delete keys older than 24 hours

---

### 5.4 Background Jobs (Worker Service)

**Implementation**: Separate Go worker service using Asynq (Redis-based queue)

**Jobs**:

1. **No-Show Processing** (every 1 minute)
   - Find bookings past grace period with no check-in
   - Auto-cancel (idempotent, transactional)
   - See section 2.3 for implementation

2. **Booking Window Materialization** (daily at 00:05)
   - Expand booking window as time progresses
   - Archive old bookings (optional, Phase 2)

3. **Idempotency Key Cleanup** (daily at 00:10)
   - Delete keys older than 24 hours

**Deployment**: Same Render service, separate process (or separate service)

---

### 5.5 Caching Strategy

**Approach**: No caching for MVP (keep it simple, query DB every time)

**Redis Usage**: Limited to:
- Sessions (refresh tokens) - **moved to Postgres**
- Idempotency keys
- Rate limiting

**Phase 2**: Consider caching desk availability (5min TTL)

---

### 5.6 Rate Limiting

**Implementation**: Redis-based (auth endpoints only for MVP)

**Limits**:
- **Auth endpoints** (login, register): 5 requests/min per IP
- **Booking endpoints** (create, modify): 10 requests/min per user
- **Read endpoints** (availability, my bookings): 100 requests/min per user

**Library**: Use existing Go rate limiter (e.g., `go-redis/redis_rate`)

---

### 5.7 Code Organization

**Monorepo Structure**:
```
hotdesk-booking/
├── frontend/
│   ├── app/              # Next.js App Router
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   └── register/
│   │   ├── dashboard/
│   │   ├── bookings/
│   │   ├── admin/
│   │   └── layout.tsx
│   ├── features/         # Feature-based modules
│   │   ├── bookings/
│   │   │   ├── components/
│   │   │   ├── hooks/
│   │   │   └── types/
│   │   ├── desks/
│   │   ├── auth/
│   │   └── admin/
│   ├── components/       # Shared UI components
│   │   ├── ui/           # shadcn components
│   │   └── layout/
│   ├── lib/              # Utilities, API client
│   └── styles/
│
├── backend/
│   ├── cmd/
│   │   ├── api/          # Main API server
│   │   │   └── main.go
│   │   └── worker/       # Background worker
│   │       └── main.go
│   ├── internal/
│   │   ├── features/     # Feature-based modules
│   │   │   ├── bookings/
│   │   │   │   ├── handler.go
│   │   │   │   ├── service.go
│   │   │   │   ├── repository.go
│   │   │   │   └── types.go
│   │   │   ├── desks/
│   │   │   ├── auth/
│   │   │   └── admin/
│   │   ├── shared/
│   │   │   ├── middleware/  # Auth, rate limiting, logging
│   │   │   ├── database/    # DB connection, migrations
│   │   │   └── utils/
│   │   └── config/
│   ├── migrations/       # SQL migrations (goose/atlas)
│   └── go.mod
│
├── docker-compose.yml    # Local dev (Postgres + Redis)
├── .github/
│   └── workflows/
│       └── ci.yml
├── README.md
├── CONTRIBUTING.md
├── ARCHITECTURE.md
└── DEPLOYMENT.md
```

---

## Deployment & Operations

### 6.1 Platform & Environments

**Platform**: Render

**Environments**:
- **Staging**: Auto-deploy on merge to `main`, manual smoke test
- **Production**: Manual promotion from staging (after approval)

**Regions**: Single region (start with US, multi-region in Phase 2)

**Scaling**: Fixed capacity initially, add auto-scaling in Phase 2

---

### 6.2 Database

**Provider**: Render Postgres (managed)

**Backup**: Automated daily backups + weekly retention for 30 days

**Migrations**: Run via CI/CD job before deployment
- GitHub Actions job: `db-migrate` → runs `goose up` against staging/prod
- Must pass before deploying app

**Connection Pooling**: App-level (pgxpool, max 20 connections)

---

### 6.3 Redis

**Provider**: Render Redis (managed)

**Persistence**: Ephemeral (no RDB/AOF, acceptable for sessions/idempotency keys)

**Fallback**: Critical data (refresh tokens) stored in Postgres

---

### 6.4 Monitoring & Observability

**Logging**:
- Structured JSON logs (via zap) to stdout
- Render aggregates logs (searchable in dashboard)
- Log levels: ERROR, WARN, INFO (no DEBUG in production)

**What to Log** (see Q157):
- All HTTP requests (method, path, status, duration)
- Database queries (with timing, no sensitive params)
- Auth attempts (success/failure, no passwords)
- Booking create/modify/cancel (audit trail)
- Errors + stack traces
- Background job execution (start/end/results)

**Metrics**: Prometheus `/metrics` endpoint (skip Grafana for MVP)

**Tracing**: OpenTelemetry (day 1, sends to Sentry or OTLP collector)

**Uptime Monitoring**: Skip for MVP (rely on Render dashboard)

**Error Tracking**: Sentry (free tier)

**Health Check**:
```go
GET /api/health
Response:
{
  "status": "healthy",
  "timestamp": "2025-01-25T10:00:00Z",
  "services": {
    "database": "connected",
    "redis": "connected"
  }
}
```

---

### 6.5 CI/CD Pipeline

**Platform**: GitHub Actions

**Triggers**:
- On pull request: Run tests, lint, typecheck
- On merge to `main`: Deploy to staging (auto)
- On manual trigger: Promote staging → production (manual approval)

**Required Checks** (must pass before merge):
- ✅ All tests pass (unit + integration)
- ✅ Linting passes (ESLint for frontend, golangci-lint for backend)
- ✅ Type checking passes (TypeScript, Go)
- ✅ Build succeeds (frontend build, backend compile)
- ✅ Security scan (Dependabot, npm audit, gosec)
- ✅ Database migration dry-run succeeds

**PR Requirements**:
- All CI checks passing
- No merge conflicts
- Squash commits on merge
- Delete branch after merge

**Workflow Example** (`.github/workflows/ci.yml`):
```yaml
name: CI/CD

on:
  pull_request:
  push:
    branches: [main]

jobs:
  test-backend:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: test
          POSTGRES_PASSWORD: test
      redis:
        image: redis:7-alpine
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: cd backend && go test ./... -cover
      - run: golangci-lint run

  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: cd frontend && npm ci
      - run: npm run lint
      - run: npm run typecheck
      - run: npm run test
      - run: npm run build

  deploy-staging:
    needs: [test-backend, test-frontend]
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - run: # Trigger Render deploy webhook for staging

  deploy-production:
    needs: [test-backend, test-frontend]
    if: github.event_name == 'workflow_dispatch'
    runs-on: ubuntu-latest
    steps:
      - run: # Trigger Render deploy webhook for production
```

---

### 6.6 Secrets Management

**Local Development**: `.env.local` (gitignored)

**Staging/Production**: Render environment variables (platform-native)

**Example `.env.example`** (checked into repo):
```bash
# Database
DATABASE_URL=postgresql://user:pass@host:5432/dbname

# Redis
REDIS_URL=redis://host:6379

# Auth
JWT_SECRET=changeme-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# App
PORT=8080
NODE_ENV=development

# External Services
SENTRY_DSN=https://...
```

---

## Testing Strategy

### 7.1 Backend Testing

**Coverage Target**: 80% (services, middleware, utilities)

**Unit Tests** (cover these layers):
- ✅ Services (business logic)
- ✅ Middleware (auth, validation)
- ✅ Utilities (helpers, formatters)
- ❌ Handlers (covered by integration tests)
- ❌ Repositories (covered by integration tests)

**Integration Tests** (critical flows):
- Booking creation (happy path)
- Booking creation with conflict (collision detection)
- Booking cancellation
- Booking modification
- Auth flow (register, login, refresh token)
- Desk CRUD operations
- Admin force-cancel booking

**Test Environment**: Docker Compose (Postgres + Redis)

**Test Database Strategy**: Transactions (rollback after each test)

**Load Testing**: Skip for MVP (add in Phase 2 with k6)

**API Contract Testing**: OpenAPI spec validation (ensure responses match schema)

---

### 7.2 Frontend Testing

**Component Tests** (Vitest + RTL):
- BookingForm (creation wizard)
- DeskCalendar/AvailabilityView
- MyBookings list/calendar
- Admin desk management (CRUD)
- Auth forms (login, register)

**Coverage Target**: No specific %, just test critical flows

**E2E Tests** (Playwright):
- Full booking flow (login → search → create → confirm)
- Booking cancellation flow
- Booking modification flow
- Admin login → create desk → verify
- Admin force-cancel user booking
- Conflict handling (double-book attempt → verify error + suggestions)

**Test Environments**:
- Local: Docker Compose
- CI: Against staging environment

**Visual Regression**: Skip for MVP

---

### 7.3 Test Data & Fixtures

**Seed Script**: `npm run seed` (for local dev)

**Creates**:
- 1 admin user (admin@example.com / admin123)
- 2-3 regular users (user1@example.com / password123)
- 10-20 desks (E-101 to E-110, W-201 to W-210)
- 5-10 sample bookings (upcoming, past, cancelled)
- 1-2 conflicting bookings (for testing conflict detection)

**Reset Strategy**: Transactions (rollback after each test)

---

### 7.4 CI Testing Pipeline

**When Tests Run**:
- On every pull request
- Before merging to `main` (required check)

**Parallelization**: Yes, split backend/frontend test suites into separate jobs

**Failure Handling**:
- Unit/integration tests: Block merge (must fix)
- E2E tests: Warning only, allow merge with approval (flakiness tolerance)

---

## MVP Scope & Success Criteria

### 8.1 MVP Features (Must-Have)

**Authentication & Users**:
- ✅ Self-service user registration
- ✅ Login with JWT (15min access, 7-day refresh)
- ✅ Admin password reset (manual)
- ✅ Two roles: Member, Admin
- ✅ "Log out everywhere" functionality

**Desk Management**:
- ✅ Admin: Create/edit/delete desks
- ✅ Desk attributes: number, wing
- ✅ Admin: Block desks for maintenance (auto-cancels conflicts)

**Booking Core**:
- ✅ Create booking (30min blocks, 8am-10pm, 30min-12hr)
- ✅ View availability (desk-first → calendar)
- ✅ Conflict detection (Postgres exclusion + app validation)
- ✅ Booking window: Current + next week (rolling)
- ✅ Policy: Max 10hrs/day
- ✅ Same-day booking allowed
- ✅ Multi-step wizard (time → desk → confirm)

**Booking Management**:
- ✅ View my bookings (calendar view)
- ✅ Modify booking (full modification)
- ✅ Cancel booking (anytime)
- ✅ End booking early (release desk)
- ✅ Check-in flow (15min grace period)
- ✅ Auto-cancel if no check-in
- ✅ "Book again" feature

**Conflict Handling**:
- ✅ Reject + suggest alternatives (same time priority, then different times)

**Admin Features (Priority 1)**:
- ✅ Desk management (CRUD + maintenance)
- ✅ System settings UI (hours, limits)
- ✅ User management (disable, reset passwords)

**Admin Features (Priority 2)**:
- ✅ View all bookings (table + filters)
- ✅ Force-cancel bookings (with reason)
- ✅ Audit log (read-only view)

**Notifications**:
- ✅ In-app only (confirmed, cancelled, modified, auto-cancelled)

**Technical**:
- ✅ OpenTelemetry tracing
- ✅ Sentry error tracking
- ✅ Structured JSON logging
- ✅ OpenAPI spec
- ✅ 80% backend coverage
- ✅ Full CI/CD (staging + production)
- ✅ DB migrations via CI/CD
- ✅ Dark mode (auto-detect)
- ✅ Mobile-responsive (desktop-first)
- ✅ Best-effort accessibility

---

### 8.2 Success Criteria

**Launch Readiness**:
- ✅ All MVP features implemented and tested
- ✅ All CI/CD checks passing
- ✅ Deployed to staging + production
- ✅ At least 5 test users successfully create bookings
- ✅ No critical bugs in first week
- ✅ Admin can fully manage desks and bookings
- ✅ Response time <500ms for booking creation
- ✅ Zero double-bookings (100% conflict detection)

**Timeline**: No fixed deadline (iterate until ready)

---

## Technical Implementation Details

### 9.1 Conflict Detection - Detailed Implementation

**Performance Requirements**:
- 50 concurrent users
- 200 bookings/day
- 20 peak concurrent booking attempts
- <500ms booking creation latency

**Implementation**:

```go
// BookingService.CreateBooking
func (s *BookingService) CreateBooking(ctx context.Context, req CreateBookingRequest) (*Booking, error) {
    // 1. Validate booking window
    if err := validateBookingWindow(req.StartTime, time.Now()); err != nil {
        return nil, err
    }

    // 2. Validate duration constraints (30min - 12hrs)
    duration := req.EndTime.Sub(req.StartTime)
    if duration < 30*time.Minute || duration > 12*time.Hour {
        return nil, errors.New("invalid duration")
    }

    // 3. Validate daily limit (10hrs)
    totalHours, err := s.repo.GetUserDailyHours(ctx, req.UserID, req.StartTime)
    if err != nil {
        return nil, err
    }
    if totalHours + duration.Hours() > 10 {
        return nil, errors.New("exceeds daily 10-hour limit")
    }

    // 4. Check idempotency key
    if req.IdempotencyKey != "" {
        cached, err := s.idempotency.Get(ctx, req.IdempotencyKey)
        if err == nil {
            return cached, nil // Already processed
        }
    }

    // 5. App-level availability check (fast fail)
    available, err := s.repo.IsDeskAvailable(ctx, req.DeskID, req.StartTime, req.EndTime)
    if err != nil {
        return nil, err
    }
    if !available {
        // Generate suggestions
        suggestions := s.generateSuggestions(ctx, req)
        return nil, &ConflictError{Suggestions: suggestions}
    }

    // 6. Insert booking (DB exclusion constraint catches races)
    booking, err := s.repo.CreateBooking(ctx, req)
    if err != nil {
        if isConstraintViolation(err) {
            // Race condition caught by DB
            suggestions := s.generateSuggestions(ctx, req)
            return nil, &ConflictError{Suggestions: suggestions}
        }
        return nil, err
    }

    // 7. Store idempotency key
    if req.IdempotencyKey != "" {
        s.idempotency.Set(ctx, req.IdempotencyKey, booking, 24*time.Hour)
    }

    // 8. Create audit log
    s.audit.Log(ctx, req.UserID, "booking", booking.ID, "created", nil)

    return booking, nil
}

func (s *BookingService) generateSuggestions(ctx context.Context, req CreateBookingRequest) Suggestions {
    // Priority 1: Alternative desks for same time
    altDesks := s.repo.FindAvailableDesks(ctx, req.StartTime, req.EndTime, 5)

    // Priority 2: Alternative times for same desk
    altTimes := s.repo.FindAvailableTimes(ctx, req.DeskID, req.StartTime.Truncate(24*time.Hour), 3)

    return Suggestions{
        AlternativeDesks: altDesks,
        AlternativeTimes: altTimes,
    }
}
```

---

### 9.2 Booking Window Calculation

```go
func getCurrentBookableRange(now time.Time) (start, end time.Time) {
    // Start: can't book in the past
    start = now

    // End: Sunday 23:59:59 of next week
    currentWeekStart := startOfWeek(now) // Monday 00:00
    nextWeekEnd := endOfWeek(currentWeekStart.Add(7 * 24 * time.Hour))
    end = nextWeekEnd

    return start, end
}

func startOfWeek(t time.Time) time.Time {
    // Monday = 1, Sunday = 0
    offset := (int(t.Weekday()) + 6) % 7
    return time.Date(t.Year(), t.Month(), t.Day()-offset, 0, 0, 0, 0, t.Location())
}

func endOfWeek(t time.Time) time.Time {
    return startOfWeek(t).Add(7*24*time.Hour - time.Second)
}
```

**Phase 2 Recurring Bookings**: Create only instances within booking window
- Example: User wants "Every Tue for 6 weeks" on Thu Week 3
- System creates: Tue Week 4 only (Week 5+ outside window)

---

### 9.3 Check-In Flow

**Check-In Window**: 15 minutes before → 15 minutes after start time

**Frontend Logic**:
```typescript
function getCheckInState(booking: Booking): CheckInState {
  const now = new Date();
  const startTime = new Date(booking.start_time);
  const checkInStart = new Date(startTime.getTime() - 15 * 60 * 1000);
  const checkInEnd = new Date(startTime.getTime() + 15 * 60 * 1000);

  if (booking.checked_in_at) {
    return { status: 'checked_in', message: 'Checked in' };
  }

  if (now < checkInStart) {
    return {
      status: 'too_early',
      message: `Check-in available at ${format(checkInStart, 'h:mm a')}`
    };
  }

  if (now >= checkInStart && now <= checkInEnd) {
    return { status: 'available', message: 'Check in now' };
  }

  if (now > checkInEnd) {
    return { status: 'expired', message: 'Booking cancelled (no check-in)' };
  }
}
```

**Backend Check-In Handler**:
```go
func (h *BookingHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
    bookingID := getIDParam(r)
    userID := getUserIDFromContext(r)

    booking, err := h.service.GetBooking(r.Context(), bookingID)
    if err != nil {
        // handle error
    }

    // Verify ownership
    if booking.UserID != userID {
        // unauthorized
    }

    // Verify check-in window
    now := time.Now()
    checkInStart := booking.StartTime.Add(-15 * time.Minute)
    checkInEnd := booking.StartTime.Add(15 * time.Minute)

    if now.Before(checkInStart) {
        // too early
    }

    if now.After(checkInEnd) {
        // too late (should be auto-cancelled already)
    }

    // Update booking
    err = h.service.CheckInBooking(r.Context(), bookingID)
    // ...
}
```

---

### 9.4 Time Zone Handling

**Approach**: Single timezone (all users assumed in same timezone)

**Storage**: Store all times in UTC (database TIMESTAMPTZ)

**Display**: Convert to local timezone in frontend (date-fns)

**Opening Hours**: Hardcoded in UTC equivalent
- If office is in Singapore (UTC+8):
  - 8am SGT = 00:00 UTC
  - 10pm SGT = 14:00 UTC
- Store as `opening_start = '00:00:00'::time`, `opening_end = '14:00:00'::time`

**Phase 2**: Multi-timezone support (store office timezone, convert per user)

---

### 9.5 Idempotency Implementation

**Flow**:
1. Frontend generates UUID on booking creation
2. Include in `Idempotency-Key` header
3. Backend checks Postgres `idempotency_keys` table
4. If exists → return cached response
5. If not → process request, store response (24hr TTL)

**Schema**:
```sql
CREATE TABLE idempotency_keys (
  key VARCHAR(255) PRIMARY KEY,
  response JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Cleanup**: Daily job deletes keys older than 24 hours

---

### 9.6 Data Consistency - Early Booking End

**Approach**: Keep original `time_range`, add `actual_end_time` field

**Schema**:
```sql
-- bookings table
actual_end_time TIMESTAMPTZ, -- NULL if booking runs to scheduled end
```

**Logic**:
```go
func (s *BookingService) EndBookingEarly(ctx context.Context, bookingID int) error {
    now := time.Now()

    // Update booking
    err := s.repo.UpdateBooking(ctx, bookingID, map[string]interface{}{
        "actual_end_time": now,
        "status": "completed",
    })

    // Desk immediately available (queries check actual_end_time OR end_time)
    return err
}
```

**Availability Query** (updated):
```sql
SELECT * FROM desks
WHERE id NOT IN (
  SELECT desk_id FROM bookings
  WHERE status IN ('confirmed', 'completed')
    AND (
      (actual_end_time IS NOT NULL AND time_range && tstzrange($1, $2))
      OR (actual_end_time IS NULL AND time_range && tstzrange($1, $2))
    )
)
```

**Simplified**: Just check if `time_range` overlaps, exclude `completed` with `actual_end_time < query_start`

---

### 9.7 Audit Trail - Booking Modification

**Approach**: Create audit log entry with before/after values

**Example**:
```json
{
  "user_id": 42,
  "entity_type": "booking",
  "entity_id": 123,
  "action": "updated",
  "changes": {
    "desk_id": { "from": 101, "to": 105 },
    "start_time": { "from": "2025-01-16T14:00:00Z", "to": "2025-01-16T16:00:00Z" }
  },
  "metadata": null,
  "created_at": "2025-01-25T10:30:00Z"
}
```

**Force Cancel** (admin):
```json
{
  "user_id": 1, // admin
  "entity_type": "booking",
  "entity_id": 123,
  "action": "cancelled",
  "changes": null,
  "metadata": {
    "reason": "Desk maintenance required",
    "cancelled_user_id": 42
  },
  "created_at": "2025-01-25T10:30:00Z"
}
```

---

## Development Workflow

### 10.1 Local Development Setup

**Prerequisites**:
- Go 1.22+
- Node.js 20+
- Docker + Docker Compose

**Setup Steps**:
```bash
# 1. Clone repo
git clone <repo-url>
cd hotdesk-booking

# 2. Start Docker services
docker-compose up -d

# 3. Backend setup
cd backend
cp .env.example .env.local
# Edit .env.local with local values
go mod download
goose -dir migrations up  # Run migrations
go run cmd/api/main.go    # Start API server (localhost:8080)

# 4. Frontend setup (new terminal)
cd frontend
cp .env.example .env.local
npm install
npm run dev               # Start Next.js (localhost:3000)

# 5. Seed data (optional)
npm run seed              # Creates test users, desks, bookings
```

**Docker Compose** (`docker-compose.yml`):
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: hotdesk_dev
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

**Environment Variables** (`.env.local` example):
```bash
# Backend
DATABASE_URL=postgresql://dev:dev@localhost:5432/hotdesk_dev
REDIS_URL=redis://localhost:6379
JWT_SECRET=local-dev-secret-change-in-prod
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
PORT=8080
LOG_LEVEL=debug

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

---

### 10.2 Code Quality Tools

**Installed Tools**:
- ✅ Prettier (auto-format)
- ✅ ESLint (frontend linting)
- ✅ golangci-lint (backend linting)
- ✅ Husky + lint-staged (pre-commit hooks)
- ✅ EditorConfig (consistent editor settings)

**Pre-Commit Hook** (`.husky/pre-commit`):
```bash
#!/bin/sh
. "$(dirname "$0")/_/husky.sh"

# Run lint-staged (formats and lints staged files)
npx lint-staged

# Run backend linter
cd backend && golangci-lint run --fix
```

**TypeScript Config** (`tsconfig.json`):
```json
{
  "compilerOptions": {
    "strict": true,
    "target": "ES2022",
    "lib": ["ES2022", "DOM"],
    "module": "ESNext",
    "moduleResolution": "bundler",
    "jsx": "preserve",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

**Commit Convention**: Conventional Commits (enforced by commitlint)
```bash
# Valid examples:
feat: add booking creation wizard
fix: resolve conflict detection race condition
chore: update dependencies
docs: add API documentation

# Invalid (will reject):
Added booking feature  # Missing type prefix
```

**Branch Naming**:
- `feature/booking-wizard`
- `bugfix/conflict-detection-race`
- `hotfix/auth-token-expiry`

---

### 10.3 Documentation

**Required Docs** (in repo):
- ✅ `README.md` - Project overview, setup instructions
- ✅ `CONTRIBUTING.md` - How to contribute, dev setup
- ✅ `API.md` - API endpoint documentation (or link to OpenAPI)
- ✅ `ARCHITECTURE.md` - System design, key decisions
- ✅ `DEPLOYMENT.md` - Deployment guide
- ✅ Inline code comments (sparingly, for non-obvious logic)
- ✅ Database schema diagram (simple ERD)

**API Documentation**: OpenAPI spec + examples in markdown

**Database Migrations**: Minimal comments (purpose only)
```sql
-- Migration: 001_create_bookings_table
-- Purpose: Create bookings table with exclusion constraint

CREATE TABLE bookings (...);
```

---

### 10.4 Feature Development Workflow

**Process**: Issue → Branch → PR → Review → Merge

1. **Create GitHub Issue** (describe feature/bug)
2. **Create Branch** (`feature/ISSUE-123-booking-wizard`)
3. **Develop + Test Locally**
4. **Open PR** (link to issue, describe changes)
5. **CI Runs** (all checks must pass)
6. **Self-Review** + AI-Assisted Review (ChatGPT)
7. **Merge to main** (squash commits, delete branch)
8. **Auto-Deploy to Staging**
9. **Smoke Test Staging**
10. **Manual Promotion to Production**

**PR Template** (`.github/pull_request_template.md`):
```markdown
## Description
Fixes #ISSUE_NUMBER

Brief description of changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Checklist
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No new warnings in console
- [ ] Self-reviewed code
- [ ] AI-reviewed code (ChatGPT)
```

---

### 10.5 Debugging & Monitoring

**Health Check Endpoint**:
```go
GET /api/health
{
  "status": "healthy",
  "timestamp": "2025-01-25T10:00:00Z",
  "services": {
    "database": "connected",
    "redis": "connected"
  },
  "version": "1.0.0"
}
```

**Admin Debug Endpoints** (protected by feature flag, disabled in production):
```go
// Only enabled if DEBUG_MODE=true
GET /api/admin/debug/booking/:id
GET /api/admin/debug/desk/:id/schedule
POST /api/admin/debug/simulate-no-show  // Trigger worker job manually
```

**Logging Levels**:
- **Production**: ERROR, WARN, INFO
- **Staging**: ERROR, WARN, INFO, DEBUG
- **Local**: ALL

**Example Log Entry** (structured JSON):
```json
{
  "level": "info",
  "timestamp": "2025-01-25T10:30:00Z",
  "request_id": "uuid",
  "method": "POST",
  "path": "/api/v1/bookings",
  "status": 201,
  "duration_ms": 45,
  "user_id": 42
}
```

---

### 10.6 Security Best Practices

**Rate Limiting**:
- Auth endpoints: 5/min per IP
- Booking endpoints: 10/min per user
- Read endpoints: 100/min per user

**Password Hashing**: bcrypt (cost factor 10-12)

**CORS**:
- Development: `http://localhost:3000`
- Staging: `https://hotdesk-staging.onrender.com`
- Production: `https://hotdesk.onrender.com`

**SQL Injection Prevention**: Use pgx parameterized queries (never string concatenation)

**CSP Headers**: Permissive for MVP, tighten in Phase 2

**Admin Impersonation**: Not allowed (security risk)

---

### 10.7 Performance Optimization

**Database Indexes** (critical for MVP):
```sql
-- Bookings
CREATE INDEX idx_bookings_desk_id ON bookings(desk_id);
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_time_range ON bookings USING GIST(time_range);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_start_time ON bookings(start_time);

-- Users
CREATE INDEX idx_users_email ON users(email);

-- Sessions
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
```

**Caching**: None for MVP

**Bundle Size**: Monitor, optimize if >200KB initial load

---

### 10.8 UX Details

**Loading States**: Combination approach
- Skeleton loaders for lists (desk list, booking list)
- Spinners for actions (creating booking, cancelling)
- Optimistic updates where safe (check-in button)

**Error Messages**: Detailed + actionable (option C)
```
❌ Booking failed: Desk E-101 is unavailable 2:00-4:00pm

Suggestions:
• Try Desk E-105 (East Wing) for same time
• Try Desk W-203 (West Wing) for same time
• Book Desk E-101 at 4:00-6:00pm instead
```

**Empty States**: Helpful message + CTA
```
📭 No bookings yet

You haven't made any bookings. Ready to book your first desk?

[Create Booking]
```

**Onboarding**: Just a "Help" page (no interactive tutorial)

---

### 10.9 Miscellaneous Decisions

**Time Format**: 24-hour (14:00)

**Date Format**: Auto-detect from locale (default DD/MM in Singapore)

**Terms/Privacy**: Placeholder for MVP (required before production launch)

**Analytics**: Server-side only (no client-side tracking like GA)

**Feedback**: Contact email link (no in-app form)

---

## Phase 2 Features

**Recurring Bookings**:
- Weekly recurring with custom days
- All-or-nothing conflict handling
- Expand instances on creation

**Waitlist**:
- Join waitlist for fully-booked slots
- Auto-notify when desk available

**No-Show Tracking & Penalties**:
- Track no-show counts per user
- Admin report: Most no-shows
- Optional: Auto-restrict after X no-shows

**Advanced Analytics**:
- Desk utilization metrics
- Peak times analysis
- Popular desk reports

**Email Notifications**:
- Email service integration (SendGrid/Mailgun)
- User notification preferences
- Configurable reminder timing

**Mobile Optimization**:
- Mobile-first redesign
- Touch-optimized interactions

**Enhanced Admin**:
- Exportable CSV reports
- Bulk operations (import desks, bulk cancel)

---

## Out of Scope

**Explicitly Excluded** (not in roadmap):
- Floor plan visualization
- Check-in kiosks/QR codes
- Payment integration
- Multi-location support
- Resource-based permissions
- SSO/OAuth
- Slack/Teams integration
- Public API for third-parties

---

## Appendix: Database Schema (Complete)

```sql
-- Extensions
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Enums
CREATE TYPE wing_type AS ENUM ('East', 'West');
CREATE TYPE desk_status AS ENUM ('available', 'maintenance');
CREATE TYPE booking_status AS ENUM ('confirmed', 'cancelled', 'completed', 'no_show');
CREATE TYPE user_role AS ENUM ('member', 'admin');
CREATE TYPE user_status AS ENUM ('active', 'disabled');

-- Users
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role user_role NOT NULL DEFAULT 'member',
  status user_status NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Sessions
CREATE TABLE sessions (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_token VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Desks
CREATE TABLE desks (
  id SERIAL PRIMARY KEY,
  desk_number VARCHAR(20) NOT NULL UNIQUE,
  wing wing_type NOT NULL,
  status desk_status NOT NULL DEFAULT 'available',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_desks_wing ON desks(wing);
CREATE INDEX idx_desks_status ON desks(status);

-- Bookings
CREATE TABLE bookings (
  id SERIAL PRIMARY KEY,
  desk_id INTEGER NOT NULL REFERENCES desks(id),
  user_id INTEGER NOT NULL REFERENCES users(id),
  start_time TIMESTAMPTZ NOT NULL,
  end_time TIMESTAMPTZ NOT NULL,
  time_range TSTZRANGE GENERATED ALWAYS AS (tstzrange(start_time, end_time)) STORED,
  status booking_status NOT NULL DEFAULT 'confirmed',
  checked_in_at TIMESTAMPTZ,
  actual_end_time TIMESTAMPTZ,
  cancelled_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT no_overlapping_bookings
    EXCLUDE USING GIST (
      desk_id WITH =,
      time_range WITH &&
    ) WHERE (status NOT IN ('cancelled', 'no_show'))
);

CREATE INDEX idx_bookings_desk_id ON bookings(desk_id);
CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_time_range ON bookings USING GIST(time_range);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_start_time ON bookings(start_time);

-- Audit Logs
CREATE TABLE audit_logs (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  entity_type VARCHAR(50) NOT NULL,
  entity_id INTEGER NOT NULL,
  action VARCHAR(50) NOT NULL,
  changes JSONB,
  metadata JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- Idempotency Keys
CREATE TABLE idempotency_keys (
  key VARCHAR(255) PRIMARY KEY,
  response JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_idempotency_keys_created_at ON idempotency_keys(created_at);

-- Notifications
CREATE TABLE notifications (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type VARCHAR(50) NOT NULL,
  message TEXT NOT NULL,
  read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_read ON notifications(read);

-- Settings (single row table for global settings)
CREATE TABLE settings (
  id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  opening_start TIME NOT NULL DEFAULT '08:00:00',
  opening_end TIME NOT NULL DEFAULT '22:00:00',
  daily_hour_limit INTEGER NOT NULL DEFAULT 10,
  check_in_grace_period_minutes INTEGER NOT NULL DEFAULT 15,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO settings (id) VALUES (1);
```

---

## Summary

**🎯 All decisions finalized. Ready for implementation.**

**Key Highlights**:
- **Tech Stack**: Go backend, Next.js frontend, Postgres, Redis, Render deployment
- **Core Feature**: Conflict-free desk booking with 1-week lookahead window
- **Performance**: <500ms booking creation, zero double-bookings
- **Testing**: 80% backend coverage, E2E tests for critical flows
- **CI/CD**: Staging + production with full automation
- **Timeline**: No fixed deadline, iterate until success criteria met

**Next Steps**:
1. Set up repo structure (/frontend, /backend)
2. Initialize database schema + migrations
3. Implement auth layer (JWT, bcrypt)
4. Build booking engine (conflict detection first)
5. Frontend booking wizard
6. Admin dashboard
7. Testing + CI/CD
8. Deploy to staging
9. User testing
10. Production launch

---

**Document Version**: 1.0 (All Decisions Final)
**Last Updated**: 2025-01-25
**Status**: ✅ APPROVED - READY FOR DEVELOPMENT
