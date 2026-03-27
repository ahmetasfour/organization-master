# 🤖 COPILOT AGENT PROMPTS
## Membership Management System — Step-by-Step Build Prompts
### Usage: Paste one prompt at a time. Complete each phase before moving to the next.

---

## ⚙️ HOW TO USE THESE PROMPTS

1. Open GitHub Copilot Chat (Agent mode) or Copilot Workspace
2. Start EVERY session by pasting the **SESSION HEADER** below
3. Then paste the prompt for the current phase
4. Wait for Copilot to complete all tasks in that phase
5. Review, test, commit
6. Move to the next phase

---

## 🔰 SESSION HEADER
> Paste this at the START of every new Copilot session before any phase prompt.

```
You are a senior full-stack engineer building a production-grade
Membership Management System. The complete specification is in
INSTRUCTIONS.md at the root of this repository.

Before doing anything:
1. Read INSTRUCTIONS.md completely
2. Understand all state machines, veto rules, token system,
   RBAC matrix, and forbidden actions
3. Follow the exact folder structure defined in Section 3
4. Apply all coding standards from Section 20
5. Never violate any rule in Section 22 (Forbidden Actions)

All decisions must be consistent with INSTRUCTIONS.md.
When in doubt, re-read the relevant section before proceeding.
```

---

---

# PHASE 1 — Infrastructure & Monorepo Setup

```
Read INSTRUCTIONS.md Section 3 (Monorepo Structure) and Section 2 (Tech Stack).

Set up the complete monorepo infrastructure. Do all of the following:

TURBOREPO SETUP:
- Initialize Turborepo at the root
- Create turbo.json with pipelines: build, dev, lint, test
- Create root package.json with workspaces:
    ["apps/*", "libs/*"]
- Add .gitignore covering: node_modules, dist, .next, tmp, .env, *.local

DOCKER:
- Create docker-compose.yml with these services:
    mysql:
      image: mysql:8.0
      environment: MYSQL_ROOT_PASSWORD, MYSQL_DATABASE=membership_db
      ports: 3306:3306
      volumes: mysql_data:/var/lib/mysql

    maildev:
      image: maildev/maildev
      ports: 1080:1080 (web UI), 1025:1025 (SMTP)

    api:
      build: ./apps/api
      depends_on: mysql, maildev
      env_file: .env
      ports: 8080:8080

    admin:
      build: ./apps/admin
      ports: 3000:3000

ENV:
- Create .env.example with ALL these variables:
    # Database
    DB_HOST=localhost
    DB_PORT=3306
    DB_NAME=membership_db
    DB_USER=root
    DB_PASS=secret

    # JWT
    JWT_SECRET=change_me_in_production
    JWT_REFRESH_SECRET=change_me_refresh
    JWT_ACCESS_TTL=15m
    JWT_REFRESH_TTL=168h

    # App
    APP_BASE_URL=http://localhost:3000
    APP_PORT=8080
    APP_ENV=development

    # Email
    MAIL_HOST=localhost
    MAIL_PORT=1025
    MAIL_FROM=noreply@membership.local
    MAIL_FROM_NAME=Membership System

GO BACKEND (apps/api):
- Run: go mod init membership-system/api
- Create go.mod with these dependencies:
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/gofiber/jwt/v3 v3.3.10
    github.com/golang-jwt/jwt/v5 v5.2.0
    gorm.io/gorm v1.25.7
    gorm.io/driver/mysql v1.5.4
    github.com/google/uuid v1.6.0
    github.com/spf13/viper v1.18.2
    golang.org/x/crypto v0.19.0
    gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
    github.com/go-playground/validator/v10 v10.18.0
- Create apps/api/Dockerfile (multi-stage: builder + runner)
- Create apps/api/cmd/main.go with:
    - Viper config load
    - GORM MySQL connection
    - Fiber app init
    - Health check route: GET /api/v1/health → { "status": "ok" }
    - Graceful shutdown
- Create apps/api/config/config.go (all env vars as typed struct)
- Create apps/api/config/database.go (GORM DSN builder + AutoMigrate call)

NEXT.JS FRONTEND (apps/admin):
- Initialize Next.js 14 with App Router, TypeScript, Tailwind CSS
- Install dependencies:
    @tanstack/react-query@5
    @tanstack/react-table@8
    react-hook-form
    @hookform/resolvers
    zod
    axios
    zustand
    lucide-react
    date-fns
- Install and configure shadcn/ui (init with slate theme, CSS variables)
- Create apps/admin/Dockerfile
- Create apps/admin/src/lib/api/client.ts:
    Axios instance with:
    - baseURL from NEXT_PUBLIC_API_URL env var
    - Authorization header injection from auth store
    - 401 interceptor → clear store + redirect to /login
    - Unified error response handling

SHARED LIBRARIES:
- Create libs/shared-types/package.json and tsconfig.json
- Create libs/shared-types/src/index.ts exporting:
    - ApplicationStatus (all status values from INSTRUCTIONS.md Section 15)
    - MembershipType enum
    - UserRole enum
    - ApiResponse<T> generic interface
    - All entity interfaces: Application, User, Vote, Reference, Log
- Create libs/validators/package.json and tsconfig.json
- Create libs/validators/src/index.ts (empty, will be filled in later phases)
- Create libs/ui/package.json (re-exports shadcn/ui components)

VERIFICATION:
After completing setup, confirm:
- `docker compose up` starts all 4 services without error
- GET http://localhost:8080/api/v1/health returns { "status": "ok" }
- http://localhost:1080 shows MailDev UI
- http://localhost:3000 shows Next.js default page
```

---

---

# PHASE 2 — Database Migrations & GORM Models

```
Read INSTRUCTIONS.md Section 15 (Database Schema) and Section 18 (Audit & Immutability Rules).

Create all database migrations and GORM models.

MIGRATION FILES (apps/api/migrations/):
Create these SQL files in order. Each file must be idempotent (IF NOT EXISTS).

001_users.sql:
  Create users table exactly as specified in INSTRUCTIONS.md Section 15.
  Include: id (CHAR36 PK), email (UNIQUE), password_hash, full_name,
  role (ENUM with all 6 roles), is_active, created_at, updated_at.

002_applications.sql:
  Create applications table with ALL status enum values listed in
  INSTRUCTIONS.md Section 15 under "All Status Values".
  Include: proposed_by_user_id, proposal_reason (Onursal fields),
  rejection_reason (with WRITE-ONCE constraint),
  rejected_by_role, web_publish_consent, is_published,
  previous_app_id (self-referential FK).

003_references.sql:
  Create references table.
  Note: column is token_hash (not token) — stores SHA-256 hash only.
  Include: is_replacement BOOLEAN, round INT DEFAULT 1.

004_reference_responses.sql:
  Create reference_responses table with response_type ENUM
  ('positive','unknown','negative').

005_consultations.sql:
  Create consultations table for Profesyonel/Öğrenci flow.

006_reputation_contacts.sql:
  Create reputation_contacts table for Asil/Akademik flow.
  response_type ENUM('clean','negative').

007_votes.sql:
  Create votes table with:
  - vote_stage ENUM('yk_prelim','yik','yk_final')
  - vote_type ENUM('approve','abstain','reject')
  - is_veto BOOLEAN DEFAULT FALSE
  - UNIQUE KEY on (application_id, voter_id, vote_stage)
  - created_at only (no updated_at — immutable)

008_web_publish_consents.sql:
  Create web_publish_consents table.
  UNIQUE on application_id.

009_logs.sql:
  Create logs table.
  Include metadata JSON column.
  Add comment: "This table is append-only. No UPDATE or DELETE."

IMMUTABILITY TRIGGER (in 002_applications.sql or separate 010_triggers.sql):
  Create MySQL trigger:
  BEFORE UPDATE ON applications — if OLD.rejection_reason IS NOT NULL
  and NEW.rejection_reason differs, SIGNAL error with message
  'rejection_reason is immutable once set'.

GORM MODELS:
For each feature in apps/api/internal/features/, create models.go:

features/auth/models.go:
  User struct with GORM tags matching 001_users.sql exactly.
  Include BeforeCreate hook to validate UUID.

features/applications/models.go:
  Application struct.
  Define ApplicationStatus type with all status constants.
  Define MembershipType type with all 5 type constants.
  Include GORM associations: HasMany References, HasMany Votes,
  HasMany Logs, BelongsTo User (proposed_by).

features/references/models.go:
  Reference struct + ReferenceResponse struct.
  ResponseType: positive, unknown, negative.

features/consultations/models.go:
  Consultation struct with all fields.

features/reputation/models.go:
  ReputationContact struct.

features/voting/models.go:
  Vote struct. VoteStage + VoteType enums.

features/logs/models.go:
  Log struct. Metadata field as JSON (use datatypes.JSON from gorm.io/datatypes).

MIGRATION RUNNER:
In apps/api/config/database.go:
  - Run AutoMigrate for all models on startup
  - Execute raw SQL trigger creation (idempotent — DROP TRIGGER IF EXISTS first)
  - Log each migration step

VERIFICATION:
- docker compose up runs migrations without error
- All 9 tables exist in MySQL
- UNIQUE constraints verified
- Trigger exists: SHOW TRIGGERS from membership_db
- Attempt to update rejection_reason twice → expect SQL error
```

---

---

# PHASE 3 — Authentication & RBAC

```
Read INSTRUCTIONS.md Section 17 (Security Rules) and Section 4 (Roles & Permissions).

Implement complete authentication system and role-based access control.

SHARED UTILITIES (apps/api/internal/shared/):

token.go:
  Implement GenerateToken(entityID, secret string) function:
  - Generate UUID v4 as raw token
  - Compute SHA-256 hash of raw token
  - Set expiry: time.Now().Add(48 * time.Hour)
  - Return: rawToken string, hashedToken string, expiresAt time.Time
  - Raw token goes in email URL — NEVER stored in DB
  - Hashed token stored in DB

response.go:
  Implement unified API response helpers:
  - Success(c *fiber.Ctx, data interface{}) error
  - Error(c *fiber.Ctx, status int, code string, message string) error
  - ValidationError(c *fiber.Ctx, fields map[string]string) error
  Response shape:
  { "success": bool, "data": any, "error": { "code", "message", "fields" } }

errors.go:
  Define sentinel errors:
  - ErrApplicationTerminated
  - ErrTokenExpired
  - ErrTokenUsed
  - ErrInvalidCredentials
  - ErrForbidden
  - ErrNotFound
  - ErrDuplicateVote

AUTH FEATURE (apps/api/internal/features/auth/):

dtos.go:
  LoginRequest { Email, Password } with validation tags
  LoginResponse { AccessToken, RefreshToken, User { ID, Name, Email, Role } }
  RefreshRequest { RefreshToken }

repository.go:
  FindByEmail(ctx, email) (*User, error)
  FindByID(ctx, id) (*User, error)
  Create(ctx, user) error

service.go:
  Login(ctx, email, password) (*LoginResponse, error):
    1. FindByEmail
    2. bcrypt.CompareHashAndPassword
    3. Generate access JWT (15min, includes: sub, role, email, iat, exp)
    4. Generate refresh JWT (7d)
    5. Log action: auth.login
    6. Return tokens + user info

  Refresh(ctx, refreshToken) (*LoginResponse, error):
    1. Validate refresh JWT
    2. Find user by sub claim
    3. Generate new access + refresh tokens (rotation)
    4. Return new token pair

  Logout(ctx, userID) error:
    1. Log action: auth.logout
    (stateless JWT — just log the event)

handler.go:
  POST /api/v1/auth/login
  POST /api/v1/auth/refresh
  POST /api/v1/auth/logout

MIDDLEWARE (apps/api/internal/middleware/):

auth.go (JWT validation middleware):
  - Extract Bearer token from Authorization header
  - Validate using JWT_SECRET
  - Check expiry
  - Store parsed claims in fiber.Ctx locals: "userID", "userRole", "userEmail"
  - Return 401 with code "UNAUTHORIZED" on any failure

rbac.go (Role-based access control middleware):
  Implement: RequireRole(roles ...string) fiber.Handler
  - Read "userRole" from ctx locals (set by auth middleware)
  - If role not in allowed list → 403 with code "FORBIDDEN"

  Create named middleware shortcuts:
  AdminOnly()       → RequireRole("admin")
  YKOnly()          → RequireRole("yk")
  YIKOnly()         → RequireRole("yik")
  KoordinatorOnly() → RequireRole("koordinator")
  YKOrAdmin()       → RequireRole("yk", "admin")
  YKOrKoordinator() → RequireRole("yk", "koordinator", "admin")
  ProposerOnly()    → RequireRole("asil_uye", "yik_uye")

audit.go (Auto-audit middleware):
  Implement middleware that fires AFTER response for all non-GET routes:
  - Captures: actor_id, actor_role, method, path, status_code, ip
  - Writes to logs table with action: "http.{METHOD}.{path_template}"
  - Must not block the response

cors.go:
  Configure Fiber CORS for localhost:3000 in development

ROUTER (apps/api/internal/router/router.go):
  Wire all auth routes:
  POST /api/v1/auth/login    → auth handler (no auth middleware)
  POST /api/v1/auth/refresh  → auth handler (no auth middleware)
  POST /api/v1/auth/logout   → auth handler (JWT required)

SEED DATA:
  Create apps/api/cmd/seed.go that creates default users when run:
  - admin@system.local / Admin123! / role: admin
  - koordinator@system.local / Koord123! / role: koordinator
  - yk1@system.local / YK123! / role: yk
  - yk2@system.local / YK123! / role: yk
  - yik1@system.local / YIK123! / role: yik
  - asil1@system.local / Asil123! / role: asil_uye

FRONTEND AUTH (apps/admin/src/):
  lib/store/auth.store.ts:
    Zustand store with: user, accessToken, refreshToken,
    login(), logout(), setTokens()
    Persist to localStorage (accessToken only)

  lib/api/auth.ts:
    loginApi(email, password) → calls POST /auth/login
    refreshApi(token) → calls POST /auth/refresh
    logoutApi() → calls POST /auth/logout

  app/(auth)/login/page.tsx:
    Login form using React Hook Form + Zod
    Schema: { email: z.string().email(), password: z.string().min(8) }
    On success: store tokens, redirect to /applications
    Show error message on invalid credentials

VERIFICATION:
- POST /api/v1/auth/login with valid credentials → returns token pair
- POST /api/v1/auth/login with wrong password → 401
- GET /api/v1/applications without token → 401
- GET /api/v1/applications with yk token but admin-only route → 403
- Login page works in browser
- Token stored in Zustand + localStorage
```

---

---

# PHASE 4 — Application Module & State Machine

```
Read INSTRUCTIONS.md Section 6 (State Machines), Section 7 (Veto Model),
Section 5 (Membership Types), and Section 4 (RBAC Matrix).

Build the core application module with strict state machine enforcement.

STATE MACHINE (apps/api/internal/features/applications/statemachine.go):

  Define all allowed transitions as a map:
  map[MembershipType]map[ApplicationStatus][]ApplicationStatus

  Asil & Akademik allowed transitions:
    başvuru_alındı         → [referans_bekleniyor]
    referans_bekleniyor    → [referans_tamamlandı, referans_red]
    referans_tamamlandı    → [yk_ön_incelemede]
    yk_ön_incelemede       → [ön_onaylandı, yk_red]
    ön_onaylandı           → [itibar_taramasında]
    itibar_taramasında     → [itibar_temiz, itibar_red]
    itibar_temiz           → [gündemde]
    gündemde               → [kabul, reddedildi]
    (all RED states)       → [] (no further transitions)

  Profesyonel & Öğrenci:
    başvuru_alındı    → [danışma_sürecinde]
    danışma_sürecinde → [gündemde, danışma_red]
    gündemde          → [kabul, reddedildi]

  Onursal:
    öneri_alındı           → [yk_ön_incelemede]
    yk_ön_incelemede       → [ön_onaylandı, yk_red]
    ön_onaylandı           → [yik_değerlendirmede]
    yik_değerlendirmede    → [gündemde, yik_red]
    gündemde               → [kabul, reddedildi]

  Implement functions:
  - ValidateTransition(membershipType, currentStatus, nextStatus) error
  - IsTerminated(status) bool → true for all *_red and reddedildi states
  - GetInitialStatus(membershipType) ApplicationStatus
  - AllowedTransitions(membershipType, currentStatus) []ApplicationStatus

RED GUARD (apps/api/internal/shared/redguard.go):

  RedGuard struct with DB dependency.

  Terminate(ctx, applicationID, reason, actorID, actorRole string) error:
    Within a single DB transaction:
    1. Load application — verify not already terminated
    2. Set status = reddedildi
    3. Set rejection_reason (only if currently NULL — never overwrite)
    4. Set rejected_by_role = actorRole
    5. Write to logs: action="application.terminated",
       metadata: { reason, actorRole, previousStatus }
    6. Trigger notification: send rejection email to applicant
       (email must NOT include the reason)
    7. Commit transaction

  AssertNotTerminated(ctx, applicationID) error:
    Load application status.
    If IsTerminated(status) → return ErrApplicationTerminated

  IsTerminated(ctx, applicationID) (bool, error)

APPLICATION FEATURE (apps/api/internal/features/applications/):

dtos.go:
  CreateApplicationRequest:
    ApplicantName    string (required, min 2, max 255)
    ApplicantEmail   string (required, email format)
    LinkedInURL      string (required for asil/akademik, valid URL)
    PhotoURL         string (optional)
    MembershipType   string (required, one of 5 types)
    References       []ReferenceInput (required for asil/akademik, min 3)
    ProposalReason   string (required for onursal, min 100 chars)

  ReferenceInput:
    UserID string (required — must be existing system user)

  ApplicationListResponse, ApplicationDetailResponse
  TimelineEntry { Status, ChangedAt, ChangedBy, Notes }

repository.go:
  Create(ctx, application) error
  FindByID(ctx, id) (*Application, error)
  FindAll(ctx, filters) ([]*Application, int64, error)
  UpdateStatus(ctx, id, status) error
  FindByApplicantEmail(ctx, email) ([]*Application, error)
  GetTimeline(ctx, id) ([]TimelineEntry, error)

service.go:
  Submit(ctx, req, actorID) (*Application, error):
    1. Validate membership type rules (refs required for asil/akademik)
    2. Check LinkedIn URL uniqueness
    3. REPEAT APPLICANT CHECK:
       Query logs WHERE action='application.terminated'
       AND metadata->applicant_email = req.email
       If found: set previous_app_id, add flag to response
    4. Create application with initial status from GetInitialStatus()
    5. For asil/akademik: create reference records (handled by ref service)
    6. Log: action="application.created"
    7. Return application

  GetByID(ctx, id, requestorRole) (*Application, error):
    Return application detail.
    If requestorRole is not yk/admin: exclude rejection_reason, rejected_by_role

  ListAll(ctx, filters, requestorRole) ([]*Application, int64, error)

  GetTimeline(ctx, id) ([]TimelineEntry, error)

  GetRedHistory(ctx, id) ([]Application, error):
    RBAC: only yk/admin may call this.
    Find previous applications by same email.
    Return with full rejection details.

handler.go:
  POST   /api/v1/applications         → public
  GET    /api/v1/applications         → YKOrKoordinator middleware
  GET    /api/v1/applications/:id     → auth required
  GET    /api/v1/applications/:id/timeline   → YKOrAdmin
  GET    /api/v1/applications/:id/red-history → YKOrAdmin only

REPEAT APPLICANT DETECTION:
  In Submit service method, after saving application:
  Query: SELECT metadata FROM logs
         WHERE entity_type = 'application'
         AND action = 'application.terminated'
         AND JSON_EXTRACT(metadata, '$.applicant_email') = ?
  If any rows found:
    Set application.previous_app_id = most recent terminated app ID
    Add "repeat_applicant": true to response metadata for YK

FRONTEND (apps/admin/src/):

  lib/api/applications.ts:
    getApplications(filters) → GET /api/v1/applications
    getApplication(id) → GET /api/v1/applications/:id
    getTimeline(id) → GET /api/v1/applications/:id/timeline
    getRedHistory(id) → GET /api/v1/applications/:id/red-history

  lib/hooks/useApplications.ts:
    useApplications(filters) → React Query, paginated
    useApplication(id) → React Query
    useTimeline(id) → React Query

  components/applications/StatusBadge.tsx:
    Color-coded badge for each status value.
    Red for all *_red and reddedildi states.
    Green for kabul. Yellow for in-progress states.

  components/applications/StatusTimeline.tsx:
    Visual step tracker.
    Renders correct steps based on membership_type.
    Shows completed/current/pending/terminated states.

  components/applications/RedHistoryBanner.tsx:
    Displayed only when repeat_applicant = true.
    Only rendered if user role is yk or admin.
    Shows: "Bu başvuran daha önce reddedilmiştir." + link to history.

  app/(dashboard)/applications/page.tsx:
    TanStack Table with:
    - Columns: Name, Type (badge), Status (StatusBadge), Created, Actions
    - Server-side pagination
    - Filters: membership_type (select), status (select), search (text)
    - URL-synced filter state (useSearchParams)
    - Link to detail page per row

  app/(dashboard)/applications/[id]/page.tsx:
    Application detail with:
    - Applicant info card
    - StatusTimeline component
    - RedHistoryBanner (conditional)
    - Tabs: References | Voting | Reputation | Consultation
      (tabs shown/hidden based on membership_type)

VERIFICATION:
- POST /api/v1/applications creates application with correct initial status
- ValidateTransition returns error for invalid transitions
- IsTerminated returns true for referans_red, yk_red, reddedildi, etc.
- RepeatApplicant flag appears when same email submits again
- Application list loads in admin UI with filters
- StatusBadge shows correct colors
- StatusTimeline shows correct steps per type
```

---

---

# PHASE 5 — Reference System & Token Engine

```
Read INSTRUCTIONS.md Section 8 (Reference System), Section 13 (Token System),
and Section 14 (Email System).

Build the complete tokenized reference system.

TOKEN ENGINE (apps/api/internal/shared/token.go):
  Already scaffolded in Phase 3. Verify it implements:
  - GenerateToken(entityID, secret) (raw, hashed, expiresAt)
  - ValidateAndConsumeToken(ctx, db, tokenHash, now) error:
      Load record by hashed token
      If not found → ErrTokenNotFound
      If token_expires < now → ErrTokenExpired (410)
      If token_used = true → ErrTokenUsed (409)
      Atomically set token_used = true
      Return nil (success)

REFERENCE FEATURE (apps/api/internal/features/references/):

dtos.go:
  ReferenceResponseRequest:
    ResponseType string (required: positive|unknown|negative)
    Reason       string (required if negative, min 30 chars)

  ReferenceFormData (returned on GET token endpoint):
    ApplicantName   string
    MembershipType  string
    RefereeName     string
    ExpiresAt       string

repository.go:
  CreateBatch(ctx, refs []Reference) error
  FindByTokenHash(ctx, hash) (*Reference, error)
  FindByApplicationID(ctx, appID) ([]*Reference, error)
  MarkTokenUsed(ctx, refID) error
  CreateResponse(ctx, response ReferenceResponse) error
  CountByApplicationAndType(ctx, appID, responseType) (int, error)
  CreateReplacement(ctx, appID, name, email string, round int) (*Reference, error)

service.go:

  CreateForApplication(ctx, appID, referees []ReferenceInput) error:
    For each referee:
    1. Look up user by ID (must exist and be active)
    2. Generate token (raw, hashed, expiresAt)
    3. Create Reference record with hashed token
    4. Send reference_request.html email with raw token in URL
    5. Log: action="ref.sent"
    Update application status → referans_bekleniyor

  GetFormData(ctx, rawToken) (*ReferenceFormData, error):
    1. Hash the raw token
    2. Find reference by hash
    3. Check expiry: if expired return ErrTokenExpired
    4. Check token_used: if true return ErrTokenUsed
    5. Load application data
    6. Return form data (do NOT consume token yet — only on POST)

  SubmitResponse(ctx, rawToken, req, ipAddress, userAgent string) error:
    1. Hash raw token
    2. Find reference by hash
    3. Validate + consume token (atomic):
       - Check expiry → 410 if expired
       - Check token_used → 409 if used
       - Set token_used = true (in same transaction as response save)
    4. Save ReferenceResponse
    5. Log: action="ref.responded", metadata: { response_type, ip }
    6. Execute response logic:

       IF negative:
         Call RedGuard.Terminate(appID, req.Reason, "system", "referee")
         Return

       IF unknown:
         Send applicant_new_ref_needed.html email to applicant
         Create replacement reference slot (is_replacement=true, round++)
         Generate new token for replacement
         (Do NOT advance status — wait for new ref)
         Log: action="ref.replacement_requested"
         Return

       IF positive:
         Check if ALL references for this application are now positive
         (positive count + replacement positive count >= 3,
          no pending unknowns remaining)
         If complete:
           ValidateTransition(type, current, referans_tamamlandı)
           Update status → referans_tamamlandı
           Log: action="status.change"

  ResendToken(ctx, refID, koordinatorID) error:
    1. Find reference by ID
    2. Generate new token (new expiry)
    3. Update token_hash and token_expires in DB
    4. Resend reference_request.html email
    5. Log: action="ref.resent"

HANDLER (public routes — NO authentication middleware):

handler.go:
  GET  /api/v1/ref/respond/:token
    → Call service.GetFormData(rawToken)
    → Return form data as JSON

  POST /api/v1/ref/respond/:token
    → Parse body, validate with validator
    → Call service.SubmitResponse(rawToken, req, ip, ua)
    → Return success message

  POST /api/v1/applications/:id/references/resend/:refId
    → KoordinatorOnly middleware
    → Call service.ResendToken

EMAIL TEMPLATES (apps/api/internal/features/notifications/templates/):

reference_request.html:
  Subject: "[Membership System] Referans Onayı Bekleniyor"
  Body:
    "Sayın {{.RefereeName}},
    {{.ApplicantName}} adlı kişi {{.MembershipType}} üyeliği için başvurmuştur
    ve sizi referans olarak göstermiştir.
    Görüşünüzü bildirmek için aşağıdaki linke tıklayınız:
    {{.ResponseURL}}
    Bu link {{.ExpiresAt}} tarihinde geçerliliğini yitirecektir.
    Link yalnızca bir kez kullanılabilir."

applicant_new_ref_needed.html:
  Subject: "[Membership System] Yeni Referans Gerekiyor"
  Body:
    "Sayın {{.ApplicantName}},
    {{.UnknownRefereeName}} adlı referansınız sizi tanımadığını bildirmiştir.
    Lütfen sisteme giriş yaparak yeni bir referans ekleyiniz.
    Portal: {{.PortalURL}}"

FRONTEND PUBLIC PAGES (apps/admin/src/app/respond/):

respond/reference/[token]/page.tsx:
  1. On mount: call GET /api/v1/ref/respond/[token]
     - 410 → render <TokenExpiredPage />
     - 409 → render <TokenUsedPage />
     - 200 → render response form

  Response form:
    Three radio options:
    - "Tanıyorum ve görüşüm olumludur" (positive)
    - "Bu kişiyi tanımıyorum" (unknown)
    - "Tanıyorum ancak görüşüm olumsuzdur" (negative)

    Conditional textarea:
    - Show ONLY when negative is selected
    - Label: "Lütfen olumsuz görüşünüzün gerekçesini yazınız"
    - Required, min 30 characters
    - Zod validation: reason required when negative

  On submit:
    POST /api/v1/ref/respond/[token]
    Success → render <ThankYouPage message="Yanıtınız kaydedildi." />
    Error → show inline error

  TokenExpiredPage: "Bu referans linki süresi dolmuştur."
  TokenUsedPage: "Bu link daha önce kullanılmıştır."
  ThankYouPage: "Yanıtınız alındı. Katkınız için teşekkür ederiz."

  All pages: mobile responsive, centered card layout, no nav/header.

VERIFICATION:
- Application submission creates 3 reference records
- MailDev receives reference_request emails
- GET /ref/respond/:token returns form data
- POST /ref/respond/:token with positive advances status when all done
- POST /ref/respond/:token with negative terminates application
- POST /ref/respond/:token with expired token returns 410
- POST /ref/respond/:token second time returns 409
- Public pages render correctly without authentication
```

---

---

# PHASE 6 — Consultation System (Profesyonel & Öğrenci)

```
Read INSTRUCTIONS.md Section 9 (Consultation System).

Build the consultation module for Profesyonel and Öğrenci applications.

CONSULTATION FEATURE (apps/api/internal/features/consultations/):

dtos.go:
  AddConsultationsRequest:
    Consultees []ConsulteeInput (required, min 2)

  ConsulteeInput:
    UserID string (required — must be existing active system user)

  ConsultationResponseRequest:
    ResponseType string (required: positive|negative)
    Reason       string (required if negative, min 30 chars)

  ConsultationFormData:
    ApplicantName  string
    MembershipType string
    MemberName     string
    ExpiresAt      string

repository.go:
  CreateBatch(ctx, consultations []Consultation) error
  FindByTokenHash(ctx, hash) (*Consultation, error)
  FindByApplicationID(ctx, appID) ([]*Consultation, error)
  MarkTokenUsed(ctx, id) error
  SaveResponse(ctx, id, responseType, reason string) error
  CountByApplicationAndType(ctx, appID, responseType) (int, error)
  CountTotal(ctx, appID) (int, error)

service.go:

  AddConsultees(ctx, appID, req, koordinatorID) error:
    1. Assert application type is profesyonel or ogrenci
    2. Assert application status = başvuru_alındı
    3. Assert RedGuard: not terminated
    4. Validate minimum 2 consultees
    5. For each consultee:
       a. Load user by ID (must exist, must be active)
       b. Generate token
       c. Create Consultation record
       d. Send consultation_request.html email
       e. Log: action="consult.sent"
    6. Advance status → danışma_sürecinde
    7. Log: action="status.change"

  GetFormData(ctx, rawToken) (*ConsultationFormData, error):
    1. Hash raw token
    2. Find consultation by hash
    3. Check expiry → ErrTokenExpired
    4. Check used → ErrTokenUsed
    5. Load application
    6. Return form data

  SubmitResponse(ctx, rawToken, req, ip string) error:
    1. Hash + validate + consume token (atomic)
    2. Save response to consultation record
    3. Log: action="consult.responded"
    4. IF negative:
         RedGuard.Terminate(appID, req.Reason, "system", "consulted_member")
         Return
    5. IF positive:
         Check: all consultations for this appID have response_type = positive
         (no pending, no negative)
         If all positive:
           ValidateTransition(type, danışma_sürecinde, gündemde)
           Update status → gündemde
           Log: action="status.change"

handler.go:
  POST /api/v1/applications/:id/consultations
    → KoordinatorOnly middleware
    → service.AddConsultees

  GET  /api/v1/consult/respond/:token
    → Public (no auth)
    → service.GetFormData

  POST /api/v1/consult/respond/:token
    → Public (no auth)
    → service.SubmitResponse

EMAIL TEMPLATE (notifications/templates/consultation_request.html):
  Subject: "[Membership System] Üye Danışma Talebi"
  Body:
    "Sayın {{.MemberName}},
    {{.ApplicantName}} adlı kişi {{.MembershipType}} üyeliği için başvurmuştur.
    LinkedIn: {{.ApplicantLinkedIn}}
    Görüşünüzü bildirmek için: {{.ResponseURL}}
    Link geçerlilik süresi: {{.ExpiresAt}}"

FRONTEND PUBLIC PAGE:

respond/consultation/[token]/page.tsx:
  Same pattern as reference response page.
  Two options (no unknown option):
  - "Olumlu görüşüm var"  (positive)
  - "Olumsuz görüşüm var" (negative) → shows reason textarea

FRONTEND ADMIN PANEL:

components/consultation/ConsultationPanel.tsx:
  Shown in application detail under "Danışma" tab.
  Only rendered for profesyonel/ogrenci type applications.
  Shows: list of consultees, their response status (pending/positive/negative).
  Koordinator action: "Danışman Ekle" button → opens member search modal.

app/(dashboard)/applications/[id]/consultation/page.tsx:
  Consultation management page.
  List of all consultations with status.
  Add consultees form (autocomplete member search).
  Accessible to: koordinator, admin.

VERIFICATION:
- POST /consultations with < 2 consultees → 422
- POST /consultations creates records + sends emails
- Positive response from all → status advances to gündemde
- Any negative response → application terminated immediately
- Consultation tab hidden for asil/akademik applications
```

---

---

# PHASE 7 — Email Notification System

```
Read INSTRUCTIONS.md Section 14 (Email System) and Section 7 (Veto Model).

Build the complete email service with all templates and retry logic.

MAILER (apps/api/internal/features/notifications/mailer.go):

  MailerConfig struct:
    Host, Port string
    From, FromName string

  Mailer struct with gomail.Dialer

  Send(to, subject, htmlBody, textBody string) error:
    Compose gomail.Message
    Set headers: From, To, Subject, Content-Type
    Set HTML body + plain-text alternative
    Dial and Send
    On error: retry 3 times with backoff (1s, 2s, 4s)
    After all retries fail: return wrapped error

TEMPLATE RENDERER (notifications/templates.go):

  TemplateData interface{} (each template has its own typed struct)

  Render(templateName string, data interface{}) (html, text string, error):
    Load and parse HTML template from templates/ directory
    Execute with data
    Generate plain-text version (strip HTML tags)
    Return both

  All template data structs:
    ReferenceRequestData { RefereeName, ApplicantName, MembershipType,
                           ResponseURL, ExpiresAt }
    ReferenceReminderData { RefereeName, ApplicantName, ResponseURL,
                            HoursRemaining }
    NewRefNeededData      { ApplicantName, UnknownRefereeName, PortalURL }
    ReputationQueryData   { ContactName, ApplicantName, ApplicantLinkedIn,
                            ResponseURL, ExpiresAt }
    ConsultationData      { MemberName, ApplicantName, ApplicantLinkedIn,
                            MembershipType, ResponseURL, ExpiresAt }
    AcceptedData          { ApplicantName, MembershipType }
    RejectedData          { ApplicantName }
                          ← NO reason field. Never disclose why.
    HonoraryProposalData  { YKMemberName, ProposerName, NomineeName,
                            NomineeLinkedIn, ProposalReason, ReviewURL }

NOTIFICATION SERVICE (notifications/service.go):

  NotificationService with Mailer + LogService dependencies.

  Methods (one per email type):
  SendReferenceRequest(ctx, ref Reference, app Application) error
  SendReferenceReminder(ctx, ref Reference, app Application) error
  SendNewRefNeeded(ctx, applicantEmail, applicantName, unknownRefName string) error
  SendReputationQuery(ctx, contact ReputationContact, app Application) error
  SendConsultationRequest(ctx, consult Consultation, app Application) error
  SendAccepted(ctx, app Application) error
  SendRejected(ctx, app Application) error
    ← CRITICAL: must NOT include rejection_reason in email body
  SendHonoraryProposal(ctx, app Application, ykMembers []User) error

  Each method:
    1. Render template
    2. Call mailer.Send
    3. Log action="email.sent" with metadata: { template, recipient, appID }
    4. On failure: log action="email.failed" with error details

ALL EMAIL TEMPLATES (notifications/templates/):
  Create all 8 HTML templates listed above.
  Each template:
  - Professional header with system name
  - Clear call-to-action button
  - Expiry warning where applicable
  - Plain Turkish language
  - Mobile-responsive HTML table layout
  - Footer: "Bu e-posta otomatik olarak gönderilmiştir."

REMINDER CRON JOB:
  Create apps/api/internal/features/notifications/cron.go:
  ReminderJob that runs every hour:
    SELECT references WHERE:
      token_used = false
      AND responded_at IS NULL
      AND token_expires BETWEEN NOW() AND NOW() + INTERVAL 24 HOUR
    For each: SendReferenceReminder
    Log: action="email.reminder_sent"

  Register cron in main.go using time.Ticker (every 1 hour).

VERIFICATION:
- All 8 email types appear in MailDev UI when triggered
- Rejection email does NOT contain rejection reason
- Retry logic: mock SMTP failure → 3 attempts logged
- Reminder cron fires and sends emails for near-expiry refs
- All email sends logged in logs table
```

---

---

# PHASE 8 — Reputation System

```
Read INSTRUCTIONS.md Section 10 (Reputation System).

Build the reputation contact query system for Asil & Akademik applications.

REPUTATION FEATURE (apps/api/internal/features/reputation/):

dtos.go:
  AddContactsRequest:
    Contacts []ContactInput (required, exactly 10)

  ContactInput:
    Name  string (required)
    Email string (required, valid email)

  ContactResponseRequest:
    ResponseType string (required: clean|negative)
    Reason       string (required if negative, min 30 chars)

  ReputationStatusResponse:
    ApplicationID string
    TotalContacts int
    Responded     int
    Clean         int
    Flagged       int
    Contacts      []ContactStatus

  ContactStatus:
    ContactName   string
    Email         string (masked: j***@example.com)
    Status        string (pending|clean|flagged)
    RespondedAt   *time.Time

repository.go:
  CreateBatch(ctx, contacts []ReputationContact) error
  FindByApplicationID(ctx, appID) ([]*ReputationContact, error)
  FindByTokenHash(ctx, hash) (*ReputationContact, error)
  MarkTokenUsed(ctx, id) error
  SaveResponse(ctx, id, responseType, reason string) error
  CountByApplicationAndType(ctx, appID, responseType) (int, error)
  CountTotal(ctx, appID) (int, error)
  CountResponded(ctx, appID) (int, error)

service.go:

  AddContacts(ctx, appID, req, actorID, actorRole) error:
    1. Assert application type = asil OR akademik
    2. Assert application status = ön_onaylandı
    3. Assert RedGuard: not terminated
    4. CRITICAL: Validate exactly 10 contacts (len(req.Contacts) == 10)
       If not: return validation error "Exactly 10 contacts required"
    5. For each contact:
       a. Generate token
       b. Create ReputationContact record
       c. Send reputation_query.html email
       d. Log: action="rep.contact_added"
    6. Advance application status → itibar_taramasında
    7. Log: action="status.change"

  GetFormData(ctx, rawToken) (*ReputationFormData, error):
    Validate token (not expired, not used).
    Return: ContactName, ApplicantName, ApplicantLinkedIn, ExpiresAt.

  SubmitResponse(ctx, rawToken, req, ip string) error:
    1. Hash + validate + consume token (atomic)
    2. Save response
    3. Log: action="rep.responded"

    4. IF negative:
         Do NOT auto-terminate.
         Instead:
         a. Update contact record: response_type = negative
         b. Notify YK members via email:
            "İtibar taramasında olumsuz geri dönüş var. İnceleme gerekiyor."
         c. Advance status → (keep itibar_taramasında, flag = has_negative)
            OR create a separate "flagged" state in metadata
         d. YK must then cast a vote to terminate or continue
         Return

    5. IF clean:
         Check: all 10 contacts responded AND none are negative
         If yes: Advance status → itibar_temiz
                 Log: action="status.change"

  GetStatus(ctx, appID) (*ReputationStatusResponse, error):
    Return aggregated status.
    Mask contact email addresses (show j***@domain.com).
    Accessible to: yk, koordinator, admin.

handler.go:
  POST /api/v1/applications/:id/reputation/contacts
    → YKOrKoordinator middleware
    → service.AddContacts

  GET  /api/v1/applications/:id/reputation
    → YKOrKoordinator middleware
    → service.GetStatus

  GET  /api/v1/reputation/respond/:token
    → Public (no auth)
    → service.GetFormData

  POST /api/v1/reputation/respond/:token
    → Public (no auth)
    → service.SubmitResponse

EMAIL TEMPLATE (reputation_query.html):
  Subject: "[Membership System] Üye Adayı Hakkında Bilgi Talebi"
  Body:
    "Sayın {{.ContactName}},
    {{.ApplicantName}} adlı kişi hakkında üyelik başvurusu bulunmaktadır.
    LinkedIn: {{.ApplicantLinkedIn}}
    SORU: Bu kişi hakkında olumsuz bir bilginiz var mı?
    Yanıtlamak için: {{.ResponseURL}}
    Link geçerliliği: {{.ExpiresAt}}"

FRONTEND PUBLIC PAGE:

respond/reputation/[token]/page.tsx:
  Two options:
  - "Hayır, olumsuz bir bilgim yok" (clean)
  - "Evet, olumsuz bilgim var" (negative) → reason textarea (required, min 30 chars)

FRONTEND ADMIN PANEL:

components/reputation/ReputationPanel.tsx:
  Only shown for asil/akademik applications.
  Shows: progress (X/10 responded), clean count, flagged count.
  Contact list with masked emails and status chips.
  Add contacts form: 10 email inputs (bulk entry).

components/reputation/ContactQueryList.tsx:
  Table of contacts with: masked email, name, status, response date.

app/(dashboard)/applications/[id]/reputation/page.tsx:
  Full reputation management page.
  Accessible to: yk, koordinator, admin.
  Hidden for profesyonel/ogrenci/onursal types.

VERIFICATION:
- POST /reputation/contacts with != 10 contacts → 422
- All 10 contacts receive emails in MailDev
- Clean responses from all 10 → status advances to itibar_temiz
- Negative response → YK notified, status stays, YK can then vote to terminate
- Status endpoint returns masked emails
- Reputation tab hidden for non-asil/akademik applications
```

---

---

# PHASE 9 — Voting Module

```
Read INSTRUCTIONS.md Section 11 (Voting System) and Section 7 (Veto Model).

Build the three-stage voting system with single-veto termination.

VOTING FEATURE (apps/api/internal/features/voting/):

dtos.go:
  CastVoteRequest:
    VoteType string (required: approve|abstain|reject)
    Reason   string (required if reject, min 20 chars)

  VoteResponse:
    ID            string
    VoterRole     string
    VoteStage     string
    VoteType      string
    Reason        string  ← only returned to yk/admin roles
    IsVeto        bool
    CreatedAt     string

  VoteSummaryResponse:
    Stage         string
    TotalVoters   int
    Approved      int
    Abstained     int
    Rejected      int
    IsTerminated  bool
    Votes         []VoteResponse

repository.go:
  Create(ctx, vote Vote) error
  FindByApplicationAndStage(ctx, appID, stage) ([]*Vote, error)
  FindByApplicationVoterStage(ctx, appID, voterID, stage) (*Vote, error)
  CountActiveVotersByRole(ctx, role) (int, error)
  CountVotesByStageAndType(ctx, appID, stage, voteType) (int, error)

service.go:

  CastVote(ctx, appID, voterID, voterRole, stage, req) error:

    PRE-CHECKS:
    1. AssertNotTerminated(appID)
    2. Load application
    3. Validate stage is appropriate for this application type:
       - yk_prelim: only asil, akademik, onursal
       - yik: only onursal
       - yk_final: all types (when in gündemde)
    4. Validate voter role matches stage:
       - yk_prelim + yk_final → role must be "yk"
       - yik → role must be "yik"
       If mismatch → ErrForbidden
    5. Check duplicate vote:
       FindByApplicationVoterStage(appID, voterID, stage)
       If found → ErrDuplicateVote (409)
    6. Validate required reason if reject:
       If VoteType = reject AND len(Reason) < 20 → validation error

    EXECUTE VOTE:
    7. Create Vote record (is_veto = VoteType == reject)
    8. Log: action="vote.cast", metadata: { stage, voteType, voterRole }

    VETO CHECK:
    9. If VoteType = reject:
       a. Set vote.is_veto = true
       b. Call RedGuard.Terminate(appID, reason, voterID, voterRole)
       c. Log: action="vote.veto"
       d. Return (process terminated)

    COMPLETION CHECK (only if approve/abstain):
    10. Count total active voters for this role
        Count votes cast in this stage for this application
        If all eligible voters have voted AND no reject:
          Determine next status:
          - yk_prelim complete → ön_onaylandı
          - yik complete (no negative) → gündemde
          - yk_final complete → kabul
          ValidateTransition and advance status
          Log: action="status.change"

  GetVotes(ctx, appID, stage, requestorRole) (*VoteSummaryResponse, error):
    Load votes for stage.
    If requestorRole is not yk/admin:
      Omit Reason field from all VoteResponse objects
      Return counts only
    Return full details to yk/admin.

handler.go:
  POST /api/v1/applications/:id/votes/yk-prelim
    → Auth + YKOnly middleware
    → service.CastVote(appID, voterID, "yk", "yk_prelim", req)

  POST /api/v1/applications/:id/votes/yik
    → Auth + YIKOnly middleware
    → Validate application type = onursal first
    → service.CastVote(appID, voterID, "yik", "yik", req)

  POST /api/v1/applications/:id/votes/yk-final
    → Auth + YKOnly middleware
    → Validate application status = gündemde first
    → service.CastVote(appID, voterID, "yk", "yk_final", req)

  GET  /api/v1/applications/:id/votes
    → Auth + YKOrAdmin middleware
    → service.GetVotes(appID, stage query param, requestorRole)

FRONTEND:

components/voting/VotePanel.tsx:
  Props: applicationID, stage, existingVote (optional)

  States:
  - If existingVote present: show read-only "You voted: [type]" message
  - If application terminated: show VetoAlert, disable all buttons
  - Default: show 3 buttons

  Vote buttons:
  - "Onayla" (green) → vote_type: approve
  - "Çekimser" (gray) → vote_type: abstain
  - "RED" (red) → vote_type: reject → shows reason textarea

  Reason textarea (shown only when RED selected):
    Label: "Lütfen red gerekçenizi yazınız (zorunlu)"
    Min 20 characters, character counter
    Required validation (Zod)

  Confirmation dialog before any vote submit:
    "Bu kararınız geri alınamaz. Emin misiniz?"
    Confirm + Cancel buttons

  On submit: POST to appropriate stage endpoint
  Optimistic UI: disable buttons immediately after submit

components/voting/VetoAlert.tsx:
  Full-width red banner shown when application is terminated.
  Content: "Bu başvuru [date] tarihinde sonlandırılmıştır."
  For yk/admin: show "[role] tarafından reddedilmiştir."
  For others: show generic message only.

components/voting/VoteSummary.tsx:
  Shows vote counts: Onayla X | Çekimser Y | RED Z
  For yk/admin: shows each voter's role + vote type in table
  For others: shows counts only (no voter identification)

app/(dashboard)/applications/[id]/voting/yk-prelim/page.tsx
app/(dashboard)/applications/[id]/voting/yik/page.tsx
app/(dashboard)/applications/[id]/voting/yk-final/page.tsx:
  Each page:
  - Shows application summary card
  - VoteSummary for previous votes in this stage
  - VotePanel (if current user hasn't voted yet)
  - VetoAlert (if terminated)
  - Correct role guards per page

VERIFICATION:
- YK member votes approve on yk-prelim → recorded
- Second vote attempt from same voter → 409
- Any reject vote → application terminated, VetoAlert shown
- Non-YK trying yk-prelim endpoint → 403
- YK trying yik endpoint → 403
- Votes summary hides reasons from non-YK
- All 3 voting pages render correctly per role
```

---

---

# PHASE 10 — Honorary Membership Flow

```
Read INSTRUCTIONS.md Section 12 (Honorary Membership Flow).

Build the complete Onursal üye proposal and approval system.

HONORARY FEATURE (apps/api/internal/features/honorary/):

dtos.go:
  ProposeRequest:
    NomineeName     string (required, min 2)
    NomineeLinkedIn string (required, valid URL)
    ProposalReason  string (required, min 100 chars)

  ProposalResponse:
    ApplicationID   string
    NomineeName     string
    NomineeLinkedIn string
    ProposalReason  string
    ProposedBy      string (proposer full name)
    Status          string
    CreatedAt       string

repository.go:
  Create(ctx, application Application) error
  FindAll(ctx) ([]*Application, error)

service.go:

  Propose(ctx, req ProposeRequest, proposerID string) (*Application, error):
    1. Validate: proposer role must be asil_uye OR yik_uye
       (load user from DB to confirm — don't trust claims alone)
    2. Check LinkedIn URL uniqueness
    3. REPEAT APPLICANT CHECK (same as Phase 4)
    4. Create Application record:
       membership_type = onursal
       status = öneri_alındı
       proposed_by_user_id = proposerID
       proposal_reason = req.ProposalReason
       applicant_name = req.NomineeName
       linkedin_url = req.NomineeLinkedIn
    5. Log: action="application.created", metadata: { proposedBy: proposerID }
    6. Load all active YK members
    7. For each YK member: send honorary_proposal_notify.html email
    8. Log: action="email.sent" for each notification
    9. Return created application

  ListProposals(ctx) ([]*ProposalResponse, error):
    Return all onursal applications with proposer name populated.

handler.go:
  POST /api/v1/honorary/propose
    → Auth + ProposerOnly middleware (asil_uye OR yik_uye)
    → service.Propose

  GET  /api/v1/honorary
    → Auth + YKOrAdmin middleware
    → service.ListProposals

After proposal is created:
  The existing voting module handles all subsequent stages:
  - YK Ön: POST /api/v1/applications/:id/votes/yk-prelim
  - YİK:   POST /api/v1/applications/:id/votes/yik
  - YK Final: POST /api/v1/applications/:id/votes/yk-final

EMAIL TEMPLATE (honorary_proposal_notify.html):
  Subject: "[Membership System] Onursal Üye Önerisi"
  Body:
    "Sayın {{.YKMemberName}},
    {{.ProposerName}} tarafından yeni bir Onursal Üye önerisi yapılmıştır.
    Aday: {{.NomineeName}}
    LinkedIn: {{.NomineeLinkedIn}}
    Gerekçe: {{.ProposalReason}}
    İncelemek için: {{.ReviewURL}}"

FRONTEND:

app/(dashboard)/honorary/page.tsx:
  List of all honorary proposals.
  Table columns: Nominee Name, LinkedIn, Proposed By, Status, Created.
  "Yeni Öneri" button → navigates to /honorary/new.
  Only shown for asil_uye, yik_uye, yk, admin roles.

app/(dashboard)/honorary/new/page.tsx:
  Proposal form:
  - Nominee Full Name (required)
  - Nominee LinkedIn URL (required, URL validation)
  - Proposal Reason (required, min 100 chars, character counter)
  Submit → POST /api/v1/honorary/propose
  Success → redirect to /applications/[new_id]

  Role guard: only asil_uye and yik_uye can access this page.
  Others → redirect to /applications.

SIDEBAR NAVIGATION UPDATE:
  Add "Onursal Öneri" menu item in sidebar.
  Visible only to: asil_uye, yik_uye, yk, admin.

VERIFICATION:
- asil_uye user can access /honorary/new
- koordinator accessing /honorary/new → redirect or 403
- POST /honorary/propose creates application with status=öneri_alındı
- All YK members receive notification email in MailDev
- Created application visible in /applications list
- Voting stages work correctly for onursal type
```

---

---

# PHASE 11 — Web Publish Consent & Public Member List

```
Read INSTRUCTIONS.md Section 16 (API Endpoints — Web Publish section).

Build the post-acceptance web publish consent system and public member list.

WEB PUBLISH FEATURE (apps/api/internal/features/webpublish/):

dtos.go:
  RecordConsentRequest:
    Consented bool (required)

  MemberListItem:
    FullName       string
    MembershipType string
    AcceptedAt     string

repository.go:
  RecordConsent(ctx, appID, consented bool, adminID string) error
  FindPublishedMembers(ctx) ([]*Application, error)
  ConsentExists(ctx, appID) (bool, error)

service.go:

  RecordConsent(ctx, appID, req, adminID) error:
    1. Load application by ID
    2. Assert: application status must be kabul
       If not: return error "Consent can only be recorded for accepted applications"
    3. Assert: consent not already recorded (ConsentExists check)
       If exists: return error "Consent already recorded"
    4. In transaction:
       a. Create WebPublishConsent record
       b. Update application.web_publish_consent = req.Consented
       c. If consented = true: set application.is_published = true
       d. Log: action="publish.consent_recorded",
              metadata: { consented, appID, adminID }
    5. Return nil

  GetPublishedMembers(ctx) ([]*MemberListItem, error):
    Load applications WHERE is_published = true AND status = kabul
    Sort alphabetically by applicant_name ASC
    Return mapped MemberListItem list (no sensitive fields)

handler.go:
  POST /api/v1/applications/:id/publish-consent
    → Auth + AdminOnly middleware
    → service.RecordConsent

  GET  /api/v1/members
    → NO authentication required (public endpoint)
    → service.GetPublishedMembers

FRONTEND:

app/(dashboard)/applications/[id]/webpublish/page.tsx:
  Web publish consent page.
  Only accessible to admin role.
  Only shown when application status = kabul.

  UI:
  - Applicant name + membership type
  - Toggle/radio: "Web sitesinde yayınlansın mı?"
    - "Evet, alfabetik listede yayınlansın"
    - "Hayır, yalnızca iç listede kalsın"
  - Submit button: "Onayla"
  - If consent already recorded: show read-only state

  After submit: show success toast, navigate back to application detail.

app/(dashboard)/members/page.tsx:
  Alphabetical published member list.
  Accessible to all authenticated users.
  Columns: Full Name, Membership Type, Accepted Date.
  Sorted A-Z by full name.
  "Web'de Yayınlananlar" page title.

UPDATE ApplicationDetail page:
  In app/(dashboard)/applications/[id]/page.tsx:
  After status = kabul:
  - Show "Web Yayın Onayı" tab/button for admin users
  - Show current web_publish_consent value as badge:
    - NULL → "Karar Bekleniyor" (gray)
    - true → "Yayında" (green)
    - false → "İç Listede" (yellow)

VERIFICATION:
- POST /publish-consent on non-kabul application → 422
- POST /publish-consent with consented=true → is_published set to true
- GET /members returns only published members alphabetically sorted
- GET /members requires no authentication
- Admin sees "Web Yayın" tab on accepted applications
- Double-consent attempt → error
```

---

---

# PHASE 12 — Admin UI (Complete)

```
Read INSTRUCTIONS.md Section 19 (Frontend Rules) and Section 4 (RBAC Matrix).

Complete all remaining admin UI pages, navigation, and polish.

LAYOUT & NAVIGATION:

app/(dashboard)/layout.tsx:
  Sidebar navigation with role-based menu items:

  ALWAYS VISIBLE (all authenticated roles):
  - Dashboard (/) → redirect to /applications
  - Applications (/applications)
  - Members (/members)

  VISIBLE TO: yk, admin, koordinator:
  - Logs (/logs)

  VISIBLE TO: asil_uye, yik_uye, yk, admin:
  - Onursal Öneriler (/honorary)

  VISIBLE TO: asil_uye, yik_uye:
  - Yeni Onursal Öneri (/honorary/new)

  Header: show current user name + role badge + logout button.
  Mobile: collapsible sidebar with hamburger menu.

DASHBOARD PAGE (app/(dashboard)/page.tsx):
  Redirect to /applications.

APPLICATION LIST (app/(dashboard)/applications/page.tsx):
  Already built in Phase 4. Add these enhancements:
  - "Yeni Başvuru" button → opens slide-over form (if public form needed)
  - Export to CSV button (admin only) → downloads filtered results
  - Statistics row above table:
    { Total, Pending, Accepted, Rejected } count cards

APPLICATION DETAIL (app/(dashboard)/applications/[id]/page.tsx):
  Already built in Phase 4. Complete the tab content:

  Tab: "Referanslar" (visible for asil/akademik):
    Import and render ReferenceGrid component.
    Show each referee: name, email, status chip, responded_at.
    Koordinator action: resend button per unanswered ref.

  Tab: "Danışma" (visible for profesyonel/ogrenci):
    Import and render ConsultationPanel component.
    Show consultees list with status.

  Tab: "İtibar Taraması" (visible for asil/akademik, yk/koordinator):
    Import and render ReputationPanel component.

  Tab: "Oylar" (visible for yk/admin):
    Show VoteSummary for each applicable voting stage.
    Show VotePanel if current user can vote in current stage.

  Tab: "Web Yayın" (visible for admin, when status=kabul):
    Web publish consent UI from Phase 11.

LOGS VIEWER (app/(dashboard)/logs/page.tsx):
  Already built in Phase 3 (stub). Complete implementation:
  TanStack Table with server-side pagination.
  Columns: Timestamp, Actor Role, Action, Entity Type, Entity ID, IP
  Filters: action (text), entity_type (select), date range (date pickers)
  Row click → LogDetailDrawer (slide-over with formatted JSON metadata)
  Actor name shown only to admin (YK sees role only).
  Route protected: admin + yk only.

GLOBAL COMPONENTS:

components/ui/RoleBadge.tsx:
  Color-coded badge per role:
  admin → purple, yk → blue, yik → indigo,
  koordinator → teal, asil_uye → green, yik_uye → cyan

components/ui/ConfirmDialog.tsx:
  Reusable confirmation dialog.
  Props: title, description, onConfirm, onCancel, destructive (bool)

components/ui/PageHeader.tsx:
  Consistent page header with title + optional action button.

components/ui/EmptyState.tsx:
  Shown when table/list has no results.
  Props: icon, title, description, action (optional button)

REACT QUERY SETUP (app/providers.tsx):
  QueryClient with defaults:
    staleTime: 30 seconds
    retry: 1
    refetchOnWindowFocus: false
  Wrap app in QueryClientProvider + Toaster (shadcn/ui)

ERROR BOUNDARIES:
  Create app/error.tsx (Next.js error boundary)
  Create app/not-found.tsx (404 page)
  Create components/ui/ErrorPage.tsx with "Hata oluştu" message

LOADING STATES:
  Create app/loading.tsx (skeleton loader)
  Add Suspense boundaries around table components
  Skeleton components for: ApplicationTable, VoteSummary, LogViewer

VERIFICATION:
  Walk through complete workflow in the UI:
  1. Login as koordinator → verify sidebar shows correct items
  2. View /applications → table loads with filters
  3. Open asil application → tabs show: Referanslar, İtibar, Oylar
  4. Open profesyonel application → tabs show: Danışma, Oylar
  5. Open onursal application → tabs show: YK/YİK/YK votes
  6. Login as yk → vote on an application → VetAlert appears on reject
  7. Login as admin → view /logs → filter by entity_type
  8. Login as asil_uye → /honorary/new visible, /logs not visible
  9. Logout → redirected to /login
```

---

---

# PHASE 13 — Audit, Security Hardening & Final Polish

```
Read INSTRUCTIONS.md Section 17 (Security Rules), Section 18 (Audit Rules),
and Section 22 (Forbidden Actions).

Complete all security hardening, audit wiring, and final verification.

AUDIT MIDDLEWARE COMPLETION (apps/api/internal/middleware/audit.go):
  Ensure ALL routes trigger audit logging.
  Verify these specific actions are logged:
  - POST /applications → application.created
  - POST /votes/* → vote.cast OR vote.veto
  - POST /ref/respond/:token → ref.responded
  - POST /consult/respond/:token → consult.responded
  - POST /reputation/respond/:token → rep.responded
  - POST /publish-consent → publish.consent_recorded
  - Any status change → status.change
  - Any RedGuard.Terminate call → application.terminated
  - Any email send → email.sent
  - Auth login/logout → auth.login / auth.logout

  Each log entry MUST include:
  - actor_id (from JWT claims or "system" for automated actions)
  - actor_role
  - entity_type + entity_id
  - ip_address (from fiber.Ctx)
  - metadata (JSON with contextual data)
  - created_at (server time)

SECURITY HEADERS (apps/api/cmd/main.go):
  Add Fiber middleware:
  - helmet (security headers: X-Frame-Options, X-XSS-Protection, etc.)
  - CORS: only allow APP_BASE_URL origin in production
  - Rate limiting on public token endpoints:
    - GET/POST /ref/respond/* → 10 req/min per IP
    - GET/POST /consult/respond/* → 10 req/min per IP
    - GET/POST /reputation/respond/* → 10 req/min per IP
    - POST /auth/login → 5 req/min per IP

IMMUTABILITY FINAL VERIFICATION:
  Write Go test in features/applications/service_test.go:
  TestRejectionReasonImmutable:
    1. Create application
    2. Set rejection_reason via RedGuard.Terminate
    3. Attempt service-level update of rejection_reason
    4. Assert: ErrImmutableField returned
    5. Assert: DB value unchanged

  Write Go test for logs immutability:
  TestLogsAppendOnly:
    1. Create log entry
    2. Attempt UPDATE via raw SQL
    3. Assert: error returned (trigger fires)
    4. Attempt DELETE via raw SQL
    5. Assert: error returned

REPEAT APPLICANT — FINAL WIRE-UP:
  Verify the complete flow:
  1. Submit application with email A → gets rejected
  2. Submit new application with email A
  3. Assert: new application has previous_app_id set
  4. Assert: GET /applications/:id returns repeat_applicant: true
  5. Assert: GET /applications/:id/red-history returns previous apps
  6. Assert: non-YK role cannot access /red-history (403)

INPUT VALIDATION HARDENING:
  All handler input validation must use go-playground/validator.
  Add these specific validations:
  - linkedin_url: must start with https://linkedin.com/ or https://www.linkedin.com/
  - email: RFC 5322 format
  - photo_url: must be valid URL if provided
  - token in URL: must be non-empty, 36+ chars (UUID format)

  On validation error: return 422 with field-level errors:
  { "error": { "code": "VALIDATION_ERROR", "fields": { "field": "message" } } }

FRONTEND SECURITY:
  apps/admin:
  - All pages under (dashboard) wrapped in auth guard:
    Check accessToken in auth store
    If missing → redirect to /login
  - Role guards on all sensitive pages:
    /logs → redirect if not yk/admin
    /honorary/new → redirect if not asil_uye/yik_uye
    /applications/:id/red-history → only render for yk/admin
  - Axios interceptor: on 401 → clear store + redirect to /login
  - Never log accessToken to console

ENVIRONMENT HARDENING:
  apps/api:
  - Validate all required env vars on startup
  - If any required var missing: log fatal error and exit
  - In production mode: enforce HTTPS-only CORS
  - Log startup: print role of running environment (dev/prod)

FINAL INTEGRATION TEST (manual walkthrough):

  Complete this full scenario and verify at each step:

  SCENARIO: Asil Üye — Happy Path
  1. Submit asil application with 3 refs → status: referans_bekleniyor
  2. All 3 refs respond positive → status: referans_tamamlandı
  3. Koordinator advances to YK review → status: yk_ön_incelemede
  4. All YK members vote approve → status: ön_onaylandı
  5. Koordinator adds 10 rep contacts → status: itibar_taramasında
  6. All 10 respond clean → status: itibar_temiz
  7. Koordinator adds to agenda → status: gündemde
  8. All YK members vote approve in yk_final → status: kabul
  9. Admin records publish consent (yes) → is_published: true
  10. GET /members → applicant appears in list

  SCENARIO: Asil Üye — RED at Reference Stage
  1. Submit application with 3 refs
  2. One ref responds negative
  3. Assert: status = reddedildi
  4. Assert: rejection_reason set and immutable
  5. Assert: applicant receives rejection email (no reason)
  6. Assert: cannot cast vote on this application (ErrApplicationTerminated)

  SCENARIO: Repeat Applicant
  1. Complete scenario above (rejected)
  2. Submit new application with same email
  3. Assert: YK sees red history banner in admin UI
  4. Assert: previous_app_id linked

  SCENARIO: Onursal — YİK Veto
  1. asil_uye proposes honorary member
  2. YK prelim: all approve
  3. YİK member submits negative
  4. Assert: status = reddedildi
  5. Assert: is_veto = true on that vote

VERIFICATION CHECKLIST:
- [ ] All 13 migrations ran successfully
- [ ] All API endpoints return correct status codes
- [ ] All email templates render in MailDev
- [ ] Token expiry returns 410
- [ ] Token reuse returns 409
- [ ] RED decisions are immutable in DB
- [ ] Logs are append-only
- [ ] Role guards enforced on all endpoints
- [ ] Repeat applicant detection works
- [ ] RED voter identity hidden from non-YK
- [ ] Public member list sorted alphabetically
- [ ] Admin UI renders all pages without errors
- [ ] All 3 voting stages work correctly
- [ ] Mobile responsive layout on public token pages
```

---

## 📌 PHASE COMPLETION CHECKLIST

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Infrastructure & Monorepo | ⬜ |
| 2 | Database Migrations & Models | ⬜ |
| 3 | Authentication & RBAC | ⬜ |
| 4 | Application Module & State Machine | ⬜ |
| 5 | Reference System & Token Engine | ⬜ |
| 6 | Consultation System | ⬜ |
| 7 | Email Notification System | ⬜ |
| 8 | Reputation System | ⬜ |
| 9 | Voting Module | ⬜ |
| 10 | Honorary Membership Flow | ⬜ |
| 11 | Web Publish Consent | ✅ |
| 12 | Admin UI (Complete) | ⬜ |
| 13 | Audit, Security & Final Polish | ⬜ |

---

*Always paste the SESSION HEADER before each phase prompt.*
*Always read INSTRUCTIONS.md before starting a new session.*