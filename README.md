# Membership Management System

## Phase 1: Infrastructure & Monorepo Setup ✅

The foundational infrastructure has been successfully set up.

### What's Been Created

#### Root-Level Files
- ✅ `package.json` - Monorepo workspace configuration
- ✅ `turbo.json` - Turborepo build pipeline
- ✅ `.gitignore` - Git ignore rules
- ✅ `.env.example` - Environment variable template
- ✅ `.env` - Local environment configuration
- ✅ `docker-compose.yml` - Docker services (MySQL, MailDev, API, Admin)

#### Go Backend (`apps/api/`)
- ✅ `go.mod` - Go module with all dependencies
- ✅ `go.sum` - Dependency checksums
- ✅ `Dockerfile` - Multi-stage build
- ✅ `cmd/main.go` - Main application entry point with Fiber v2
- ✅ `config/config.go` - Viper-based configuration loader
- ✅ `config/database.go` - GORM MySQL connection setup
- ✅ Health check endpoint: `GET /api/v1/health`

#### Next.js Frontend (`apps/admin/`)
- ✅ Next.js 14 with App Router
- ✅ TypeScript configured (strict mode)
- ✅ Tailwind CSS installed
- ✅ shadcn/ui initialized with default settings
- ✅ Dependencies installed:
  - @tanstack/react-query v5
  - @tanstack/react-table v8
  - react-hook-form
  - @hookform/resolvers
  - zod
  - axios
  - zustand
  - lucide-react
  - date-fns
- ✅ `Dockerfile` - Production build configuration
- ✅ `src/lib/api/client.ts` - Axios instance with auth interceptors

#### Shared Libraries
- ✅ `libs/shared-types/` - TypeScript type definitions
  - All enums: MembershipType, ApplicationStatus, UserRole, VoteStage, VoteType
  - All entity interfaces: User, Application, Reference, Vote, Log, etc.
  - ApiResponse generic interface
- ✅ `libs/validators/` - Zod validation schemas (placeholder)
- ✅ `libs/ui/` - shadcn/ui component library (placeholder)

### Verification Steps

#### Without Docker (Recommended for Development)

**1. Install Dependencies**
```bash
# Root-level dependencies
npm install

# Go dependencies
cd apps/api && go mod tidy
```

**2. Start Next.js Admin**
```bash
cd apps/admin
npm run dev

# Access in browser
open http://localhost:3000
```

**3. Start Go API** (requires MySQL running separately)
```bash
cd apps/api
go run cmd/main.go

# Test health endpoint
curl http://localhost:8080/api/v1/health
# Expected: {"status":"ok","version":"1.0.0","env":"development"}
```

#### With Docker (Production-like Environment)

**⚠️ Note**: MySQL and API services have port conflicts in Docker. Currently working: Admin (port 3000) and MailDev (port 1080/1025).

```bash
# Start admin and maildev services only
docker compose up admin maildev -d

# Access admin UI
open http://localhost:3000

# Access MailDev
open http://localhost:1080

# Check logs
docker compose logs -f admin
docker compose logs -f maildev
```

**Current Status**: ✅ Next.js Admin and ✅ MailDev running in Docker
**Known Issue**: MySQL port 3306 conflict prevents full stack from running (API depends on MySQL)

### Known Issues / Notes

1. **Docker Daemon**: Docker Desktop must be running for `docker compose` commands
2. **Go Build**: If you see "no such file or directory" errors, run from the correct directory:
   ```bash
   cd apps/api/cmd && go run main.go
   ```
3. **Config Files**: There were some initial file corruption issues that have been resolved. All config files are now clean.

### Environment Variables

All required environment variables are in [.env.example](.env.example). Copy to `.env` and adjust as needed:

```bash
cp .env.example .env
```

Key variables:
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASS` - MySQL connection
- `JWT_SECRET`, `JWT_REFRESH_SECRET` - JWT signing keys
- `APP_PORT` - API server port (default: 8080)
- `MAIL_HOST`, `MAIL_PORT` - SMTP settings (MailDev for dev)
- `NEXT_PUBLIC_API_URL` - Frontend API URL

### Next Steps

**Phase 2: Database Migrations & GORM Models**

To continue, paste the Phase 2 prompt from [COPILOT_PROMPTS.md](COPILOT_PROMPTS.md):

```
Read INSTRUCTIONS.md Section 15 (Database Schema) and Section 18 (Audit & Immutability Rules).

Create all database migrations and GORM models...
```

---

## Project Structure

```
membership-system/
├── apps/
│   ├── api/                    # Go backend (Fiber v2)
│   │   ├── cmd/main.go
│   │   ├── config/
│   │   │   ├── config.go      # Viper config loader
│   │   │   └── database.go    # GORM setup
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── admin/                  # Next.js 14 frontend
│       ├── src/
│       │   ├── app/           # App Router pages
│       │   ├── components/    # React components
│       │   └── lib/
│       │       └── api/
│       │           └── client.ts  # Axios instance
│       ├── Dockerfile
│       └── package.json
│
├── libs/
│   ├── shared-types/          # TypeScript types
│   ├── validators/            # Zod schemas
│   └── ui/                    # shadcn/ui components
│
├── docker-compose.yml
├── turbo.json
├── .env.example
└── package.json
```

## Tech Stack

### Backend
- **Language**: Go 1.22+
- **Framework**: Fiber v2
- **ORM**: GORM v2
- **Database**: MySQL 8.0
- **Auth**: JWT (golang-jwt/jwt v5)
- **Config**: Viper
- **Email**: gomail.v2
- **Validation**: go-playground/validator v10

### Frontend
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript (strict mode)
- **Forms**: React Hook Form + Zod
- **Tables**: TanStack Table v8
- **Data Fetching**: TanStack Query v5
- **State**: Zustand
- **UI**: shadcn/ui + Tailwind CSS
- **Icons**: lucide-react
- **HTTP**: Axios

### Infrastructure
- **Monorepo**: Turborepo
- **Dev Email**: MailDev
- **Containers**: Docker + docker-compose

---

**Status**: Phase 1 Complete ✅ | Ready for Phase 2
