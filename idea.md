### Problem statement

Coworking spaces (or campus labs / study rooms) have **limited desks/rooms** and a lot of **time-based demand**. Today, booking often breaks down because:

* people accidentally **double-book** the same resource,
* recurring bookings are manual and error-prone,
* cancellations don’t release capacity cleanly,
* admins can’t easily enforce policies (buffers, grace periods, max hours),
* users can’t reliably see availability for a time range.

**Core challenge:** build a system that guarantees **no overlapping bookings** for the same desk/room while supporting recurring reservations and policy rules.

---

## Goal

Build a **desk/room booking platform** that lets members reserve resources across time ranges with:

* correct conflict detection (no overlaps)
* recurring bookings (weekly patterns, date ranges)
* cancellation + modification policies
* an auditable, reliable source of truth

---

## Success metrics (what “done well” looks like)

### Correctness

* **0 double-bookings** (hard invariant)
* **100% deterministic conflict resolution** (same inputs → same result)
* **Idempotent operations** (retries don’t create duplicates)

### Product/UX outcomes (measurable)

* **Booking success rate**: % of booking attempts that complete without error (excluding genuine conflicts)
* **Time-to-book**: median time from search → confirmed booking
* **Conflict rate**: % of booking attempts rejected due to conflicts (should drop over time if availability UI is good)
* **Cancellation release latency**: time from cancel → resource visible as available (target: near-instant)

### Reliability & scale (engineering)

* **p95 booking API latency** under load (e.g., 200ms–500ms depending on infra)
* **Throughput**: sustained X booking attempts/sec without violating invariants
* **Recovery**: server restart does not lose bookings; state is consistent

---

## Users

1. **Member** (regular user): books desks/rooms
2. **Admin/Operator**: configures resources and policies, handles exceptions
3. **Space Manager** (optional): views utilization, creates recurring blocks (maintenance)

---

## User stories (by role)

### Member (bookings)

* As a member, I want to **search availability** for a specific time range so I can plan my work session.
* As a member, I want to **book a desk** from 2–5pm so I’m guaranteed a spot.
* As a member, I want to **book recurring sessions** (e.g., every Tue/Thu 7–9pm for 6 weeks) so I don’t rebook manually.
* As a member, I want to **modify** a booking (change time/resource) so I can adapt when plans change.
* As a member, I want to **cancel** a booking and release the desk.
* As a member, I want to **see my upcoming bookings** and their status (confirmed/cancelled/expired).
* As a member, I want a **waitlist** option when my preferred slot is full (optional but fun).

### Member (policies)

* As a member, I want to know **booking rules** (max duration, earliest booking window, cancellation cutoff) so I can comply.
* As a member, I want **fairness protections** (e.g., limit total hours/week) so heavy users don’t monopolize desks (optional).

### Admin / Operator

* As an admin, I want to **create/edit/delete resources** (desks, rooms) and tag them (quiet zone, standing desk).
* As an admin, I want to configure **opening hours**, buffer times, and **maintenance blocks**.
* As an admin, I want to define **policies**:

  * max booking duration
  * max bookings per day/week
  * cancellation rules (e.g., no-cancel within 30 minutes)
  * recurring booking limits
* As an admin, I want to **override** bookings (force cancel / move a booking) with audit logs.
* As an admin, I want to view **utilization metrics** (optional: daily occupancy, peak hours).

---

## Use cases (system flows)

### 1) Search availability (time-range query)

**Given** a time window (start, end), capacity constraints, and optional filters
**When** the user searches
**Then** the system returns:

* available desks/rooms for the whole interval
* optionally, “next available slots” suggestions

Key technical detail: efficient conflict check across time intervals.

---

### 2) Create a booking (single instance)

**When** a member requests Desk A from 2–5pm
**System must:**

* validate policy rules (opening hours, duration, lead time)
* detect conflicts (overlap check)
* reserve atomically (no race-condition double booking)
* return a confirmed booking id

Hard part: concurrency-safe booking creation.

---

### 3) Create a recurring booking (weekly rule)

**When** a member requests “Desk A every Tue 7–9pm for 6 weeks”
**System must:**

* expand recurrence into instances
* check conflicts for each instance
* apply policy limits (max recurring count)
* book all-or-nothing (transactional) or partial with clear reporting (choose one)

Hard part: interval checks at scale + transactional consistency.

---

### 4) Modify booking

**When** user changes time from 2–5pm to 3–6pm
**System must:**

* re-check conflicts and policies
* apply change atomically
* maintain audit history

Hard part: “update” is logically cancel+create but must be safe.

---

### 5) Cancel booking

**When** user cancels
**System must:**

* enforce cancellation policy (cutoff, penalties)
* release capacity immediately
* notify waitlist if supported

---

### 6) No-show / auto-expire (optional)

**When** a booking start time passes and user didn’t check in
**System may:**

* mark no-show
* free the desk after grace period
* record penalties

Hard part: background jobs + eventual consistency.

---

### 7) Admin maintenance block

**When** admin blocks Desk A for repairs next Monday 1–6pm
**System must:**

* prevent new bookings in that window
* optionally auto-cancel/move existing bookings (policy-driven)
* notify impacted users

Hard part: priority rules + cascading effects.

---

## Scope decisions that make this “backend deep”

If you want this to read like a serious system, explicitly define:

* **Invariants**

  * No overlapping confirmed bookings per resource
  * Idempotent booking creation with idempotency keys
  * All recurring instances are consistent with the recurrence rule

* **Consistency model**

  * Strong consistency for booking writes (transactional)
  * Eventual consistency acceptable for analytics

* **Policies**

  * Opening hours + buffer times
  * Max duration and booking limits
  * Cancellation cutoff

---

