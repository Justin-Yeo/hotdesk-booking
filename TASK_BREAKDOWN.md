# Hotdesk Booking System - Detailed Task Breakdown

**Version**: 1.0
**Date**: 2025-01-25
**Implementation Order**: Sequential (solo development)

---

## Table of Contents

1. [Phase 1: Project Setup & Infrastructure](#phase-1-project-setup--infrastructure)
2. [Phase 2: Database Layer](#phase-2-database-layer)
3. [Phase 3: Backend - Authentication & Authorization](#phase-3-backend---authentication--authorization)
4. [Phase 4: Backend - Core Booking Engine](#phase-4-backend---core-booking-engine)
5. [Phase 5: Backend - Desk Management](#phase-5-backend---desk-management)
6. [Phase 6: Backend - Worker Service](#phase-6-backend---worker-service)
7. [Phase 7: Frontend - Project Setup](#phase-7-frontend---project-setup)
8. [Phase 8: Frontend - Authentication](#phase-8-frontend---authentication)
9. [Phase 9: Frontend - Desk Browsing & Availability](#phase-9-frontend---desk-browsing--availability)
10. [Phase 10: Frontend - Booking Management](#phase-10-frontend---booking-management)
11. [Phase 11: Frontend - Admin Dashboard](#phase-11-frontend---admin-dashboard)
12. [Phase 12: Testing](#phase-12-testing)
13. [Phase 13: CI/CD & Deployment](#phase-13-cicd--deployment)
14. [Phase 14: Documentation](#phase-14-documentation)
15. [Phase 15: Launch Preparation](#phase-15-launch-preparation)

---

## Phase 1: Project Setup & Infrastructure

### Task 1.1: Initialize Monorepo Structure (DONE)
**Description**: Create the monorepo directory structure with frontend and backend folders.

**Steps**:
1. Create root directory `hotdesk-booking/`
2. Create `frontend/` and `backend/` subdirectories
3. Create `.gitignore` for both frontend and backend
4. Initialize Git repository
5. Create basic README.md with project overview

**Acceptance Criteria**:
- ✅ Directory structure matches plan (see IMPLEMENTATION_PLAN.md section 5.7)
- ✅ `.gitignore` excludes `node_modules`, `.env.local`, `dist/`, `build/`
- ✅ Git repository initialized with initial commit
- ✅ README.md contains project name, description, and basic setup instructions

**Technical Criteria**:
- Directory structure validated
- Git history shows initial commit

**User-Facing Criteria**:
- Developer can clone and see organized project structure

---

### Task 1.2: Setup Docker Compose for Local Development (DONE)
**Description**: Configure Docker Compose to run PostgreSQL 16 and Redis 7 for local development.

**Steps**:
1. Create `docker-compose.yml` in root directory
2. Configure PostgreSQL service (port 5432)
3. Configure Redis service (port 6379)
4. Add volume mounts for data persistence
5. Create `.env.example` with connection strings
6. Test services startup

**Acceptance Criteria**:
- ✅ `docker-compose up -d` starts PostgreSQL and Redis successfully
- ✅ PostgreSQL accessible on localhost:5432 with credentials from env
- ✅ Redis accessible on localhost:6379
- ✅ Data persists across container restarts
- ✅ `.env.example` includes all required connection strings

**Technical Criteria**:
- PostgreSQL version 16-alpine running
- Redis version 7-alpine running
- Services respond to connection attempts within 5 seconds

**User-Facing Criteria**:
- Developer can start local services with single command
- Services start within 10 seconds

**Performance**:
- Container startup time <10s

---

### Task 1.3: Setup Backend Go Project
**Description**: Initialize Go project with Fiber framework and basic project structure.

**Steps**:
1. Initialize Go module (`go mod init`)
2. Install dependencies: fiber, pgx, zap, golang-jwt
3. Create directory structure (cmd/, internal/, migrations/)
4. Create basic main.go with health check endpoint
5. Create `.env.example` for backend
6. Configure zap logger for structured logging

**Acceptance Criteria**:
- ✅ `go mod tidy` runs without errors
- ✅ All dependencies installed and verified
- ✅ Directory structure matches plan
- ✅ Server starts on port 8080
- ✅ Health check endpoint `/api/health` returns 200 OK
- ✅ Structured logging outputs to stdout in JSON format

**Technical Criteria**:
- Go version 1.22+ verified
- Health check response format matches spec:
  ```json
  {
    "status": "healthy",
    "timestamp": "2025-01-25T10:00:00Z",
    "services": {
      "database": "not_connected",
      "redis": "not_connected"
    }
  }
  ```

**User-Facing Criteria**:
- Developer can run `go run cmd/api/main.go` successfully
- Health check accessible at `http://localhost:8080/api/health`

**Performance**:
- Server startup time <2s
- Health check response time <50ms

---

### Task 1.4: Setup Frontend Next.js Project (DONE)
**Description**: Initialize Next.js 14+ project with App Router, TypeScript, and Mantine UI.

**Steps**:
1. Create Next.js app with TypeScript (`npx create-next-app@latest`)
2. Configure App Router structure
3. Install Mantine packages (@mantine/core, @mantine/hooks, @mantine/form, @mantine/dates, @mantine/notifications)
4. Install Tabler Icons
5. Install TanStack Query
6. Install Zod for validation
7. Configure `tsconfig.json` with strict mode
8. Create `.env.example` with `NEXT_PUBLIC_API_BASE_URL`
9. Setup Mantine theme provider in `app/layout.tsx`

**Acceptance Criteria**:
- ✅ Next.js app created with TypeScript and App Router
- ✅ All dependencies installed without conflicts
- ✅ `npm run dev` starts development server on port 3000
- ✅ Mantine theme provider configured in root layout
- ✅ TypeScript strict mode enabled
- ✅ Default page renders successfully

**Technical Criteria**:
- Next.js version 14+
- TypeScript strict mode enabled in tsconfig.json
- No TypeScript errors on build
- Mantine theme provider wraps entire app

**User-Facing Criteria**:
- Developer can access app at `http://localhost:3000`
- Mantine components render correctly

**Performance**:
- Development server starts <5s
- Initial page load <1s

---

### Task 1.5: Setup Code Quality Tools (DONE)
**Description**: Configure Prettier, ESLint, golangci-lint, Husky, and commitlint.

**Steps**:
1. Install and configure Prettier (frontend)
2. Install and configure ESLint (frontend)
3. Install golangci-lint (backend)
4. Setup Husky for git hooks
5. Configure lint-staged
6. Setup commitlint for conventional commits
7. Create EditorConfig file
8. Test pre-commit hooks

**Acceptance Criteria**:
- ✅ Prettier formats code on save
- ✅ ESLint catches frontend code issues
- ✅ golangci-lint catches backend code issues
- ✅ Pre-commit hook runs formatters and linters
- ✅ Commit messages validated against conventional commits
- ✅ Invalid commit messages rejected

**Technical Criteria**:
- Prettier configured to format .ts, .tsx, .js, .json files
- ESLint extends recommended rules
- golangci-lint configured with standard rules
- Pre-commit hook exits with error if linting fails

**User-Facing Criteria**:
- Code automatically formatted before commit
- Helpful error messages for invalid commits

**Performance**:
- Pre-commit hook execution <5s

---

## Phase 2: Database Layer

### Task 2.1: Install and Configure Database Migration Tool
**Description**: Setup goose for database migrations.

**Steps**:
1. Install goose CLI
2. Create migrations directory
3. Create initial migration configuration
4. Test migration commands (up/down)
5. Document migration workflow in README

**Acceptance Criteria**:
- ✅ goose installed and accessible via CLI
- ✅ Migrations directory created at `backend/migrations/`
- ✅ `goose up` and `goose down` commands work
- ✅ Migration workflow documented

**Technical Criteria**:
- goose version matches recommendation
- Migrations use PostgreSQL dialect
- Connection string read from environment variable

**User-Facing Criteria**:
- Developer can run migrations with simple commands

---

### Task 2.2: Create Users Table Migration
**Description**: Create migration for users table with role and status fields.

**Steps**:
1. Create migration file `001_create_users_table.sql`
2. Define user_role enum (member, admin)
3. Define user_status enum (active, disabled)
4. Create users table schema
5. Add indexes on email and role
6. Add down migration
7. Test migration up and down

**Acceptance Criteria**:
- ✅ Migration creates users table successfully
- ✅ Enum types created: user_role, user_status
- ✅ Indexes created on email and role columns
- ✅ Down migration drops table and enums cleanly
- ✅ All constraints defined (NOT NULL, UNIQUE, DEFAULT)

**Technical Criteria**:
- Table schema matches specification in IMPLEMENTATION_PLAN.md:
  ```sql
  CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'member',
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
  ```
- Indexes: idx_users_email, idx_users_role

**User-Facing Criteria**:
- Users can be stored in database

**Performance**:
- Email lookup via index <10ms

---

### Task 2.3: Create Sessions Table Migration
**Description**: Create migration for sessions table to store refresh tokens.

**Steps**:
1. Create migration file `002_create_sessions_table.sql`
2. Create sessions table schema
3. Add foreign key to users table with CASCADE delete
4. Add indexes on user_id, refresh_token, expires_at
5. Add down migration
6. Test migration up and down

**Acceptance Criteria**:
- ✅ Migration creates sessions table successfully
- ✅ Foreign key to users with ON DELETE CASCADE
- ✅ Indexes created on user_id, refresh_token, expires_at
- ✅ Down migration drops table cleanly
- ✅ Deleting user cascades to sessions

**Technical Criteria**:
- Table schema matches specification
- UNIQUE constraint on refresh_token
- expires_at indexed for cleanup queries

**User-Facing Criteria**:
- User sessions can persist across API restarts

**Performance**:
- Refresh token lookup via index <10ms

---

### Task 2.4: Create Desks Table Migration
**Description**: Create migration for desks table with wing and status fields.

**Steps**:
1. Create migration file `003_create_desks_table.sql`
2. Define wing_type enum (East, West)
3. Define desk_status enum (available, maintenance)
4. Create desks table schema
5. Add indexes on wing and status
6. Add down migration
7. Test migration up and down

**Acceptance Criteria**:
- ✅ Migration creates desks table successfully
- ✅ Enum types created: wing_type, desk_status
- ✅ UNIQUE constraint on desk_number
- ✅ Indexes created on wing and status columns
- ✅ Down migration drops table and enums cleanly

**Technical Criteria**:
- Table schema matches specification
- Default status is 'available'

**User-Facing Criteria**:
- Desks can be created and managed

**Performance**:
- Desk lookup by number <10ms

---

### Task 2.5: Create Bookings Table Migration with Exclusion Constraint
**Description**: Create migration for bookings table with conflict detection exclusion constraint.

**Steps**:
1. Create migration file `004_create_bookings_table.sql`
2. Enable btree_gist extension
3. Define booking_status enum
4. Create bookings table with time_range generated column
5. Add exclusion constraint for overlap detection
6. Add indexes on desk_id, user_id, time_range, status, start_time
7. Add down migration
8. Test constraint by attempting overlapping inserts

**Acceptance Criteria**:
- ✅ Migration creates bookings table successfully
- ✅ btree_gist extension enabled
- ✅ Exclusion constraint prevents overlapping bookings
- ✅ Generated column time_range created correctly
- ✅ All indexes created
- ✅ Attempting overlapping booking raises constraint violation
- ✅ Cancelled bookings don't trigger constraint

**Technical Criteria**:
- Exclusion constraint definition:
  ```sql
  CONSTRAINT no_overlapping_bookings
    EXCLUDE USING GIST (
      desk_id WITH =,
      time_range WITH &&
    ) WHERE (status NOT IN ('cancelled', 'no_show'))
  ```
- GIST index on time_range for range queries

**User-Facing Criteria**:
- Double-bookings prevented at database level

**Performance**:
- Conflict check via exclusion constraint <50ms

---

### Task 2.6: Create Audit Logs Table Migration
**Description**: Create migration for audit_logs table to track all entity changes.

**Steps**:
1. Create migration file `005_create_audit_logs_table.sql`
2. Create audit_logs table with JSONB columns
3. Add indexes on user_id, entity type/id, created_at
4. Add down migration
5. Test migration up and down

**Acceptance Criteria**:
- ✅ Migration creates audit_logs table successfully
- ✅ JSONB columns for changes and metadata
- ✅ Indexes created for common queries
- ✅ Down migration drops table cleanly

**Technical Criteria**:
- Composite index on (entity_type, entity_id)
- created_at indexed for time-based queries

**User-Facing Criteria**:
- All changes tracked for auditing

---

### Task 2.7: Create Idempotency Keys Table Migration
**Description**: Create migration for idempotency_keys table.

**Steps**:
1. Create migration file `006_create_idempotency_keys_table.sql`
2. Create idempotency_keys table
3. Add index on created_at for cleanup
4. Add down migration
5. Test migration up and down

**Acceptance Criteria**:
- ✅ Migration creates idempotency_keys table successfully
- ✅ Primary key on key column
- ✅ JSONB response column
- ✅ Index on created_at

**Technical Criteria**:
- 24-hour TTL cleanup supported via created_at index

**User-Facing Criteria**:
- Duplicate requests handled gracefully

---

### Task 2.8: Create Notifications Table Migration
**Description**: Create migration for notifications table.

**Steps**:
1. Create migration file `007_create_notifications_table.sql`
2. Create notifications table
3. Add foreign key to users with CASCADE delete
4. Add indexes on user_id and read status
5. Add down migration
6. Test migration up and down

**Acceptance Criteria**:
- ✅ Migration creates notifications table successfully
- ✅ Foreign key to users with CASCADE delete
- ✅ Indexes on user_id and read columns
- ✅ Default value for read is FALSE

**Technical Criteria**:
- Efficient queries for unread notifications

**User-Facing Criteria**:
- In-app notifications supported

---

### Task 2.9: Create Settings Table Migration
**Description**: Create migration for global settings table (single row).

**Steps**:
1. Create migration file `008_create_settings_table.sql`
2. Create settings table with CHECK constraint (id = 1)
3. Insert default settings row
4. Add down migration
5. Test migration up and down

**Acceptance Criteria**:
- ✅ Migration creates settings table successfully
- ✅ CHECK constraint ensures only one row
- ✅ Default values: opening 8:00-22:00, 10hr limit, 15min grace
- ✅ Initial row inserted

**Technical Criteria**:
- Attempting second insert fails due to CHECK constraint

**User-Facing Criteria**:
- Admin can modify global settings

---

### Task 2.10: Create Database Connection Pool
**Description**: Implement PostgreSQL connection pooling using pgxpool.

**Steps**:
1. Create `internal/shared/database/postgres.go`
2. Implement connection pool configuration
3. Parse DATABASE_URL from environment
4. Configure pool size (max 20 connections)
5. Add connection health check
6. Update health check endpoint to test DB connection
7. Write unit tests for connection logic

**Acceptance Criteria**:
- ✅ Connection pool created successfully
- ✅ Max 20 connections configured
- ✅ Health check endpoint shows database status
- ✅ Connection errors handled gracefully
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Connection string parsed from DATABASE_URL environment variable
- Pool configuration: max_conns=20, min_conns=2
- Health check performs `SELECT 1` query

**User-Facing Criteria**:
- API can connect to database
- Health check shows database connectivity

**Performance**:
- Connection pool initialization <1s
- Health check query <50ms

---

## Phase 3: Backend - Authentication & Authorization

### Task 3.1: Implement Password Hashing Utility
**Description**: Create utility functions for bcrypt password hashing and verification.

**Steps**:
1. Create `internal/shared/utils/password.go`
2. Implement HashPassword function (bcrypt cost 12)
3. Implement VerifyPassword function
4. Write unit tests for both functions
5. Test edge cases (empty password, wrong password)

**Acceptance Criteria**:
- ✅ HashPassword generates bcrypt hash with cost 12
- ✅ VerifyPassword correctly validates passwords
- ✅ Different inputs produce different hashes
- ✅ Unit tests achieve 100% coverage
- ✅ Edge cases handled (empty password returns error)

**Technical Criteria**:
- bcrypt cost factor 12
- Hash output format validated
- Timing-safe comparison

**User-Facing Criteria**:
- User passwords stored securely

**Performance**:
- Hashing time ~200-300ms (intentionally slow for security)

---

### Task 3.2: Implement JWT Token Generation and Validation
**Description**: Create JWT utility for access and refresh token generation/validation.

**Steps**:
1. Create `internal/shared/utils/jwt.go`
2. Implement GenerateAccessToken (15min expiry)
3. Implement GenerateRefreshToken (7 day expiry)
4. Implement ValidateToken function
5. Read JWT_SECRET from environment
6. Write unit tests for all functions
7. Test token expiry scenarios

**Acceptance Criteria**:
- ✅ Access tokens expire after 15 minutes
- ✅ Refresh tokens expire after 7 days
- ✅ Tokens include user_id and role claims
- ✅ ValidateToken rejects expired tokens
- ✅ ValidateToken rejects invalid signatures
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- JWT library: golang-jwt/jwt/v5
- Claims structure includes: user_id, role, exp, iat
- HMAC-SHA256 signing

**User-Facing Criteria**:
- Users remain logged in for 7 days with refresh token

**Performance**:
- Token generation <10ms
- Token validation <5ms

---

### Task 3.3: Create User Repository Layer
**Description**: Implement database operations for users table.

**Steps**:
1. Create `internal/features/auth/repository.go`
2. Implement CreateUser function
3. Implement GetUserByEmail function
4. Implement GetUserByID function
5. Implement UpdateUser function
6. Write unit tests using transactions (rollback after each test)

**Acceptance Criteria**:
- ✅ CreateUser inserts user with hashed password
- ✅ GetUserByEmail returns user or nil
- ✅ GetUserByID returns user or nil
- ✅ UpdateUser modifies user fields
- ✅ All database errors handled
- ✅ Unit tests achieve >80% coverage
- ✅ Tests use transaction rollback

**Technical Criteria**:
- All queries use pgx parameterized queries (SQL injection prevention)
- Context passed to all database operations
- Errors wrapped with context

**User-Facing Criteria**:
- User data persisted reliably

**Performance**:
- CreateUser <50ms
- GetUserByEmail <10ms (indexed)

---

### Task 3.4: Create Session Repository Layer
**Description**: Implement database operations for sessions table.

**Steps**:
1. Create session repository in `internal/features/auth/repository.go`
2. Implement CreateSession function
3. Implement GetSessionByRefreshToken function
4. Implement DeleteSession function
5. Implement DeleteAllUserSessions function
6. Write unit tests

**Acceptance Criteria**:
- ✅ CreateSession inserts session with 7-day expiry
- ✅ GetSessionByRefreshToken returns session or nil
- ✅ DeleteSession removes single session
- ✅ DeleteAllUserSessions removes all user sessions
- ✅ Expired sessions handled correctly
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Session expiry calculated as NOW() + 7 days
- Refresh token uniqueness enforced

**User-Facing Criteria**:
- Users can log out from one or all devices

**Performance**:
- Session lookup <10ms (indexed)

---

### Task 3.5: Implement Auth Service Layer
**Description**: Create business logic for authentication operations.

**Steps**:
1. Create `internal/features/auth/service.go`
2. Implement Register function (validate, hash password, create user)
3. Implement Login function (verify credentials, generate tokens)
4. Implement RefreshToken function (validate refresh, issue new tokens)
5. Implement Logout function (delete session)
6. Implement LogoutAll function (delete all user sessions)
7. Write unit tests mocking repository

**Acceptance Criteria**:
- ✅ Register validates email format and password length (min 8 chars)
- ✅ Register rejects duplicate emails
- ✅ Login validates credentials and returns tokens
- ✅ Login fails for wrong password
- ✅ RefreshToken validates token and issues new pair
- ✅ RefreshToken rotates refresh token
- ✅ Logout invalidates session
- ✅ LogoutAll invalidates all user sessions
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Email validation regex
- Password minimum 8 characters
- Refresh token rotation on each refresh

**User-Facing Criteria**:
- Users can register, login, and logout securely

**Performance**:
- Register <300ms (bcrypt hashing)
- Login <300ms (bcrypt verification)
- RefreshToken <50ms

---

### Task 3.6: Create Auth HTTP Handlers
**Description**: Implement HTTP handlers for auth endpoints.

**Steps**:
1. Create `internal/features/auth/handler.go`
2. Implement RegisterHandler (POST /api/v1/auth/register)
3. Implement LoginHandler (POST /api/v1/auth/login)
4. Implement RefreshHandler (POST /api/v1/auth/refresh)
5. Implement LogoutHandler (POST /api/v1/auth/logout)
6. Implement LogoutAllHandler (POST /api/v1/auth/logout-all)
7. Add request/response validation with Zod-like validation
8. Add proper error handling and response formatting

**Acceptance Criteria**:
- ✅ All endpoints return proper JSON responses
- ✅ Request validation rejects invalid input
- ✅ Success responses match API spec format
- ✅ Error responses match API spec format
- ✅ HTTP status codes correct (201, 200, 400, 401, 500)
- ✅ Handlers integrated with service layer

**Technical Criteria**:
- Response envelope format:
  ```json
  {
    "success": true,
    "data": {...},
    "error": null,
    "meta": {
      "timestamp": "2025-01-25T10:00:00Z",
      "request_id": "uuid"
    }
  }
  ```
- Login response includes access_token and refresh_token

**User-Facing Criteria**:
- Users can register and login via API

**Performance**:
- Response time <500ms for all auth endpoints

---

### Task 3.7: Implement Auth Middleware
**Description**: Create middleware to protect routes with JWT authentication.

**Steps**:
1. Create `internal/shared/middleware/auth.go`
2. Implement RequireAuth middleware (validates access token)
3. Implement RequireRole middleware (checks user role)
4. Extract user_id and role from token into context
5. Handle missing/invalid/expired tokens
6. Write unit tests

**Acceptance Criteria**:
- ✅ RequireAuth extracts and validates access token from header
- ✅ Invalid tokens return 401 Unauthorized
- ✅ Missing tokens return 401 Unauthorized
- ✅ Expired tokens return 401 Unauthorized
- ✅ Valid tokens allow request to proceed
- ✅ User ID and role added to request context
- ✅ RequireRole blocks non-admin users from admin routes
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Token read from Authorization header: `Bearer <token>`
- Context keys for user_id and role

**User-Facing Criteria**:
- Protected routes require valid authentication

**Performance**:
- Middleware execution <5ms

---

### Task 3.8: Implement Rate Limiting Middleware
**Description**: Create Redis-based rate limiting for auth endpoints.

**Steps**:
1. Create `internal/shared/middleware/rate_limit.go`
2. Setup Redis client connection
3. Implement rate limiter using Redis (sliding window)
4. Configure limits: auth endpoints 5/min per IP
5. Add rate limit headers to response
6. Write unit tests

**Acceptance Criteria**:
- ✅ Auth endpoints limited to 5 requests/min per IP
- ✅ Exceeding limit returns 429 Too Many Requests
- ✅ Response includes X-RateLimit headers
- ✅ Rate limits reset after 1 minute
- ✅ Different IPs tracked separately
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Redis key format: `ratelimit:auth:{ip_address}`
- Sliding window algorithm
- Headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset

**User-Facing Criteria**:
- Brute force attacks mitigated

**Performance**:
- Rate limit check <10ms

---

## Phase 4: Backend - Core Booking Engine

### Task 4.1: Create Booking Repository Layer
**Description**: Implement database operations for bookings table.

**Steps**:
1. Create `internal/features/bookings/repository.go`
2. Implement CreateBooking function (handles exclusion constraint)
3. Implement GetBookingByID function
4. Implement GetUserBookings function (with filters)
5. Implement UpdateBooking function
6. Implement DeleteBooking function (soft delete)
7. Implement IsDeskAvailable function (app-level check)
8. Implement GetUserDailyHours function (for 10hr limit validation)
9. Write unit tests

**Acceptance Criteria**:
- ✅ CreateBooking inserts booking and handles constraint violations
- ✅ GetBookingByID returns booking or nil
- ✅ GetUserBookings filters by date range and status
- ✅ UpdateBooking modifies booking fields
- ✅ DeleteBooking sets status to 'cancelled' (soft delete)
- ✅ IsDeskAvailable returns true/false correctly
- ✅ GetUserDailyHours calculates total hours for given day
- ✅ Constraint violation returns specific error type
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- IsDeskAvailable query uses GIST index on time_range
- Constraint violation detected by error code "23P01"
- GetUserDailyHours excludes cancelled/no_show bookings

**User-Facing Criteria**:
- Bookings stored reliably with conflict detection

**Performance**:
- CreateBooking <100ms
- IsDeskAvailable <50ms
- GetUserDailyHours <50ms

---

### Task 4.2: Implement Booking Window Calculation Logic
**Description**: Create utility to calculate valid booking window (current + next week).

**Steps**:
1. Create `internal/shared/utils/booking_window.go`
2. Implement GetCurrentBookableRange function
3. Implement startOfWeek helper (Monday 00:00)
4. Implement endOfWeek helper (Sunday 23:59:59)
5. Write unit tests for various dates
6. Test edge cases (Sunday, Monday, midweek)

**Acceptance Criteria**:
- ✅ GetCurrentBookableRange returns correct start (now) and end (next Sunday 23:59:59)
- ✅ Week boundaries calculated correctly (Mon-Sun)
- ✅ Edge cases handled (today is Sunday returns correct range)
- ✅ Unit tests achieve 100% coverage
- ✅ All times in UTC

**Technical Criteria**:
- Week starts Monday 00:00
- Week ends Sunday 23:59:59
- Returns time.Time values in UTC

**User-Facing Criteria**:
- Users can book within current + next week

**Performance**:
- Calculation time <1ms

---

### Task 4.3: Implement Conflict Suggestion Generation
**Description**: Create logic to suggest alternative desks/times when booking conflicts.

**Steps**:
1. Create `internal/features/bookings/suggestions.go`
2. Implement FindAvailableDesks function (same time, different desk)
3. Implement FindAvailableTimes function (same desk, different time)
4. Limit results to top 5 desks and 3 time slots
5. Write unit tests

**Acceptance Criteria**:
- ✅ FindAvailableDesks returns up to 5 available desks for given time
- ✅ FindAvailableTimes returns up to 3 available time slots for given desk
- ✅ Results exclude desks on maintenance
- ✅ Same-wing desks prioritized
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Query optimized with indexes
- Results sorted by wing preference (same wing first)

**User-Facing Criteria**:
- Users get helpful alternatives when booking fails

**Performance**:
- Suggestion generation <100ms

---

### Task 4.4: Implement Idempotency Handler
**Description**: Create idempotency key handling for booking creation.

**Steps**:
1. Create `internal/shared/utils/idempotency.go`
2. Implement CheckIdempotencyKey function (query DB)
3. Implement SaveIdempotencyKey function (store response)
4. Implement idempotency middleware
5. Write unit tests

**Acceptance Criteria**:
- ✅ Idempotency-Key header extracted from request
- ✅ Duplicate requests return cached response
- ✅ New keys stored with 24-hour TTL
- ✅ Missing key allows request to proceed
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Key stored in idempotency_keys table
- Response stored as JSONB
- Cleanup job removes keys older than 24 hours (implemented later)

**User-Facing Criteria**:
- Duplicate booking requests handled gracefully

**Performance**:
- Idempotency check <20ms

---

### Task 4.5: Implement Booking Service Layer
**Description**: Create business logic for booking operations.

**Steps**:
1. Create `internal/features/bookings/service.go`
2. Implement CreateBooking function with all validations
3. Implement GetBooking function
4. Implement GetUserBookings function
5. Implement UpdateBooking function
6. Implement CancelBooking function
7. Implement CheckInBooking function
8. Implement EndBookingEarly function
9. Add audit logging for all operations
10. Write unit tests mocking repository

**Acceptance Criteria**:
- ✅ CreateBooking validates booking window (current + next week)
- ✅ CreateBooking validates duration (30min - 12hrs)
- ✅ CreateBooking validates daily limit (10 hours)
- ✅ CreateBooking checks idempotency key
- ✅ CreateBooking performs availability check (fast fail)
- ✅ CreateBooking handles DB constraint violations
- ✅ CreateBooking generates suggestions on conflict
- ✅ CreateBooking creates audit log entry
- ✅ CheckInBooking validates check-in window (15min before to 15min after)
- ✅ CheckInBooking updates checked_in_at timestamp
- ✅ EndBookingEarly sets actual_end_time and status
- ✅ CancelBooking soft deletes (status = 'cancelled')
- ✅ UpdateBooking validates new time/desk for conflicts
- ✅ All operations create audit log entries
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Validation errors return clear error messages
- Conflict errors include suggestions
- Audit log includes before/after values for updates

**User-Facing Criteria**:
- Users can create, modify, and cancel bookings
- Clear error messages for invalid operations

**Performance**:
- CreateBooking <500ms
- CheckInBooking <100ms
- CancelBooking <100ms

---

### Task 4.6: Create Booking HTTP Handlers
**Description**: Implement HTTP handlers for booking endpoints.

**Steps**:
1. Create `internal/features/bookings/handler.go`
2. Implement CreateBookingHandler (POST /api/v1/bookings)
3. Implement GetBookingsHandler (GET /api/v1/bookings)
4. Implement GetBookingHandler (GET /api/v1/bookings/:id)
5. Implement UpdateBookingHandler (PATCH /api/v1/bookings/:id)
6. Implement CancelBookingHandler (DELETE /api/v1/bookings/:id)
7. Implement CheckInHandler (POST /api/v1/bookings/:id/check-in)
8. Implement EndEarlyHandler (POST /api/v1/bookings/:id/end-early)
9. Add request validation
10. Add authorization checks (user can only modify own bookings)

**Acceptance Criteria**:
- ✅ All endpoints return proper JSON responses
- ✅ CreateBooking validates request body
- ✅ CreateBooking returns 201 with booking object on success
- ✅ CreateBooking returns 409 with suggestions on conflict
- ✅ Authorization prevents users from modifying others' bookings
- ✅ Check-in validates ownership and time window
- ✅ Proper HTTP status codes (201, 200, 400, 401, 403, 404, 409)
- ✅ Response format matches API spec
- ✅ Idempotency-Key header handled for CreateBooking

**Technical Criteria**:
- Request body validation with clear error messages
- User ID extracted from auth context
- Conflict response includes suggestions array

**User-Facing Criteria**:
- Users can manage bookings via API

**Performance**:
- All endpoints respond <500ms

---

### Task 4.7: Implement Audit Log Service
**Description**: Create audit logging service for tracking changes.

**Steps**:
1. Create `internal/shared/audit/service.go`
2. Implement LogChange function (creates audit log entry)
3. Create repository for audit_logs table
4. Support before/after value tracking
5. Write unit tests

**Acceptance Criteria**:
- ✅ LogChange inserts audit log entry
- ✅ Before/after values stored in changes JSONB
- ✅ Metadata field supports additional context
- ✅ All entity types supported (booking, desk, user)
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- JSONB format for changes: `{"field": {"from": "old", "to": "new"}}`
- Audit logs immutable (no updates/deletes)

**User-Facing Criteria**:
- All changes tracked for accountability

**Performance**:
- Audit log insertion <20ms (async preferred)

---

## Phase 5: Backend - Desk Management

### Task 5.1: Create Desk Repository Layer
**Description**: Implement database operations for desks table.

**Steps**:
1. Create `internal/features/desks/repository.go`
2. Implement CreateDesk function
3. Implement GetDeskByID function
4. Implement GetAllDesks function (with filters)
5. Implement UpdateDesk function
6. Implement DeleteDesk function
7. Implement GetDeskAvailability function (time range query)
8. Write unit tests

**Acceptance Criteria**:
- ✅ CreateDesk inserts desk with unique desk_number
- ✅ GetDeskByID returns desk or nil
- ✅ GetAllDesks filters by wing and status
- ✅ UpdateDesk modifies desk fields
- ✅ DeleteDesk removes desk (hard delete only if no bookings)
- ✅ GetDeskAvailability returns time slots with availability status
- ✅ Duplicate desk_number rejected
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- GetDeskAvailability uses GIST index on bookings.time_range
- Filters use indexed columns

**User-Facing Criteria**:
- Admins can manage desks

**Performance**:
- GetAllDesks <50ms
- GetDeskAvailability <100ms

---

### Task 5.2: Implement Desk Service Layer
**Description**: Create business logic for desk operations.

**Steps**:
1. Create `internal/features/desks/service.go`
2. Implement CreateDesk function
3. Implement GetDesk function
4. Implement GetAllDesks function
5. Implement UpdateDesk function
6. Implement DeleteDesk function
7. Implement SetDeskMaintenance function (auto-cancels conflicts)
8. Add audit logging for desk changes
9. Write unit tests

**Acceptance Criteria**:
- ✅ CreateDesk validates desk_number format
- ✅ SetDeskMaintenance changes status to 'maintenance'
- ✅ SetDeskMaintenance auto-cancels conflicting bookings
- ✅ SetDeskMaintenance sends notifications to affected users
- ✅ All operations create audit log entries
- ✅ DeleteDesk prevents deletion if active bookings exist
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Auto-cancel queries all bookings for desk with status 'confirmed'
- Notification created for each affected user

**User-Facing Criteria**:
- Admins can manage desks and maintenance

**Performance**:
- SetDeskMaintenance <500ms (includes cancellations)

---

### Task 5.3: Create Desk HTTP Handlers
**Description**: Implement HTTP handlers for desk endpoints.

**Steps**:
1. Create `internal/features/desks/handler.go`
2. Implement GetDesksHandler (GET /api/v1/desks)
3. Implement GetDeskHandler (GET /api/v1/desks/:id)
4. Implement GetDeskAvailabilityHandler (GET /api/v1/desks/:id/availability)
5. Implement CreateDeskHandler (POST /api/v1/desks) - Admin only
6. Implement UpdateDeskHandler (PATCH /api/v1/desks/:id) - Admin only
7. Implement DeleteDeskHandler (DELETE /api/v1/desks/:id) - Admin only
8. Add request validation and authorization

**Acceptance Criteria**:
- ✅ GetDesks returns list with filters (wing, status)
- ✅ GetDeskAvailability returns calendar data for date range
- ✅ Admin-only endpoints protected with RequireRole middleware
- ✅ Request validation rejects invalid input
- ✅ Proper HTTP status codes
- ✅ Response format matches API spec

**Technical Criteria**:
- GetDeskAvailability accepts start/end query parameters
- Returns 30-minute slot availability

**User-Facing Criteria**:
- Users can browse desks and view availability
- Admins can manage desks

**Performance**:
- GetDesks <100ms
- GetDeskAvailability <200ms

---

## Phase 6: Backend - Worker Service

### Task 6.1: Setup Asynq Worker Service
**Description**: Create separate worker service for background jobs using Asynq.

**Steps**:
1. Create `cmd/worker/main.go`
2. Install Asynq library
3. Configure Asynq server with Redis connection
4. Setup task queue configuration
5. Add graceful shutdown
6. Add structured logging
7. Test worker startup

**Acceptance Criteria**:
- ✅ Worker service starts successfully
- ✅ Connects to Redis for queue
- ✅ Graceful shutdown on SIGTERM/SIGINT
- ✅ Structured logging configured
- ✅ Can register task handlers

**Technical Criteria**:
- Asynq configuration includes concurrency settings
- Redis connection from environment variable

**User-Facing Criteria**:
- Background jobs can be scheduled

**Performance**:
- Worker startup <2s

---

### Task 6.2: Implement No-Show Processing Job
**Description**: Create background job to auto-cancel bookings with no check-in.

**Steps**:
1. Create `internal/features/bookings/worker.go`
2. Implement ProcessNoShows function
3. Query bookings past grace period (15min) with no check-in
4. Cancel each booking in transaction (idempotent)
5. Create audit log for each cancellation
6. Create notification for each user
7. Schedule job to run every 1 minute
8. Write unit tests

**Acceptance Criteria**:
- ✅ Job runs every 1 minute
- ✅ Finds bookings past grace period with no check-in
- ✅ Updates status to 'no_show'
- ✅ Job is idempotent (re-checks status in transaction)
- ✅ Audit log created for each cancellation
- ✅ Notification sent to each affected user
- ✅ Job completes within 30 seconds
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Query: `WHERE status = 'confirmed' AND checked_in_at IS NULL AND start_time <= NOW() - INTERVAL '15 minutes'`
- Transaction with SELECT FOR UPDATE for idempotency

**User-Facing Criteria**:
- No-shows automatically detected and processed

**Performance**:
- Job execution <30s for 200 bookings/day load

---

### Task 6.3: Implement Idempotency Key Cleanup Job
**Description**: Create daily job to delete expired idempotency keys.

**Steps**:
1. Create cleanup function in idempotency service
2. Delete keys older than 24 hours
3. Schedule job to run daily at 00:10
4. Add logging for cleanup metrics
5. Write unit tests

**Acceptance Criteria**:
- ✅ Job runs daily at 00:10
- ✅ Deletes keys where created_at < NOW() - INTERVAL '24 hours'
- ✅ Logs number of keys deleted
- ✅ Job completes within 5 minutes
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Single DELETE query with time filter
- Indexed cleanup using idx_idempotency_keys_created_at

**User-Facing Criteria**:
- Database stays clean

**Performance**:
- Cleanup <1s for typical load

---

### Task 6.4: Implement Session Cleanup Job
**Description**: Create daily job to delete expired sessions.

**Steps**:
1. Create cleanup function in session repository
2. Delete sessions where expires_at < NOW()
3. Schedule job to run daily at 00:15
4. Add logging for cleanup metrics
5. Write unit tests

**Acceptance Criteria**:
- ✅ Job runs daily at 00:15
- ✅ Deletes sessions where expires_at < NOW()
- ✅ Logs number of sessions deleted
- ✅ Job completes within 5 minutes
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Indexed cleanup using idx_sessions_expires_at

**User-Facing Criteria**:
- Database stays clean
- Expired sessions removed

**Performance**:
- Cleanup <1s for typical load

---

## Phase 7: Frontend - Project Setup

### Task 7.1: Configure TanStack Query Provider
**Description**: Setup TanStack Query for server state management.

**Steps**:
1. Create `lib/react-query.tsx` with QueryClient configuration
2. Wrap app in QueryClientProvider in root layout
3. Configure default query options (stale time, retry)
4. Add React Query DevTools for development
5. Test query client initialization

**Acceptance Criteria**:
- ✅ QueryClient created with proper defaults
- ✅ QueryClientProvider wraps entire app
- ✅ DevTools visible in development mode
- ✅ No console errors on app load

**Technical Criteria**:
- Stale time: 5 minutes
- Retry: 1 attempt
- Refetch on window focus disabled

**User-Facing Criteria**:
- Data fetching infrastructure ready

---

### Task 7.2: Create API Client Utility
**Description**: Create typed fetch wrapper for API calls.

**Steps**:
1. Create `lib/api-client.ts`
2. Implement fetchAPI wrapper function
3. Read base URL from NEXT_PUBLIC_API_BASE_URL
4. Add Authorization header injection
5. Add X-Request-Id correlation header
6. Add error handling and response parsing
7. Create TypeScript types for API responses
8. Write unit tests

**Acceptance Criteria**:
- ✅ fetchAPI function accepts URL and options
- ✅ Automatically adds base URL
- ✅ Injects access token from localStorage
- ✅ Generates and adds X-Request-Id header
- ✅ Parses JSON responses
- ✅ Throws typed errors for non-200 responses
- ✅ TypeScript types for success/error responses
- ✅ Unit tests achieve >80% coverage

**Technical Criteria**:
- Base URL from environment variable
- Authorization header: `Bearer <token>`
- Error responses parsed to typed error objects

**User-Facing Criteria**:
- API calls work reliably

**Performance**:
- Wrapper overhead <5ms

---

### Task 7.3: Create Auth Context and Hooks
**Description**: Implement auth context for managing user session.

**Steps**:
1. Create `features/auth/context/AuthContext.tsx`
2. Create useAuth hook
3. Implement login, logout, refreshToken functions
4. Store tokens in localStorage
5. Auto-refresh token before expiry
6. Create ProtectedRoute wrapper component
7. Test auth flow

**Acceptance Criteria**:
- ✅ AuthContext provides user, login, logout, refreshToken
- ✅ Access token stored in localStorage
- ✅ Refresh token stored in localStorage
- ✅ Auto-refresh triggers 1 minute before access token expiry
- ✅ useAuth hook accessible throughout app
- ✅ ProtectedRoute redirects to login if not authenticated
- ✅ User state persists across page refreshes

**Technical Criteria**:
- useEffect hook sets up auto-refresh interval
- Tokens parsed to extract expiry time

**User-Facing Criteria**:
- Users stay logged in across page refreshes
- Session seamlessly refreshed

**Performance**:
- Token refresh <200ms

---

### Task 7.4: Setup Mantine Theme and Dark Mode
**Description**: Configure Mantine theme with dark mode support.

**Steps**:
1. Create `lib/theme.ts` with theme configuration
2. Configure MantineProvider in root layout
3. Setup ColorSchemeScript for dark mode
4. Auto-detect system preference
5. Create theme toggle component (optional)
6. Test light and dark modes

**Acceptance Criteria**:
- ✅ Mantine theme configured with custom colors
- ✅ Dark mode auto-detects system preference
- ✅ All Mantine components render correctly in both modes
- ✅ No FOUC (Flash of Unstyled Content)
- ✅ ColorSchemeScript added to HTML head

**Technical Criteria**:
- Theme uses CSS custom properties
- Dark mode preference saved to localStorage

**User-Facing Criteria**:
- App respects user's dark mode preference

---

### Task 7.5: Create Notification System
**Description**: Setup in-app notification display using Mantine notifications.

**Steps**:
1. Install @mantine/notifications
2. Configure Notifications provider
3. Create useNotifications hook
4. Create notification helper functions (success, error, info)
5. Test notifications

**Acceptance Criteria**:
- ✅ Notifications provider configured
- ✅ useNotifications hook provides show/hide functions
- ✅ Success, error, info notification styles configured
- ✅ Notifications auto-dismiss after 5 seconds
- ✅ Multiple notifications stack correctly

**Technical Criteria**:
- Position: top-right
- Auto-dismiss: 5000ms
- Max notifications: 3

**User-Facing Criteria**:
- Users see clear feedback for actions

---

## Phase 8: Frontend - Authentication

### Task 8.1: Create Login Page
**Description**: Build login page with email/password form.

**Steps**:
1. Create `app/(auth)/login/page.tsx`
2. Create login form using Mantine useForm
3. Add email and password validation (Zod)
4. Implement form submission with API call
5. Handle success (store tokens, redirect)
6. Handle errors (display error message)
7. Add "Remember me" checkbox (visual only, always use standard token lifetime)
8. Add link to register page
9. Test login flow

**Acceptance Criteria**:
- ✅ Form validates email format
- ✅ Form validates password minimum 8 characters
- ✅ Successful login stores tokens and redirects to dashboard
- ✅ Invalid credentials show error message
- ✅ Loading state shown during submission
- ✅ Form accessible via keyboard
- ✅ Link to register page works

**Technical Criteria**:
- POST to /api/v1/auth/login
- Response tokens stored in localStorage
- Redirect to /dashboard on success

**User-Facing Criteria**:
- Users can log in with email and password
- Clear error messages for invalid credentials

**Performance**:
- Login request <500ms

---

### Task 8.2: Create Registration Page
**Description**: Build registration page with email/password form.

**Steps**:
1. Create `app/(auth)/register/page.tsx`
2. Create registration form using Mantine useForm
3. Add email, password, confirm password fields
4. Add validation (Zod): email format, password min 8 chars, passwords match
5. Implement form submission with API call
6. Handle success (show message, redirect to login)
7. Handle errors (duplicate email, etc.)
8. Add link to login page
9. Test registration flow

**Acceptance Criteria**:
- ✅ Form validates email format
- ✅ Form validates password minimum 8 characters
- ✅ Form validates passwords match
- ✅ Successful registration shows success message and redirects to login
- ✅ Duplicate email shows error message
- ✅ Loading state shown during submission
- ✅ Form accessible via keyboard
- ✅ Link to login page works

**Technical Criteria**:
- POST to /api/v1/auth/register
- Redirect to /login on success with success notification

**User-Facing Criteria**:
- Users can self-register
- Clear validation errors

**Performance**:
- Registration request <500ms

---

### Task 8.3: Create Logout Functionality
**Description**: Implement logout and logout all devices functionality.

**Steps**:
1. Add logout function to AuthContext
2. Implement logout API call (POST /api/v1/auth/logout)
3. Clear tokens from localStorage
4. Redirect to login page
5. Add logout button to navigation
6. Add "logout all devices" option
7. Test logout flow

**Acceptance Criteria**:
- ✅ Logout button accessible in navigation
- ✅ Clicking logout clears tokens and redirects to login
- ✅ Logout API call invalidates refresh token
- ✅ "Logout all devices" option available
- ✅ "Logout all" invalidates all user sessions
- ✅ Success notification shown

**Technical Criteria**:
- POST to /api/v1/auth/logout
- POST to /api/v1/auth/logout-all for all devices
- localStorage.clear() called

**User-Facing Criteria**:
- Users can log out from one or all devices

**Performance**:
- Logout completes <200ms

---

## Phase 9: Frontend - Desk Browsing & Availability

### Task 9.1: Create Desk List Component
**Description**: Build component to display all desks with filters.

**Steps**:
1. Create `features/desks/components/DeskList.tsx`
2. Fetch desks using TanStack Query
3. Display desks in grid layout
4. Add wing filter dropdown (East/West/All)
5. Add search input for desk number
6. Add status badge (available/maintenance)
7. Add click handler to view desk details
8. Handle loading and error states
9. Test component

**Acceptance Criteria**:
- ✅ Desks displayed in responsive grid
- ✅ Wing filter works correctly
- ✅ Search filters by desk number
- ✅ Status badge shows correct color
- ✅ Clicking desk opens detail view
- ✅ Loading skeleton shown while fetching
- ✅ Error message shown on API failure
- ✅ Empty state shown when no desks match filters

**Technical Criteria**:
- GET /api/v1/desks with query params
- Client-side filtering for search (or server-side if implemented)
- TanStack Query caching enabled

**User-Facing Criteria**:
- Users can browse and filter desks

**Performance**:
- Desk list renders <500ms
- Filter updates <100ms

---

### Task 9.2: Create Desk Calendar Component
**Description**: Build calendar view showing desk availability for selected desk.

**Steps**:
1. Create `features/desks/components/DeskCalendar.tsx`
2. Fetch desk availability using TanStack Query
3. Display calendar grid (days across top, time slots down left)
4. Show 30-minute time slots from 8:00 AM to 10:00 PM
5. Color-code cells: green (available), gray (occupied)
6. Display current week + next week only
7. Add click handler for available slots (open booking wizard)
8. Handle loading and error states
9. Test component

**Acceptance Criteria**:
- ✅ Calendar displays 2 weeks (current + next)
- ✅ Time slots in 30-minute increments (8:00-22:00)
- ✅ Available slots shown in green
- ✅ Occupied slots shown in gray
- ✅ Clicking available slot opens booking wizard
- ✅ Clicking occupied slot shows "unavailable" message
- ✅ Loading state shown while fetching
- ✅ Calendar responsive on mobile

**Technical Criteria**:
- GET /api/v1/desks/:id/availability?start=...&end=...
- Date range calculated based on booking window logic

**User-Facing Criteria**:
- Users can see desk availability visually

**Performance**:
- Calendar renders <1s
- API call <500ms

---

### Task 9.3: Create Availability Search Page
**Description**: Build main availability page with desk list and calendar.

**Steps**:
1. Create `app/availability/page.tsx`
2. Integrate DeskList component
3. Integrate DeskCalendar component (shown when desk selected)
4. Add state management for selected desk
5. Add page layout and navigation
6. Test page

**Acceptance Criteria**:
- ✅ Page displays desk list by default
- ✅ Selecting desk shows calendar for that desk
- ✅ Back button returns to desk list
- ✅ Page accessible from main navigation
- ✅ Page requires authentication

**Technical Criteria**:
- ProtectedRoute wrapper applied
- State managed with useState or URL params

**User-Facing Criteria**:
- Users can browse desks and view availability

**Performance**:
- Page load <1s

---

## Phase 10: Frontend - Booking Management

### Task 10.1: Create Booking Wizard Component (Step 1: Date/Time)
**Description**: Build first step of booking wizard for date and time selection.

**Steps**:
1. Create `features/bookings/components/BookingWizard.tsx`
2. Create Step 1 form with date picker, start time, duration
3. Restrict date picker to bookable range (current + next week)
4. Populate start time dropdown (30min slots, 8:00-21:30)
5. Populate duration dropdown (30min to 12hrs)
6. Validate selections
7. Add "Next" button
8. Test step 1

**Acceptance Criteria**:
- ✅ Date picker restricted to valid booking window
- ✅ Start time dropdown shows all available slots
- ✅ Duration dropdown shows valid durations
- ✅ Form validates selections before proceeding
- ✅ "Next" button enabled only when valid
- ✅ Validation errors shown clearly

**Technical Criteria**:
- Date picker uses @mantine/dates
- Booking window calculated using same logic as backend

**User-Facing Criteria**:
- Users can select date and time easily

---

### Task 10.2: Create Booking Wizard Component (Step 2: Desk Selection)
**Description**: Build second step of booking wizard for desk selection.

**Steps**:
1. Create Step 2 component
2. Fetch available desks for selected time
3. Display desk list with wing labels
4. Show warning if user approaching 10hr daily limit
5. Add "Back" and "Next" buttons
6. Test step 2

**Acceptance Criteria**:
- ✅ Available desks fetched based on Step 1 time selection
- ✅ Desks displayed with wing information
- ✅ Warning shown if within 1 hour of daily limit
- ✅ Selected desk highlighted
- ✅ "Back" returns to Step 1 with data preserved
- ✅ "Next" enabled only when desk selected
- ✅ Loading state shown while fetching desks

**Technical Criteria**:
- API call to check desk availability for time range
- Daily limit calculated client-side or fetched from API

**User-Facing Criteria**:
- Users can select available desk

**Performance**:
- Desk availability fetch <500ms

---

### Task 10.3: Create Booking Wizard Component (Step 3: Review & Confirm)
**Description**: Build final step of booking wizard for review and confirmation.

**Steps**:
1. Create Step 3 component
2. Display summary: date, time, duration, desk
3. Show policy warnings if applicable
4. Add "Confirm Booking" button
5. Handle booking creation API call
6. Generate idempotency key for request
7. Handle success (redirect to My Bookings)
8. Handle conflict (show error + suggestions)
9. Add "Back" button
10. Test step 3

**Acceptance Criteria**:
- ✅ Summary shows all booking details
- ✅ Policy warnings displayed if approaching limits
- ✅ "Confirm Booking" creates booking via API
- ✅ Idempotency-Key header sent with request
- ✅ Success shows notification and redirects to My Bookings
- ✅ Conflict shows error with alternative desks/times
- ✅ User can click suggestion to restart wizard with new values
- ✅ "Back" returns to Step 2 with data preserved
- ✅ Loading state shown during API call

**Technical Criteria**:
- POST /api/v1/bookings with Idempotency-Key header
- UUID generated for idempotency key
- Conflict response parsed and displayed

**User-Facing Criteria**:
- Users can review and confirm bookings
- Helpful suggestions shown on conflict

**Performance**:
- Booking creation <500ms

---

### Task 10.4: Create My Bookings Page
**Description**: Build page to display user's bookings in calendar view.

**Steps**:
1. Create `app/bookings/page.tsx`
2. Fetch user's bookings using TanStack Query
3. Display bookings in weekly calendar view
4. Color-code by status (upcoming: blue, past: gray, cancelled: red)
5. Add click handler to open booking detail modal
6. Add filter controls (upcoming/past/all)
7. Handle loading and error states
8. Test page

**Acceptance Criteria**:
- ✅ Bookings displayed in calendar view
- ✅ Color coding by status works correctly
- ✅ Clicking booking opens detail modal
- ✅ Filter controls work
- ✅ Loading state shown while fetching
- ✅ Empty state shown when no bookings
- ✅ Page requires authentication

**Technical Criteria**:
- GET /api/v1/bookings
- TanStack Query caching and invalidation

**User-Facing Criteria**:
- Users can view all their bookings

**Performance**:
- Page load <1s
- Filter updates instant

---

### Task 10.5: Create Booking Detail Modal
**Description**: Build modal to show booking details and actions.

**Steps**:
1. Create `features/bookings/components/BookingDetailModal.tsx`
2. Display booking details (desk, time, status)
3. Add "Check In" button (if within check-in window)
4. Add "Cancel Booking" button
5. Add "Modify Booking" button
6. Add "End Early" button (if currently active)
7. Implement each action with API calls
8. Handle button visibility based on booking state
9. Test modal

**Acceptance Criteria**:
- ✅ All booking details displayed
- ✅ "Check In" visible only within check-in window (15min before to 15min after)
- ✅ Check-in updates booking and shows success
- ✅ "Cancel Booking" shows confirmation dialog
- ✅ Cancel updates booking and shows success
- ✅ "Modify Booking" opens wizard pre-filled with current values
- ✅ "End Early" button visible only for active bookings
- ✅ End early updates booking and shows success
- ✅ Modal closes on successful action

**Technical Criteria**:
- POST /api/v1/bookings/:id/check-in
- DELETE /api/v1/bookings/:id (cancel)
- POST /api/v1/bookings/:id/end-early
- TanStack Query mutation invalidates bookings cache

**User-Facing Criteria**:
- Users can manage bookings easily

**Performance**:
- Actions complete <500ms

---

### Task 10.6: Create Booking Modification Flow
**Description**: Implement modification feature to change booking time/desk.

**Steps**:
1. Add modify handler in BookingDetailModal
2. Open booking wizard pre-filled with current values
3. Allow user to change time and/or desk
4. Submit modification as PATCH request
5. Handle success (update booking, show notification)
6. Handle conflict (show suggestions)
7. Test modification flow

**Acceptance Criteria**:
- ✅ Clicking "Modify Booking" opens wizard with current values
- ✅ User can change date, time, duration, or desk
- ✅ Modification validated (booking window, daily limit)
- ✅ PATCH request sent to API
- ✅ Success updates booking and shows notification
- ✅ Conflict shows error with suggestions
- ✅ Bookings list refreshed after modification

**Technical Criteria**:
- PATCH /api/v1/bookings/:id
- Request body includes updated fields
- Conflict response handled same as creation

**User-Facing Criteria**:
- Users can modify bookings easily

**Performance**:
- Modification <500ms

---

### Task 10.7: Create Quick Rebook Feature
**Description**: Add "Book Again" button to past bookings.

**Steps**:
1. Add "Book Again" button to BookingDetailModal for past bookings
2. Copy booking details (time, desk) to wizard
3. Open wizard with pre-filled values
4. Allow user to confirm or modify
5. Test rebook flow

**Acceptance Criteria**:
- ✅ "Book Again" visible only for past bookings
- ✅ Clicking opens wizard with same time/desk
- ✅ User can modify values before confirming
- ✅ Booking created successfully

**Technical Criteria**:
- Same wizard component reused
- Initial values passed as props

**User-Facing Criteria**:
- Users can rebook past bookings easily

---

## Phase 11: Frontend - Admin Dashboard

### Task 11.1: Create Admin Navigation
**Description**: Build admin-only navigation menu.

**Steps**:
1. Create admin layout component
2. Add navigation links: Desks, Users, Bookings, Settings, Audit Log
3. Protect admin routes with role check
4. Add admin badge to user menu
5. Test navigation

**Acceptance Criteria**:
- ✅ Admin navigation visible only to admin users
- ✅ All admin pages accessible from navigation
- ✅ Non-admin users redirected to dashboard
- ✅ Active link highlighted

**Technical Criteria**:
- Role check in auth context
- ProtectedRoute wrapper checks for admin role

**User-Facing Criteria**:
- Admins can access admin features

---

### Task 11.2: Create Desk Management Page
**Description**: Build CRUD interface for desk management.

**Steps**:
1. Create `app/admin/desks/page.tsx`
2. Display desks in table view
3. Add "Create Desk" button (opens modal)
4. Add edit and delete buttons for each desk
5. Implement create desk form (desk_number, wing)
6. Implement edit desk form (desk_number, wing, status)
7. Add maintenance checkbox (sets status to 'maintenance')
8. Handle delete with confirmation
9. Show auto-cancel warning for maintenance
10. Test CRUD operations

**Acceptance Criteria**:
- ✅ Desks displayed in table with all fields
- ✅ Create desk modal validates desk_number uniqueness
- ✅ Edit desk modal allows status change
- ✅ Setting maintenance shows warning about auto-cancellations
- ✅ Delete shows confirmation dialog
- ✅ Delete prevented if active bookings exist
- ✅ Success notifications shown for all operations
- ✅ Table refreshed after each operation
- ✅ Loading states shown during operations

**Technical Criteria**:
- POST /api/v1/desks (create)
- PATCH /api/v1/desks/:id (update)
- DELETE /api/v1/desks/:id (delete)
- TanStack Query mutations with cache invalidation

**User-Facing Criteria**:
- Admins can manage desks easily

**Performance**:
- Operations complete <500ms

---

### Task 11.3: Create User Management Page
**Description**: Build interface for viewing and managing users.

**Steps**:
1. Create `app/admin/users/page.tsx`
2. Display users in table (email, role, status, created_at)
3. Add filters (role, status)
4. Add "Disable Account" button
5. Add "Reset Password" button (admin sets new password)
6. Implement disable account action
7. Implement reset password form
8. Test user management

**Acceptance Criteria**:
- ✅ Users displayed in table with pagination
- ✅ Filters work correctly
- ✅ Disable account shows confirmation dialog
- ✅ Disable sets user status to 'disabled'
- ✅ Reset password opens modal with new password form
- ✅ Reset password updates user's password
- ✅ Success notifications shown
- ✅ Table refreshed after operations

**Technical Criteria**:
- GET /api/v1/admin/users
- PATCH /api/v1/admin/users/:id
- Pagination with offset/limit

**User-Facing Criteria**:
- Admins can manage user accounts

**Performance**:
- User list load <500ms
- Pagination instant

---

### Task 11.4: Create System Settings Page
**Description**: Build interface for configuring global settings.

**Steps**:
1. Create `app/admin/settings/page.tsx`
2. Fetch current settings from API
3. Create form with fields: opening hours (start/end), daily hour limit, grace period
4. Add validation (opening hours valid time, limits positive integers)
5. Implement save settings action
6. Test settings update

**Acceptance Criteria**:
- ✅ Settings form displays current values
- ✅ Opening hours validated (end > start)
- ✅ Daily hour limit validated (positive integer)
- ✅ Grace period validated (positive integer in minutes)
- ✅ Save button updates settings via API
- ✅ Success notification shown
- ✅ Form reloaded with new values

**Technical Criteria**:
- GET /api/v1/admin/settings
- PATCH /api/v1/admin/settings
- Form validation with Zod

**User-Facing Criteria**:
- Admins can configure system settings

**Performance**:
- Settings save <500ms

---

### Task 11.5: Create All Bookings View (Admin)
**Description**: Build interface for admins to view all bookings with filters.

**Steps**:
1. Create `app/admin/bookings/page.tsx`
2. Fetch all bookings from API
3. Display in table with columns: user, desk, date, time, status
4. Add filters: date range, user (search), desk, status
5. Add pagination (20 per page)
6. Add click handler to view booking details
7. Test filters and pagination

**Acceptance Criteria**:
- ✅ All bookings displayed in table
- ✅ Date range filter works
- ✅ User search filter works
- ✅ Desk filter works
- ✅ Status filter works
- ✅ Pagination works (20 per page)
- ✅ Clicking booking opens detail modal
- ✅ Loading state shown

**Technical Criteria**:
- GET /api/v1/admin/bookings?page=1&limit=20&...filters
- TanStack Query with pagination

**User-Facing Criteria**:
- Admins can view and filter all bookings

**Performance**:
- Booking list load <1s
- Filter updates <500ms

---

### Task 11.6: Create Force Cancel Feature (Admin)
**Description**: Add admin ability to force cancel any booking with reason.

**Steps**:
1. Add "Force Cancel" button to booking detail modal (admin only)
2. Create cancel confirmation modal with reason text field
3. Implement force cancel API call
4. Create notification for affected user
5. Test force cancel

**Acceptance Criteria**:
- ✅ "Force Cancel" button visible only to admins
- ✅ Clicking shows confirmation modal with reason field
- ✅ Reason field required
- ✅ Cancel API call includes reason
- ✅ Booking cancelled successfully
- ✅ Notification sent to affected user
- ✅ Success notification shown to admin
- ✅ Booking list refreshed

**Technical Criteria**:
- DELETE /api/v1/admin/bookings/:id with reason in request body
- Audit log created with reason in metadata

**User-Facing Criteria**:
- Admins can cancel any booking with justification

**Performance**:
- Force cancel <500ms

---

### Task 11.7: Create Audit Log Viewer (Admin)
**Description**: Build read-only interface for viewing audit logs.

**Steps**:
1. Create `app/admin/audit/page.tsx`
2. Fetch audit logs from API
3. Display in table: timestamp, user, entity_type, action, changes
4. Add filters: date range, user, entity_type, action
5. Add pagination
6. Add expandable row to view full changes JSON
7. Test audit log viewer

**Acceptance Criteria**:
- ✅ Audit logs displayed in table
- ✅ Date range filter works
- ✅ User filter works
- ✅ Entity type filter works
- ✅ Action filter works
- ✅ Pagination works
- ✅ Expanding row shows full changes and metadata
- ✅ Read-only (no edit/delete)
- ✅ Loading state shown

**Technical Criteria**:
- GET /api/v1/admin/audit-logs?page=1&limit=20&...filters
- JSON viewer for changes column

**User-Facing Criteria**:
- Admins can review all system changes

**Performance**:
- Audit log load <1s

---

## Phase 12: Testing

### Task 12.1: Write Backend Unit Tests - Auth Service
**Description**: Create comprehensive unit tests for auth service layer.

**Steps**:
1. Create `internal/features/auth/service_test.go`
2. Mock repository layer
3. Test Register function (success, duplicate email, validation errors)
4. Test Login function (success, wrong password, disabled user)
5. Test RefreshToken function (success, invalid token, expired token)
6. Test Logout and LogoutAll functions
7. Achieve >80% code coverage

**Acceptance Criteria**:
- ✅ All auth service functions tested
- ✅ Success cases covered
- ✅ Error cases covered
- ✅ Edge cases covered
- ✅ Code coverage >80%
- ✅ All tests pass
- ✅ Tests run in <5 seconds

**Technical Criteria**:
- Mock repository using interface
- Test isolation (no database dependencies)

**User-Facing Criteria**:
- Auth logic verified to work correctly

---

### Task 12.2: Write Backend Unit Tests - Booking Service
**Description**: Create comprehensive unit tests for booking service layer.

**Steps**:
1. Create `internal/features/bookings/service_test.go`
2. Mock repository layer
3. Test CreateBooking (success, conflict, validation errors, daily limit)
4. Test CheckInBooking (success, wrong time, already checked in)
5. Test CancelBooking (success, unauthorized)
6. Test UpdateBooking (success, conflict)
7. Test EndBookingEarly (success, not active)
8. Achieve >80% code coverage

**Acceptance Criteria**:
- ✅ All booking service functions tested
- ✅ Conflict detection tested
- ✅ Daily limit validation tested
- ✅ Booking window validation tested
- ✅ Check-in window validation tested
- ✅ Code coverage >80%
- ✅ All tests pass
- ✅ Tests run in <10 seconds

**Technical Criteria**:
- Mock repository and audit service
- Test time-based logic with fixed time values

**User-Facing Criteria**:
- Booking logic verified to work correctly

---

### Task 12.3: Write Backend Integration Tests - Booking Flow
**Description**: Create integration tests for full booking creation flow.

**Steps**:
1. Create `internal/features/bookings/integration_test.go`
2. Setup test database with Docker
3. Test full booking creation flow (happy path)
4. Test booking conflict (attempt overlapping booking)
5. Test booking modification
6. Test booking cancellation
7. Use transactions for test isolation (rollback after each test)

**Acceptance Criteria**:
- ✅ Integration tests use real PostgreSQL database
- ✅ Booking creation end-to-end tested
- ✅ Exclusion constraint verified
- ✅ Conflict suggestions generated
- ✅ Modification flow tested
- ✅ Cancellation flow tested
- ✅ All tests pass
- ✅ Tests run in <30 seconds
- ✅ Database cleaned up after tests (transaction rollback)

**Technical Criteria**:
- Docker Compose used for test database
- Transactions used for isolation
- Parallel test execution disabled for DB tests

**User-Facing Criteria**:
- Core booking flow verified end-to-end

**Performance**:
- Integration tests complete <30s

---

### Task 12.4: Write Backend Integration Tests - Auth Flow
**Description**: Create integration tests for authentication flow.

**Steps**:
1. Create `internal/features/auth/integration_test.go`
2. Test registration → login → refresh → logout flow
3. Test invalid credentials
4. Test token expiry handling
5. Use transactions for test isolation

**Acceptance Criteria**:
- ✅ Full auth flow tested end-to-end
- ✅ Token generation and validation tested
- ✅ Session management tested
- ✅ All tests pass
- ✅ Tests run in <20 seconds

**Technical Criteria**:
- Real database used
- Bcrypt hashing tested
- JWT generation tested

**User-Facing Criteria**:
- Auth flow verified end-to-end

---

### Task 12.5: Write Frontend Component Tests - Booking Wizard
**Description**: Create component tests for booking wizard using Vitest and RTL.

**Steps**:
1. Create `features/bookings/components/BookingWizard.test.tsx`
2. Test Step 1: date/time selection and validation
3. Test Step 2: desk selection
4. Test Step 3: review and confirmation
5. Test wizard navigation (next, back)
6. Mock API calls
7. Test error handling

**Acceptance Criteria**:
- ✅ All wizard steps render correctly
- ✅ Form validation works
- ✅ Navigation between steps works
- ✅ API calls mocked correctly
- ✅ Success and error states tested
- ✅ All tests pass
- ✅ Tests run in <5 seconds

**Technical Criteria**:
- Vitest + React Testing Library
- Mock TanStack Query hooks
- Test user interactions (clicks, form inputs)

**User-Facing Criteria**:
- Booking wizard verified to work correctly

---

### Task 12.6: Write Frontend Component Tests - Auth Forms
**Description**: Create component tests for login and registration forms.

**Steps**:
1. Create `features/auth/components/LoginForm.test.tsx`
2. Create `features/auth/components/RegisterForm.test.tsx`
3. Test form rendering
4. Test form validation
5. Test successful submission
6. Test error handling
7. Mock API calls

**Acceptance Criteria**:
- ✅ Login and registration forms render correctly
- ✅ Form validation tested
- ✅ Successful submission tested
- ✅ Error messages displayed correctly
- ✅ All tests pass
- ✅ Tests run in <5 seconds

**Technical Criteria**:
- Mock auth context
- Mock API calls
- Test form interactions

**User-Facing Criteria**:
- Auth forms verified to work correctly

---

### Task 12.7: Write E2E Tests - Full Booking Flow (Playwright)
**Description**: Create end-to-end tests for booking flow using Playwright.

**Steps**:
1. Setup Playwright test environment
2. Create `tests/e2e/booking-flow.spec.ts`
3. Test: Login → Browse Desks → Create Booking → Verify in My Bookings
4. Test: Booking Conflict → View Suggestions
5. Test: Cancel Booking
6. Test: Check-in Booking
7. Run tests against local development environment

**Acceptance Criteria**:
- ✅ E2E test creates booking successfully
- ✅ Booking appears in My Bookings
- ✅ Conflict detection tested
- ✅ Suggestions displayed correctly
- ✅ Cancellation flow tested
- ✅ Check-in flow tested
- ✅ All tests pass
- ✅ Tests run in <60 seconds
- ✅ Screenshots captured on failure

**Technical Criteria**:
- Playwright configured with chromium
- Test database seeded with test data
- Tests isolated (cleanup after each test)

**User-Facing Criteria**:
- Full user journey verified

**Performance**:
- E2E tests complete <60s

---

### Task 12.8: Write E2E Tests - Admin Flow (Playwright)
**Description**: Create end-to-end tests for admin features.

**Steps**:
1. Create `tests/e2e/admin-flow.spec.ts`
2. Test: Admin Login → Create Desk → Verify Desk List
3. Test: Admin Force Cancel Booking → Verify User Notification
4. Test: Admin Set Desk to Maintenance → Verify Auto-Cancellations
5. Run tests against local development environment

**Acceptance Criteria**:
- ✅ Admin can create desk
- ✅ Admin can force cancel booking
- ✅ Force cancel creates notification
- ✅ Setting maintenance auto-cancels bookings
- ✅ All tests pass
- ✅ Tests run in <60 seconds

**Technical Criteria**:
- Admin user created in test database
- Tests isolated and cleaned up

**User-Facing Criteria**:
- Admin features verified

---

### Task 12.9: Create Test Data Seed Script
**Description**: Create script to seed database with test data for local development.

**Steps**:
1. Create `scripts/seed.js` (or seed.go)
2. Create 1 admin user (admin@example.com / admin123)
3. Create 3 regular users (user1-3@example.com / password123)
4. Create 20 desks (E-101 to E-110, W-201 to W-210)
5. Create 10 sample bookings (mix of upcoming, past, cancelled)
6. Add CLI command to run seed script
7. Test seed script

**Acceptance Criteria**:
- ✅ Seed script runs successfully
- ✅ Creates admin and regular users
- ✅ Creates desks in both wings
- ✅ Creates sample bookings
- ✅ Script idempotent (can run multiple times)
- ✅ Script can be run with `npm run seed` or `make seed`

**Technical Criteria**:
- Script connects to local database
- Uses environment variables for connection
- Clears existing data before seeding (optional flag)

**User-Facing Criteria**:
- Developers can seed local database easily

**Performance**:
- Seed script runs <10s

---

## Phase 13: CI/CD & Deployment

### Task 13.1: Create GitHub Actions CI Workflow - Backend
**Description**: Setup CI workflow for backend testing and linting.

**Steps**:
1. Create `.github/workflows/backend-ci.yml`
2. Configure PostgreSQL and Redis services
3. Add Go setup and dependency caching
4. Add golangci-lint step
5. Add unit test step
6. Add integration test step
7. Add build step
8. Configure to run on pull requests and main branch
9. Test workflow

**Acceptance Criteria**:
- ✅ Workflow runs on every pull request
- ✅ Workflow runs on push to main
- ✅ All steps complete successfully
- ✅ Linting catches code issues
- ✅ All tests pass
- ✅ Build succeeds
- ✅ Workflow completes in <10 minutes
- ✅ Failed checks block PR merge

**Technical Criteria**:
- PostgreSQL 16 service container
- Redis 7 service container
- Go version 1.22+
- golangci-lint latest version

**User-Facing Criteria**:
- Code quality verified automatically

**Performance**:
- CI pipeline <10min

---

### Task 13.2: Create GitHub Actions CI Workflow - Frontend
**Description**: Setup CI workflow for frontend testing and linting.

**Steps**:
1. Create `.github/workflows/frontend-ci.yml`
2. Add Node.js setup and dependency caching
3. Add ESLint step
4. Add TypeScript type-check step
5. Add Vitest unit test step
6. Add Next.js build step
7. Configure to run on pull requests and main branch
8. Test workflow

**Acceptance Criteria**:
- ✅ Workflow runs on every pull request
- ✅ Workflow runs on push to main
- ✅ All steps complete successfully
- ✅ ESLint catches code issues
- ✅ Type-check catches TypeScript errors
- ✅ All tests pass
- ✅ Build succeeds
- ✅ Workflow completes in <10 minutes
- ✅ Failed checks block PR merge

**Technical Criteria**:
- Node.js version 20+
- npm ci for reproducible builds
- Cache node_modules

**User-Facing Criteria**:
- Frontend code quality verified automatically

**Performance**:
- CI pipeline <10min

---

### Task 13.3: Setup Render Account and Projects
**Description**: Configure Render account and create staging/production environments.

**Steps**:
1. Create Render account
2. Create PostgreSQL database (staging)
3. Create PostgreSQL database (production)
4. Create Redis instance (staging)
5. Create Redis instance (production)
6. Create Web Service for backend (staging)
7. Create Web Service for backend (production)
8. Create Web Service for frontend (staging)
9. Create Web Service for frontend (production)
10. Configure environment variables for each service
11. Test services start successfully

**Acceptance Criteria**:
- ✅ Render account created
- ✅ All services created (staging and production)
- ✅ Environment variables configured
- ✅ Staging services accessible
- ✅ Production services accessible
- ✅ Database backups configured (daily)
- ✅ SSL certificates configured (auto)

**Technical Criteria**:
- PostgreSQL version 16
- Redis version 7
- Web Services use Docker (backend) and static site (frontend)

**User-Facing Criteria**:
- Staging and production environments ready

---

### Task 13.4: Create Database Migration Job in CI/CD
**Description**: Setup automated database migration in CI/CD pipeline.

**Steps**:
1. Create GitHub Actions workflow for migrations
2. Add migration step (runs goose up)
3. Configure to run before deployment
4. Add dry-run validation in PR checks
5. Test migration workflow

**Acceptance Criteria**:
- ✅ Migration workflow runs before deployment
- ✅ Migrations applied successfully
- ✅ Dry-run validates migrations in PRs
- ✅ Failed migrations block deployment
- ✅ Migration logs captured

**Technical Criteria**:
- goose CLI used
- Connection string from environment
- Migrations run against staging before production

**User-Facing Criteria**:
- Database schema updated automatically

**Performance**:
- Migration job <5min

---

### Task 13.5: Create Staging Deployment Workflow
**Description**: Setup auto-deployment to staging on merge to main.

**Steps**:
1. Create `.github/workflows/deploy-staging.yml`
2. Configure to trigger on push to main
3. Add step to run migrations
4. Add step to deploy backend to Render
5. Add step to deploy frontend to Render
6. Add health check validation
7. Test deployment workflow

**Acceptance Criteria**:
- ✅ Workflow triggers on merge to main
- ✅ Migrations run successfully
- ✅ Backend deployed to Render staging
- ✅ Frontend deployed to Render staging
- ✅ Health checks pass
- ✅ Deployment completes in <15 minutes
- ✅ Rollback triggered on failure

**Technical Criteria**:
- Render deploy hooks used
- Health check endpoint validated post-deployment

**User-Facing Criteria**:
- Staging automatically updated on merge

**Performance**:
- Deployment <15min

---

### Task 13.6: Create Production Deployment Workflow
**Description**: Setup manual promotion to production with approval.

**Steps**:
1. Create `.github/workflows/deploy-production.yml`
2. Configure manual trigger (workflow_dispatch)
3. Add approval gate
4. Add step to run migrations
5. Add step to deploy backend to Render
6. Add step to deploy frontend to Render
7. Add health check validation
8. Test deployment workflow

**Acceptance Criteria**:
- ✅ Workflow requires manual trigger
- ✅ Approval required before deployment
- ✅ Migrations run successfully
- ✅ Backend deployed to Render production
- ✅ Frontend deployed to Render production
- ✅ Health checks pass
- ✅ Deployment completes in <15 minutes
- ✅ Rollback option available

**Technical Criteria**:
- GitHub environments used for approval
- Production environment configured in repo settings

**User-Facing Criteria**:
- Production deployments controlled and verified

**Performance**:
- Deployment <15min

---

### Task 13.7: Setup Error Tracking with Sentry
**Description**: Integrate Sentry for error tracking in staging and production.

**Steps**:
1. Create Sentry account (free tier)
2. Create projects for backend and frontend
3. Install Sentry SDK in backend
4. Install Sentry SDK in frontend
5. Configure Sentry DSN in environment variables
6. Add error reporting in backend
7. Add error reporting in frontend
8. Test error tracking
9. Configure error alerts

**Acceptance Criteria**:
- ✅ Sentry account created
- ✅ Backend errors reported to Sentry
- ✅ Frontend errors reported to Sentry
- ✅ Source maps uploaded for frontend
- ✅ Error alerts configured
- ✅ Test error captured successfully
- ✅ Sentry dashboard accessible

**Technical Criteria**:
- Sentry SDK integrated in error handlers
- Source maps uploaded in production builds
- User context attached to errors

**User-Facing Criteria**:
- Errors tracked for debugging

---

### Task 13.8: Setup OpenTelemetry Tracing
**Description**: Integrate OpenTelemetry for distributed tracing.

**Steps**:
1. Install OpenTelemetry SDK in backend
2. Configure OTLP exporter (to Sentry or local)
3. Add tracing to HTTP handlers
4. Add tracing to database queries
5. Add trace context propagation
6. Test tracing in local and staging
7. Verify traces in observability platform

**Acceptance Criteria**:
- ✅ OpenTelemetry SDK configured
- ✅ Traces exported to observability platform
- ✅ HTTP requests traced
- ✅ Database queries traced
- ✅ Trace context propagated across services
- ✅ Traces viewable in platform

**Technical Criteria**:
- OTLP exporter configured
- Sampling rate configured (100% in staging, 10% in production)
- Trace IDs logged with structured logs

**User-Facing Criteria**:
- Performance issues traceable

---

### Task 13.9: Configure Structured Logging
**Description**: Setup structured JSON logging with zap.

**Steps**:
1. Configure zap logger in backend
2. Add request_id middleware
3. Log all HTTP requests (method, path, status, duration)
4. Log database queries with timing
5. Log authentication attempts
6. Log booking operations
7. Configure log levels (ERROR, WARN, INFO, DEBUG)
8. Test logging in local and staging

**Acceptance Criteria**:
- ✅ All logs output as structured JSON
- ✅ All HTTP requests logged
- ✅ Request IDs included in all logs
- ✅ Error logs include stack traces
- ✅ Log levels configurable via environment
- ✅ Sensitive data not logged (passwords, tokens)
- ✅ Logs searchable in Render dashboard

**Technical Criteria**:
- zap.Logger used throughout
- Log format: `{"level":"info","timestamp":"2025-01-25T10:00:00Z","msg":"...","request_id":"..."}`

**User-Facing Criteria**:
- Debugging information available

---

## Phase 14: Documentation

### Task 14.1: Write README.md
**Description**: Create comprehensive README with setup instructions.

**Steps**:
1. Create `README.md` in root directory
2. Add project overview
3. Add tech stack summary
4. Add local development setup instructions
5. Add environment variables documentation
6. Add testing instructions
7. Add deployment instructions
8. Add screenshots (optional)
9. Add links to other docs
10. Review and refine

**Acceptance Criteria**:
- ✅ README includes project overview
- ✅ Setup instructions clear and complete
- ✅ Environment variables documented
- ✅ Testing commands documented
- ✅ Links to other docs included
- ✅ Formatted with proper headings and code blocks
- ✅ New developer can set up project using README

**Technical Criteria**:
- Markdown formatting
- Code blocks with syntax highlighting
- Table of contents

**User-Facing Criteria**:
- Developers can onboard easily

---

### Task 14.2: Write CONTRIBUTING.md
**Description**: Create contribution guidelines.

**Steps**:
1. Create `CONTRIBUTING.md`
2. Document git workflow (branch naming, commit messages)
3. Document PR process and template
4. Document code quality tools and standards
5. Document testing requirements
6. Add code of conduct (optional)
7. Review and refine

**Acceptance Criteria**:
- ✅ Git workflow documented
- ✅ Branch naming conventions clear
- ✅ Commit message format documented (conventional commits)
- ✅ PR process documented
- ✅ Testing requirements clear
- ✅ Code quality standards documented

**Technical Criteria**:
- Markdown formatting
- Links to external resources

**User-Facing Criteria**:
- Contributors understand process

---

### Task 14.3: Write API Documentation (OpenAPI Spec)
**Description**: Create OpenAPI specification for REST API.

**Steps**:
1. Create `docs/openapi.yaml`
2. Document all API endpoints with request/response schemas
3. Add authentication requirements
4. Add error response examples
5. Add example requests/responses
6. Validate OpenAPI spec
7. Generate API docs UI (Swagger UI or similar)
8. Host API docs

**Acceptance Criteria**:
- ✅ All endpoints documented
- ✅ Request/response schemas defined
- ✅ Authentication documented
- ✅ Error responses documented
- ✅ Examples provided
- ✅ OpenAPI spec validates
- ✅ Docs UI accessible (local or hosted)

**Technical Criteria**:
- OpenAPI 3.0+ specification
- Spec validates with openapi-validator
- Swagger UI or Redoc for rendering

**User-Facing Criteria**:
- API usage clear for developers

---

### Task 14.4: Write ARCHITECTURE.md
**Description**: Document system architecture and design decisions.

**Steps**:
1. Create `ARCHITECTURE.md`
2. Add system overview diagram
3. Document backend architecture (layers, modules)
4. Document frontend architecture (features, components)
5. Document data flow for key operations (booking creation, auth)
6. Document database schema with ERD
7. Document key design decisions and rationale
8. Review and refine

**Acceptance Criteria**:
- ✅ System overview clear
- ✅ Backend architecture documented
- ✅ Frontend architecture documented
- ✅ Data flow diagrams included
- ✅ Database schema documented
- ✅ Design decisions explained

**Technical Criteria**:
- Diagrams in Mermaid or image format
- Links to relevant code

**User-Facing Criteria**:
- Developers understand system design

---

### Task 14.5: Write DEPLOYMENT.md
**Description**: Document deployment process and infrastructure.

**Steps**:
1. Create `DEPLOYMENT.md`
2. Document Render configuration
3. Document environment variables for each service
4. Document deployment workflow (staging and production)
5. Document rollback process
6. Document monitoring and observability setup
7. Document backup and recovery process
8. Review and refine

**Acceptance Criteria**:
- ✅ Render setup documented
- ✅ Environment variables documented
- ✅ Deployment workflow documented
- ✅ Rollback process documented
- ✅ Monitoring setup documented
- ✅ Backup/recovery documented

**Technical Criteria**:
- Step-by-step instructions
- Screenshots or examples

**User-Facing Criteria**:
- Operations team can deploy and maintain system

---

### Task 14.6: Create Database Schema Diagram (ERD)
**Description**: Create entity relationship diagram for database schema.

**Steps**:
1. Choose diagramming tool (Mermaid, dbdiagram.io, or draw.io)
2. Create ERD showing all tables
3. Show relationships (foreign keys)
4. Show key columns and types
5. Export as image or Mermaid code
6. Add to ARCHITECTURE.md
7. Review and refine

**Acceptance Criteria**:
- ✅ All tables included
- ✅ Relationships shown correctly
- ✅ Key columns visible
- ✅ Diagram clear and readable
- ✅ Embedded in ARCHITECTURE.md

**Technical Criteria**:
- ERD matches actual schema
- Foreign keys clearly indicated

**User-Facing Criteria**:
- Database structure understandable

---

## Phase 15: Launch Preparation

### Task 15.1: Create Initial Admin Account
**Description**: Manually create first admin account in production.

**Steps**:
1. Access production database
2. Insert admin user with hashed password
3. Set role to 'admin'
4. Verify admin can log in
5. Document admin credentials securely

**Acceptance Criteria**:
- ✅ Admin account created
- ✅ Admin can log in to production
- ✅ Admin has full access to admin features
- ✅ Credentials documented securely

**Technical Criteria**:
- Password hashed with bcrypt cost 12
- Email: admin@example.com (or actual admin email)

**User-Facing Criteria**:
- Admin can manage system

---

### Task 15.2: Seed Production Database with Initial Desks
**Description**: Create initial set of desks for launch.

**Steps**:
1. Prepare list of desks (desk numbers, wings)
2. Use admin interface or seed script to create desks
3. Verify desks appear in desk list
4. Test booking a desk

**Acceptance Criteria**:
- ✅ All initial desks created
- ✅ Desks visible in desk list
- ✅ Desks bookable
- ✅ Desk numbers follow naming convention

**Technical Criteria**:
- Desk numbers unique
- Wings assigned correctly

**User-Facing Criteria**:
- Desks available for booking

---

### Task 15.3: Perform Staging Smoke Tests
**Description**: Manual testing of critical flows in staging environment.

**Steps**:
1. Test user registration and login
2. Test desk browsing and availability view
3. Test booking creation (happy path)
4. Test booking conflict detection
5. Test booking modification
6. Test booking cancellation
7. Test check-in flow
8. Test admin desk management
9. Test admin user management
10. Test admin force cancel
11. Document test results

**Acceptance Criteria**:
- ✅ All critical flows tested
- ✅ No blocking bugs found
- ✅ Performance acceptable (<500ms for booking creation)
- ✅ UI responsive on desktop and mobile
- ✅ Error messages clear and helpful
- ✅ Test results documented

**Technical Criteria**:
- Tests performed in staging environment
- Different user roles tested

**User-Facing Criteria**:
- App ready for production

**Performance**:
- Booking creation <500ms
- Page loads <1s

---

### Task 15.4: Load Test Booking Creation Endpoint
**Description**: Perform load testing on booking creation to verify performance targets.

**Steps**:
1. Install k6 load testing tool
2. Create load test script for booking creation
3. Simulate 20 concurrent users
4. Run test for 5 minutes
5. Analyze results (response time, error rate)
6. Verify <500ms response time
7. Verify zero double-bookings
8. Document results

**Acceptance Criteria**:
- ✅ Load test script created
- ✅ Test simulates 20 concurrent users
- ✅ 95th percentile response time <500ms
- ✅ Zero double-bookings detected
- ✅ Error rate <1%
- ✅ Results documented

**Technical Criteria**:
- k6 script with realistic booking scenarios
- Test against staging environment

**User-Facing Criteria**:
- System handles expected load

**Performance**:
- P95 response time <500ms
- Zero conflicts

---

### Task 15.5: Security Audit and Hardening
**Description**: Perform security review and implement hardening measures.

**Steps**:
1. Run npm audit and fix vulnerabilities
2. Run gosec for Go security issues
3. Review CORS configuration
4. Review authentication implementation
5. Review authorization checks
6. Review SQL injection prevention
7. Review rate limiting configuration
8. Add security headers (CSP, X-Frame-Options, etc.)
9. Document security measures

**Acceptance Criteria**:
- ✅ No high/critical npm vulnerabilities
- ✅ No security issues in Go code
- ✅ CORS configured correctly
- ✅ All routes properly authenticated
- ✅ Authorization checks in place
- ✅ SQL injection prevented (parameterized queries)
- ✅ Rate limiting active
- ✅ Security headers configured
- ✅ Security measures documented

**Technical Criteria**:
- npm audit shows 0 high/critical
- gosec passes with no issues
- HTTPS enforced in production

**User-Facing Criteria**:
- User data protected

---

### Task 15.6: Create Terms of Service and Privacy Policy Placeholders
**Description**: Add placeholder pages for legal documents.

**Steps**:
1. Create `app/terms/page.tsx`
2. Create `app/privacy/page.tsx`
3. Add placeholder content
4. Add links in footer
5. Mark as "placeholder - to be updated before launch"

**Acceptance Criteria**:
- ✅ Terms page created
- ✅ Privacy page created
- ✅ Placeholder content added
- ✅ Links in footer work
- ✅ Pages clearly marked as placeholders

**Technical Criteria**:
- Simple static pages

**User-Facing Criteria**:
- Legal pages accessible

---

### Task 15.7: Setup Monitoring Alerts
**Description**: Configure alerts for critical errors and downtime.

**Steps**:
1. Configure Sentry alerts for error rate spikes
2. Configure alerts for health check failures
3. Configure alerts for database connection failures
4. Set up notification channels (email, Slack optional)
5. Test alerts

**Acceptance Criteria**:
- ✅ Error rate alerts configured
- ✅ Health check alerts configured
- ✅ Database alerts configured
- ✅ Notifications sent to correct channels
- ✅ Test alerts received

**Technical Criteria**:
- Alert thresholds reasonable
- Alert fatigue avoided

**User-Facing Criteria**:
- Issues detected quickly

---

### Task 15.8: Create Launch Checklist
**Description**: Create final pre-launch checklist.

**Steps**:
1. Create `LAUNCH_CHECKLIST.md`
2. List all pre-launch tasks
3. Include verification steps
4. Include rollback plan
5. Include post-launch monitoring plan
6. Review with team (if applicable)

**Acceptance Criteria**:
- ✅ Checklist comprehensive
- ✅ All critical items included
- ✅ Verification steps clear
- ✅ Rollback plan documented
- ✅ Monitoring plan included

**Technical Criteria**:
- Markdown checklist with checkboxes

**User-Facing Criteria**:
- Launch process organized

---

### Task 15.9: Production Launch
**Description**: Deploy to production and monitor.

**Steps**:
1. Review launch checklist
2. Trigger production deployment
3. Verify migrations applied successfully
4. Verify services healthy
5. Perform smoke tests in production
6. Create test booking
7. Monitor error rates and logs
8. Announce launch (internal/external)
9. Monitor for 24 hours

**Acceptance Criteria**:
- ✅ Production deployment successful
- ✅ All services healthy
- ✅ Smoke tests pass
- ✅ Test booking created successfully
- ✅ Error rates normal
- ✅ No critical bugs in first 24 hours
- ✅ Response times <500ms

**Technical Criteria**:
- All health checks passing
- Zero double-bookings
- All monitoring active

**User-Facing Criteria**:
- App live and accessible
- Users can create bookings

**Performance**:
- Booking creation <500ms
- Page loads <1s
- Uptime >99.9%

---

### Task 15.10: Post-Launch Monitoring and Support
**Description**: Monitor system for first week and address issues.

**Steps**:
1. Monitor error rates daily
2. Monitor performance metrics daily
3. Review user feedback (if available)
4. Address bugs as they arise
5. Document common issues and resolutions
6. Prepare for Phase 2 planning

**Acceptance Criteria**:
- ✅ Daily monitoring performed
- ✅ All critical bugs fixed within 24 hours
- ✅ Performance targets maintained
- ✅ User feedback reviewed
- ✅ Issues documented

**Technical Criteria**:
- Error rates <1%
- Response times <500ms
- Uptime >99.9%

**User-Facing Criteria**:
- Stable, reliable system
- Issues resolved quickly

---

## Summary

**Total Phases**: 15
**Total Tasks**: 110+

**Estimated Timeline** (solo development):
- Phase 1-2 (Setup & Database): 1-2 weeks
- Phase 3-6 (Backend): 3-4 weeks
- Phase 7-11 (Frontend): 4-5 weeks
- Phase 12 (Testing): 2-3 weeks
- Phase 13-15 (CI/CD, Docs, Launch): 2-3 weeks

**Total Estimated Time**: 12-17 weeks

**Success Criteria Reference**: See IMPLEMENTATION_PLAN.md Section 8.2

---

**Next Steps**: Begin with Phase 1, Task 1.1 - Initialize Monorepo Structure
