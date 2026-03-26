# 🧠 AI AGENT INSTRUCTIONS
## Membership Management System — Pre-Prompt & Execution Guide
### Version: 2.0 | Last Updated: 2026-03-26

---

> **READ THIS ENTIRE FILE BEFORE WRITING ANY CODE.**
> This document is the single source of truth for architecture,
> business rules, state machines, security constraints, and
> coding standards. Every decision you make must be consistent
> with what is written here.

---

## 📌 TABLE OF CONTENTS

1. [Project Identity](#1-project-identity)
2. [Tech Stack](#2-tech-stack)
3. [Monorepo Structure](#3-monorepo-structure)
4. [Roles & Permissions](#4-roles--permissions)
5. [Membership Types](#5-membership-types)
6. [State Machines](#6-state-machines)
7. [Veto Model — Critical Rule](#7-veto-model--critical-rule)
8. [Reference System](#8-reference-system)
9. [Consultation System](#9-consultation-system)
10. [Reputation System](#10-reputation-system)
11. [Voting System](#11-voting-system)
12. [Honorary Membership Flow](#12-honorary-membership-flow)
13. [Token System](#13-token-system)
14. [Email System](#14-email-system)
15. [Database Schema](#15-database-schema)
16. [API Endpoint Map](#16-api-endpoint-map)
17. [Security Rules](#17-security-rules)
18. [Audit & Immutability Rules](#18-audit--immutability-rules)
19. [Frontend Rules](#19-frontend-rules)
20. [Coding Standards](#20-coding-standards)
21. [Implementation Order](#21-implementation-order)
22. [Forbidden Actions](#22-forbidden-actions)

---

## 1. PROJECT IDENTITY

**System Name:** Membership Management System (Üyelik Yönetim Sistemi)

**Purpose:**
A secure, multi-step membership application and approval system.
Supports 5 distinct membership types, each with its own workflow,
voting stages, and termination rules.

**Core Principle:**
> Any single justified RED decision at any stage, from any authorized
> person, immediately and permanently terminates the application.
> This cannot be undone, overridden, or deleted.

---

## 2. TECH STACK

### Backend
```
Language    : Go 1.22+
Framework   : Fiber v2
ORM         : GORM v2
Database    : MySQL 8.0
Auth        : JWT (golang-jwt/jwt v5)
Config      : Viper
Email       : gopkg.in/gomail.v2
Password    : bcrypt (cost: 12)
Token       : UUID v4 + HMAC-SHA256
Validation  : go-playground/validator v10
```

### Frontend
```
Framework   : Next.js 14 (App Router)
Language    : TypeScript (strict mode)
Forms       : React Hook Form + Zod
Tables      : TanStack Table v8
Data fetch  : TanStack Query (React Query) v5
State       : Zustand
UI          : shadcn/ui + Tailwind CSS
Icons       : lucide-react
HTTP client : Axios
```

### Infrastructure
```
Monorepo    : Turborepo
Dev Email   : MailDev (Docker)
DB GUI      : (optional) phpMyAdmin
Containers  : Docker + docker-compose
```

---

## 3. MONOREPO STRUCTURE

```
membership-system/
├── apps/
│   ├── api/                          # Go backend
│   │   ├── cmd/main.go
│   │   ├── config/
│   │   │   ├── config.go             # Viper env loader
│   │   │   └── database.go           # GORM + MySQL init
│   │   ├── internal/
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go           # JWT validation
│   │   │   │   ├── rbac.go           # Role guard
│   │   │   │   ├── audit.go          # Auto-log all mutations
│   │   │   │   └── cors.go
│   │   │   ├── router/
│   │   │   │   └── router.go
│   │   │   ├── features/
│   │   │   │   ├── auth/
│   │   │   │   ├── applications/
│   │   │   │   │   └── statemachine.go   ← CRITICAL
│   │   │   │   ├── references/
│   │   │   │   ├── consultations/
│   │   │   │   ├── reputation/
│   │   │   │   ├── voting/
│   │   │   │   ├── honorary/
│   │   │   │   ├── webpublish/
│   │   │   │   ├── notifications/
│   │   │   │   └── logs/
│   │   │   └── shared/
│   │   │       ├── errors.go
│   │   │       ├── response.go
│   │   │       ├── token.go
│   │   │       └── redguard.go       ← CRITICAL
│   │   └── migrations/
│   │
│   └── admin/                        # Next.js frontend
│       └── src/
│           ├── app/
│           ├── components/
│           ├── lib/
│           └── types/
│
├── libs/
│   ├── shared-types/                 # Shared TS types
│   ├── ui/                           # shadcn/ui component lib
│   └── validators/                   # Shared Zod schemas
│
├── docker-compose.yml
├── turbo.json
├── .env.example
└── package.json
```

### Per-Feature File Convention (Backend)
Every feature folder MUST contain exactly these files:
```
models.go       # GORM models + enums
dtos.go         # Request/Response structs
repository.go   # DB access only — no business logic
service.go      # Business logic only — no HTTP context
handler.go      # HTTP layer only — calls service
```

---

## 4. ROLES & PERMISSIONS

| Role | Turkish Name | Description |
|------|-------------|-------------|
| `admin` | Sistem Yöneticisi | Full system access |
| `yk` | YK Üyesi | Decision maker, can vote and view RED identity |
| `yik` | YİK Üyesi | Veto authority (Onursal only), can propose Honorary |
| `koordinator` | Üyelik Koordinatörü | Manages process, sends emails, adds contacts |
| `asil_uye` | Asil Üye | Can propose Honorary members |

### RBAC Matrix

| Action | admin | yk | yik | koordinator | asil_uye | public |
|--------|-------|----|-----|-------------|----------|--------|
| View all applications | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ |
| View RED voter identity | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Cast YK prelim vote | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Cast YİK vote | ❌ | ❌ | ✅ | ❌ | ❌ | ❌ |
| Cast YK final vote | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Send reference emails | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| Add reputation contacts | ❌ | ✅ | ❌ | ✅ | ❌ | ❌ |
| Propose Honorary | ❌ | ❌ | ✅ | ❌ | ✅ | ❌ |
| Web publish toggle | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| View audit logs | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ |
| Respond via token link | ❌ | ❌ | ❌ | ❌ | ❌ | ✅ |

---

## 5. MEMBERSHIP TYPES

```
asil           → Asil Üye
akademik       → Akademik Üye
profesyonel    → Profesyonel Üye
ogrenci        → Öğrenci Üye
onursal        → Onursal Üye
```

### Required Fields per Type

**Asil & Akademik:**
- Full name, email
- LinkedIn URL (unique — primary identity key)
- Photo (required ONLY if LinkedIn profile has no photo)
- Minimum 3 references (must be existing system members, autocomplete)
- Membership type

**Profesyonel & Öğrenci:**
- Full name, email
- Photo (required)
- Membership type

**Onursal:**
- Nominee name
- Nominee LinkedIn URL
- Proposal reason (written justification)
- Proposed by: logged-in asil_uye or yik_uye

---

## 6. STATE MACHINES

> ⚠️ State transitions MUST be validated in `statemachine.go`.
> Direct status updates via API are FORBIDDEN.
> Only service methods may advance state.

### 6.1 Asil & Akademik State Flow

```
başvuru_alındı
    │ (system: send ref emails)
    ▼
referans_bekleniyor
    │ (all refs responded positively or unknowns resolved)
    ▼
referans_tamamlandı
    │ (koordinator / system advances)
    ▼
yk_ön_incelemede
    │ (all YK members voted, no reject)
    ▼
ön_onaylandı
    │ (koordinator adds 10 reputation contacts)
    ▼
itibar_taramasında
    │ (all contacts responded, no negative)
    ▼
itibar_temiz
    │ (koordinator adds to agenda)
    ▼
gündemde
    │ (YK final vote, all approve/abstain)
    ▼
kabul
    │ (admin records web publish consent)
    ▼
[web_publish_consent recorded]

─── RED PATHS ──────────────────────────────
referans_bekleniyor ──[any negative ref]──► referans_red ──► reddedildi
yk_ön_incelemede   ──[any YK reject]────► yk_red ──────► reddedildi
itibar_taramasında ──[negative + YK rej]► itibar_red ───► reddedildi
gündemde           ──[any YK reject]────► reddedildi
```

### 6.2 Profesyonel & Öğrenci State Flow

```
başvuru_alındı
    ▼
danışma_sürecinde
    │ (min 2 members consulted, all positive)
    ▼
gündemde
    │ (YK final vote)
    ▼
kabul / reddedildi
```

### 6.3 Onursal State Flow

```
öneri_alındı
    ▼
yk_ön_incelemede
    │ (all YK, no reject)
    ▼
ön_onaylandı
    ▼
yik_değerlendirmede
    │ (no YİK negative within deadline)
    ▼
gündemde
    │ (YK final vote)
    ▼
kabul / reddedildi
```

### 6.4 statemachine.go Contract

```go
// ValidateTransition returns error if transition is illegal
// for the given membership type and current status.
func ValidateTransition(
    membershipType MembershipType,
    currentStatus  ApplicationStatus,
    nextStatus     ApplicationStatus,
) error

// IsTerminated returns true if application is in any terminal RED state
func IsTerminated(status ApplicationStatus) bool

// AllowedTransitions returns valid next states for debugging/UI
func AllowedTransitions(
    membershipType MembershipType,
    currentStatus  ApplicationStatus,
) []ApplicationStatus
```

---

## 7. VETO MODEL — CRITICAL RULE

> **ONE JUSTIFIED RED = PERMANENT TERMINATION**

This applies at every stage, without exception:

| Stage | Who can RED |
|-------|-------------|
| Reference response | Any referee (negative response) |
| YK Ön İnceleme | Any single YK member |
| Consultation (Prof/Öğr) | Any single consulted member |
| İtibar Taraması | Any contact → reviewed by YK → YK can reject |
| YİK Değerlendirme | Any single YİK member |
| YK Nihai Oylama | Any single YK member |

### What happens on RED

```
1. Write vote/response record (immutable)
2. Call RedGuard.Terminate(applicationID, reason, actorID, role)
   a. Set application.status = reddedildi (or type-specific red state)
   b. Set application.rejection_reason (write-once — never overwrite)
   c. Set application.rejected_by_role (role only, not name)
   d. Write to logs (immutable)
   e. Trigger applicant rejection email (no reason disclosed)
3. Block ALL further transitions on this application
4. Return success to caller
```

### redguard.go Contract

```go
func (g *RedGuard) Terminate(
    applicationID string,
    reason        string,
    actorID       string,
    actorRole     string,
) error

func (g *RedGuard) IsTerminated(applicationID string) (bool, error)

func (g *RedGuard) AssertNotTerminated(applicationID string) error
// Returns ErrApplicationTerminated if already terminated
```

---

## 8. REFERENCE SYSTEM

> **Applies to: Asil & Akademik only**

### Flow

```
1. Applicant selects 3+ references from member autocomplete
2. System generates unique token per reference (see Token System)
3. Emails sent: reference_request.html
4. Reference opens tokenized link (no login required)
5. Reference submits response
```

### Response Logic

```
POSITIVE  → Record. Check if all references complete.
            If yes → advance to referans_tamamlandı

UNKNOWN   → Record. Email applicant: "Please add a replacement reference."
            New reference slot created (is_replacement=true, round++)
            New token generated and sent to new referee
            Original response preserved in logs

NEGATIVE  → RedGuard.Terminate() immediately
            No further processing
```

### Replacement Rules
- Each "unknown" response creates exactly one replacement slot
- Replacement refs also need 3 total positives (unknowns don't count)
- Applicant notified via `applicant_new_ref_needed.html`

---

## 9. CONSULTATION SYSTEM

> **Applies to: Profesyonel & Öğrenci only**

### Flow

```
1. Koordinator selects minimum 2 existing members to consult
2. System generates token per member, sends consultation_request.html
3. Member opens tokenized link, submits response
4. Any NEGATIVE → RedGuard.Terminate()
5. All POSITIVE → advance to gündemde
```

### Rules
- Minimum 2 consultations enforced at service layer
- Koordinator may add more than 2
- ALL consulted members must be positive (any negative = terminate)
- Tokens: 48h TTL, single-use

---

## 10. REPUTATION SYSTEM

> **Applies to: Asil & Akademik only**
> **Triggered: After ön_onaylandı**

### Flow

```
1. Koordinator (or YK) manually adds 10 contact email addresses
   (these are LinkedIn connections of the applicant — entered manually)
2. System generates token per contact, sends reputation_query.html
3. Contact opens tokenized link, answers:
   "Do you have any negative information about [candidate]?"
4. Response types:
   - clean    → Record
   - negative → Flag to YK for review
5. YK reviews flagged responses
6. If YK casts reject vote → RedGuard.Terminate()
7. If all clean (or YK approves despite flags) → itibar_temiz
```

### NO SCRAPING RULE
> The system MUST NOT attempt to scrape, access, or query LinkedIn.
> All contact emails are entered manually by koordinator or YK.
> This is non-negotiable.

---

## 11. VOTING SYSTEM

### Three Vote Stages

| Stage | Route Suffix | Eligible Voters | Applies To |
|-------|-------------|-----------------|------------|
| YK Ön İnceleme | `/yk-prelim` | yk | Asil, Akademik, Onursal |
| YİK Değerlendirme | `/yik` | yik | Onursal only |
| YK Nihai Oylama | `/yk-final` | yk | All types |

### Vote Types

```
approve  → No reason required
abstain  → No reason required
reject   → Reason REQUIRED (min 20 characters)
           Triggers RedGuard.Terminate() immediately
```

### YİK Special Rules
- YİK members submit ONLY negative opinions
- Silence (no vote) = no objection (positive by default)
- Any single YİK negative = TERMINATE

### Duplicate Vote Guard
```
Unique constraint: (application_id, voter_id, vote_stage)
Second vote attempt → 409 Conflict
```

### Vote Finalization Logic
```
yk-prelim: When all active YK members voted + no reject
           → advance to ön_onaylandı

yik:       When deadline passes with no negative
           → advance to gündemde

yk-final:  When all active YK members voted + no reject
           → advance to kabul
```

---

## 12. HONORARY MEMBERSHIP FLOW

### Proposer Eligibility
- Role must be: `asil_uye` OR `yik_uye`
- Proposed via logged-in session (not public form)

### Required Fields
- nominee_name (full name)
- nominee_linkedin (URL)
- proposal_reason (written justification, min 100 chars)

### State Flow
```
öneri_alındı
  → yk_ön_incelemede  (YK reviews proposal)
  → ön_onaylandı      (no YK reject)
  → yik_değerlendirmede (YİK negative-only model)
  → gündemde
  → kabul / reddedildi
```

---

## 13. TOKEN SYSTEM

> All external-facing links (references, consultations, reputation)
> use secure, expiring, single-use tokens.

### Token Generation

```go
// token.go
func GenerateToken(entityID string, secret string) (raw string, hashed string, expiresAt time.Time)

// raw     → sent in email URL (never stored)
// hashed  → stored in database (SHA-256)
// expiry  → time.Now().Add(48 * time.Hour)
```

### Token Validation (on every public endpoint)

```
1. Receive raw token from URL
2. Hash it: SHA-256(raw)
3. Look up hashed value in database
4. Check: token_expires > NOW()     → if not: return 410 Gone
5. Check: token_used = false        → if used: return 409 Conflict
6. Set token_used = true (atomic)
7. Process response
```

### Token URLs
```
Reference:    {APP_BASE_URL}/respond/reference/{raw_token}
Reputation:   {APP_BASE_URL}/respond/reputation/{raw_token}
Consultation: {APP_BASE_URL}/respond/consultation/{raw_token}
```

---

## 14. EMAIL SYSTEM

### Configuration
```
Dev:  MailDev (MAIL_HOST=localhost, MAIL_PORT=1025)
Prod: Any SMTP provider
```

### Templates (in notifications/templates/)

| File | Trigger | Recipient |
|------|---------|-----------|
| `reference_request.html` | Application submitted | Referee |
| `reference_reminder.html` | 24h before token expiry | Unanswered referee |
| `applicant_new_ref_needed.html` | Referee said "unknown" | Applicant |
| `reputation_query.html` | Koordinator adds rep contacts | Contact |
| `consultation_request.html` | Koordinator adds consultees | Member |
| `application_accepted.html` | Status → kabul | Applicant |
| `application_rejected.html` | Status → reddedildi | Applicant |
| `honorary_proposal_notify.html` | Proposal submitted | All YK members |

### Template Variables
Each template receives a typed Go struct. Example:
```go
type ReferenceRequestData struct {
    RefereeName    string
    ApplicantName  string
    ResponseURL    string
    ExpiresAt      string
    MembershipType string
}
```

### Rules
- Always send HTML + plain-text fallback
- Rejection emails MUST NOT include the rejection reason
- Retry: 3 attempts with exponential backoff (1s, 2s, 4s)
- All sent emails logged in `logs` table with action: `email.sent`

---

## 15. DATABASE SCHEMA

### Table: `users`
```sql
id              CHAR(36) PRIMARY KEY   -- UUID
email           VARCHAR(255) UNIQUE NOT NULL
password_hash   VARCHAR(255) NOT NULL
full_name       VARCHAR(255) NOT NULL
role            ENUM('admin','yk','yik','koordinator','asil_uye','yik_uye')
is_active       BOOLEAN DEFAULT TRUE
created_at      DATETIME
updated_at      DATETIME
```

### Table: `applications`
```sql
id                  CHAR(36) PRIMARY KEY
applicant_name      VARCHAR(255) NOT NULL
applicant_email     VARCHAR(255) NOT NULL
linkedin_url        VARCHAR(500) UNIQUE
photo_url           VARCHAR(500)
membership_type     ENUM('asil','akademik','profesyonel','ogrenci','onursal')
status              ENUM(-- all states listed below --)
proposed_by_user_id CHAR(36)               -- FK users.id, Onursal only
proposal_reason     TEXT                   -- Onursal only
rejection_reason    TEXT                   -- WRITE-ONCE, immutable after set
rejected_by_role    VARCHAR(50)            -- role only, never name
web_publish_consent BOOLEAN DEFAULT NULL
is_published        BOOLEAN DEFAULT FALSE
previous_app_id     CHAR(36)               -- FK self, repeat applicant ref
created_at          DATETIME
updated_at          DATETIME
```

#### All Status Values (application.status enum)
```
başvuru_alındı, referans_bekleniyor, referans_tamamlandı,
referans_red, yk_ön_incelemede, yk_red, ön_onaylandı,
itibar_taramasında, itibar_red, itibar_temiz,
danışma_sürecinde, danışma_red,
gündemde, kabul, reddedildi,
öneri_alındı, yik_değerlendirmede, yik_red
```

### Table: `references`
```sql
id              CHAR(36) PRIMARY KEY
application_id  CHAR(36) NOT NULL    -- FK applications.id
referee_user_id CHAR(36) NOT NULL    -- FK users.id (system member)
ref_name        VARCHAR(255)
ref_email       VARCHAR(255)
token_hash      VARCHAR(64) UNIQUE   -- SHA-256 hash of raw token
token_expires   DATETIME NOT NULL
token_used      BOOLEAN DEFAULT FALSE
is_replacement  BOOLEAN DEFAULT FALSE
round           INT DEFAULT 1
created_at      DATETIME
```

### Table: `reference_responses`
```sql
id              CHAR(36) PRIMARY KEY
reference_id    CHAR(36) NOT NULL    -- FK references.id
response_type   ENUM('positive','unknown','negative')
reason          TEXT                 -- required if negative
ip_address      VARCHAR(45)
user_agent      VARCHAR(500)
responded_at    DATETIME NOT NULL
```

### Table: `consultations`
```sql
id                CHAR(36) PRIMARY KEY
application_id    CHAR(36) NOT NULL  -- FK applications.id
consulted_user_id CHAR(36) NOT NULL  -- FK users.id
token_hash        VARCHAR(64) UNIQUE
token_expires     DATETIME NOT NULL
token_used        BOOLEAN DEFAULT FALSE
response_type     ENUM('positive','negative')
reason            TEXT
responded_at      DATETIME
created_at        DATETIME
```

### Table: `reputation_contacts`
```sql
id              CHAR(36) PRIMARY KEY
application_id  CHAR(36) NOT NULL    -- FK applications.id
contact_name    VARCHAR(255)
contact_email   VARCHAR(255) NOT NULL
added_by        CHAR(36) NOT NULL    -- FK users.id
token_hash      VARCHAR(64) UNIQUE
token_expires   DATETIME NOT NULL
token_used      BOOLEAN DEFAULT FALSE
response_type   ENUM('clean','negative')
reason          TEXT
responded_at    DATETIME
created_at      DATETIME
```

### Table: `votes`
```sql
id              CHAR(36) PRIMARY KEY
application_id  CHAR(36) NOT NULL    -- FK applications.id
voter_id        CHAR(36) NOT NULL    -- FK users.id
voter_role      ENUM('yk','yik')
vote_stage      ENUM('yk_prelim','yik','yk_final')
vote_type       ENUM('approve','abstain','reject')
reason          TEXT                 -- required if reject
is_veto         BOOLEAN DEFAULT FALSE
created_at      DATETIME             -- immutable

UNIQUE KEY unique_vote (application_id, voter_id, vote_stage)
```

### Table: `web_publish_consents`
```sql
id              CHAR(36) PRIMARY KEY
application_id  CHAR(36) NOT NULL UNIQUE
consented       BOOLEAN NOT NULL
recorded_by     CHAR(36) NOT NULL    -- FK users.id (admin)
created_at      DATETIME
```

### Table: `logs`
```sql
id              CHAR(36) PRIMARY KEY
actor_id        CHAR(36)             -- FK users.id (nullable = system)
actor_role      VARCHAR(50)
action          VARCHAR(100) NOT NULL
entity_type     VARCHAR(50) NOT NULL
entity_id       CHAR(36) NOT NULL
metadata        JSON
ip_address      VARCHAR(45)
created_at      DATETIME NOT NULL

-- NO UPDATE, NO DELETE permissions on this table
```

---

## 16. API ENDPOINT MAP

### Auth
```
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
```

### Applications
```
POST   /api/v1/applications                         (public)
GET    /api/v1/applications                         (admin, yk, koordinator)
GET    /api/v1/applications/:id                     (admin, yk, yik)
GET    /api/v1/applications/:id/timeline            (admin, yk)
GET    /api/v1/applications/:id/red-history         (yk, admin ONLY)
```

### References
```
GET    /api/v1/ref/respond/:token                   (public)
POST   /api/v1/ref/respond/:token                   (public)
POST   /api/v1/applications/:id/references/resend/:refId  (koordinator)
```

### Consultations
```
POST   /api/v1/applications/:id/consultations       (koordinator)
GET    /api/v1/consult/respond/:token               (public)
POST   /api/v1/consult/respond/:token               (public)
```

### Reputation
```
GET    /api/v1/applications/:id/reputation          (yk, koordinator)
POST   /api/v1/applications/:id/reputation/contacts (yk, koordinator)
GET    /api/v1/reputation/respond/:token            (public)
POST   /api/v1/reputation/respond/:token            (public)
```

### Voting
```
POST   /api/v1/applications/:id/votes/yk-prelim     (yk)
POST   /api/v1/applications/:id/votes/yik           (yik)
POST   /api/v1/applications/:id/votes/yk-final      (yk)
GET    /api/v1/applications/:id/votes               (yk, admin)
```

### Honorary
```
POST   /api/v1/honorary/propose                     (asil_uye, yik_uye)
GET    /api/v1/honorary                             (admin, yk)
```

### Web Publish
```
POST   /api/v1/applications/:id/publish-consent     (admin)
GET    /api/v1/members                              (public)
```

### Logs
```
GET    /api/v1/logs                                 (admin, yk)
GET    /api/v1/logs/:entity_type/:entity_id         (admin, yk)
```

---

## 17. SECURITY RULES

### JWT Configuration
```
Access token TTL  : 15 minutes
Refresh token TTL : 7 days
Algorithm         : HS256
Refresh rotation  : YES (new refresh token on every refresh)
```

### Password Policy
```
Hashing  : bcrypt, cost factor 12
Min length: 8 characters (enforced at registration)
```

### Token Security
```
Raw token  : UUID v4 — sent in URL, NEVER stored
Stored     : SHA-256(raw_token) in database
Expiry     : 48 hours from creation
Single-use : token_used flag set atomically on first use
```

### RED Decision Privacy
```
rejected_by_role : stored in DB, returned only to yk + admin roles
rejected_by_name : NEVER stored, NEVER returned via API
Vote reasons     : visible to yk + admin only
RED history      : endpoint gated to yk + admin via RBAC
```

### Repeat Applicant Detection
```
On new application submission:
  Query: SELECT * FROM logs
         WHERE entity_type = 'application'
           AND JSON_EXTRACT(metadata, '$.applicant_email') = :email
           AND action IN ('application.terminated')
  If found:
    - Set flag on application record
    - Return flag in application detail response (for YK panel only)
    - Link previous application IDs in response
```

---

## 18. AUDIT & IMMUTABILITY RULES

### Every action MUST be logged. No exceptions.

#### Required Log Actions
```
auth.login              auth.logout
application.created     application.status_changed
ref.sent               ref.responded             ref.expired
rep.contact_added       rep.responded
consult.sent           consult.responded
vote.cast              vote.veto
email.sent             email.failed
publish.consent_recorded
application.terminated
```

### Immutability Enforcement

**Database level:**
```sql
-- Trigger: prevent update of rejection_reason once set
CREATE TRIGGER prevent_rejection_reason_update
BEFORE UPDATE ON applications
FOR EACH ROW
BEGIN
    IF OLD.rejection_reason IS NOT NULL
       AND NEW.rejection_reason != OLD.rejection_reason THEN
        SIGNAL SQLSTATE '45000'
        SET MESSAGE_TEXT = 'rejection_reason is immutable';
    END IF;
END;
```

**Service level (redguard.go):**
```go
// Called at start of every service method that modifies an application
err := redguard.AssertNotTerminated(applicationID)
if err != nil {
    return ErrApplicationTerminated
}
```

**Logs table:**
```
- No UPDATE statements ever executed against logs
- No DELETE statements ever executed against logs
- Application-level: logs.service.go only has Create() — no Update/Delete
```

---

## 19. FRONTEND RULES

### Component Architecture
```
Page (app/...)          → Fetches data, passes to components
Feature Component       → Business UI logic
UI Component (libs/ui)  → Pure presentational, no data fetching
```

### Form Validation
- All forms use React Hook Form + Zod
- Zod schemas live in `libs/validators/src/`
- Same schema used for both client validation and API type checking

### Role-Gated UI
```typescript
// Pattern to use throughout frontend
const { role } = useAuthStore()

// Render components conditionally
{hasRole(role, ['yk', 'admin']) && <RedHistoryBanner />}
{hasRole(role, ['yk']) && <VotingPanel />}
```

### Error Handling
```
401 → Redirect to /login, clear auth store
403 → Show AccessDenied component (no redirect)
410 → Show TokenExpired page (token links)
409 → Show AlreadySubmitted page (token links)
422 → Show inline validation errors from API
500 → Show generic error with request ID
```

### API Response Shape (from backend)
```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "message": "OK"
}
```
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "LinkedIn URL already exists",
    "fields": { "linkedin_url": "already taken" }
  }
}
```

---

## 20. CODING STANDARDS

### Go (Backend)
```
- All errors wrapped with context: fmt.Errorf("service.CreateApplication: %w", err)
- No naked returns
- Context passed to all DB operations
- Services must not import net/http or Fiber — only domain types
- Handlers must not contain business logic
- All exported functions must have Go doc comments
- Use uuid.New().String() for all ID generation
```

### TypeScript (Frontend)
```typescript
// Use strict TypeScript — no `any` types
// All API response types defined in libs/shared-types
// All async operations wrapped in try/catch
// No inline styles — Tailwind only
// Component files: PascalCase.tsx
// Hook files: useCamelCase.ts
// Utility files: camelCase.ts
```

### Naming Conventions
```
DB tables     : snake_case, plural
Go structs    : PascalCase
Go interfaces : IPascalCase or PascalCase + "er" suffix
API routes    : kebab-case
React components: PascalCase
TS types/interfaces: PascalCase
```

---

## 21. IMPLEMENTATION ORDER

> Follow this order strictly. Do not skip phases.
> Each phase builds on the previous.

```
Phase 1  │ Infrastructure
         │ docker-compose.yml, turbo.json, .env.example
         │ Go module init, Next.js init, shared libs scaffold
         │
Phase 2  │ Database
         │ All 9 migration files
         │ All GORM models with correct tags
         │ DB trigger for rejection_reason immutability
         │
Phase 3  │ Auth
         │ Login, refresh, logout endpoints
         │ JWT middleware, RBAC middleware
         │ Bcrypt + token utilities
         │
Phase 4  │ Application Core
         │ Application CRUD endpoints
         │ statemachine.go (all 3 type flows)
         │ redguard.go
         │ Repeat applicant detection
         │
Phase 5  │ Reference System
         │ Token generation (token.go)
         │ Reference creation + tokenized public endpoints
         │ Replacement flow
         │
Phase 6  │ Email System
         │ Mailer client (MailDev config)
         │ All 8 email templates
         │ 24h reminder cron
         │
Phase 7  │ Consultation System
         │ (Profesyonel/Öğrenci)
         │
Phase 8  │ Reputation System
         │ (Asil/Akademik)
         │
Phase 9  │ Voting Module
         │ All 3 vote stages
         │ Veto logic + RedGuard integration
         │
Phase 10 │ Honorary Flow
         │ Proposal endpoint + YK/YİK/YK voting
         │
Phase 11 │ Web Publish
         │ Consent recording + public member list
         │
Phase 12 │ Admin UI
         │ Application list + detail + timeline
         │ Voting panels (all 3 stages)
         │ Reference/consultation/reputation sub-pages
         │ Audit log viewer
         │ Public token response pages
         │
Phase 13 │ Audit & Security Hardening
         │ Audit middleware on all routes
         │ All log actions wired up
         │ Security headers (helmet)
         │ Rate limiting on public endpoints
```

---

## 22. FORBIDDEN ACTIONS

> The AI agent MUST NEVER do any of the following.
> These are hard constraints, not suggestions.

```
❌ Scrape or access LinkedIn programmatically
❌ Allow manual status override via API (bypass state machine)
❌ Store raw tokens in database (always store SHA-256 hash)
❌ Return rejected_by_name in any API response
❌ Allow UPDATE on logs table
❌ Allow UPDATE on rejection_reason after it is set
❌ Allow DELETE on any vote record
❌ Allow a terminated application to advance in state
❌ Skip RedGuard check in any service method
❌ Use `any` type in TypeScript code
❌ Put business logic in handlers (HTTP layer)
❌ Put DB queries in services (use repository layer)
❌ Send rejection reason to applicant in email
❌ Allow non-YK roles to access /red-history endpoint
❌ Allow YK member to vote in YİK stage or vice versa
❌ Allow second vote from same person in same stage
❌ Generate tokens without expiry
❌ Allow reuse of an already-used token
```

---

## 📎 QUICK REFERENCE CARD

```
Veto rule        → 1 RED anywhere = permanent termination
Token TTL        → 48 hours, single-use, SHA-256 stored
JWT access       → 15 min | JWT refresh → 7 days
Bcrypt cost      → 12
Min references   → 3 (Asil/Akademik)
Min consultations→ 2 (Profesyonel/Öğrenci)
Reputation contacts → exactly 10 (Asil/Akademik)
Log immutability → append-only, no update/delete
RED immutability → rejection_reason write-once (DB trigger)
YİK model        → negative-only; silence = approval
LinkedIn scraping→ FORBIDDEN
Status override  → FORBIDDEN
```

---

*This document is authoritative. In case of any conflict between
this file and any other source, this file takes precedence.*