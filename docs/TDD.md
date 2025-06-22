# Vega AI Technical Design

Go web application for AI-powered job search and application tracking.

## Architecture

**Pattern:** Layered architecture with clean domain separation
**Database:** SQLite with migrations
**Authentication:** JWT + Google OAuth
**Frontend:** Go templates + HTMX + Tailwind CSS
**AI:** Google Gemini API integration

```plaintext
cmd/vega/           # Entry point
internal/
  ├── vega/         # App setup & routing
  ├── auth/         # Authentication domain
  ├── job/          # Job management domain
  ├── settings/     # User settings domain
  ├── ai/           # AI services (Gemini)
  └── api/          # REST handlers
templates/          # HTML templates
migrations/         # Database migrations
```

## Key Components

### **Handlers → Services → Repositories → Database**

**Handlers:** HTTP request/response, template rendering
**Services:** Business logic, AI integration
**Repositories:** Data access with interfaces
**Database:** SQLite with WAL mode

### **Authentication Flow**

1. Google OAuth or username/password
2. JWT tokens (access + refresh)
3. Cookie-based session storage
4. Middleware protection for routes

### **AI Integration**

- **JobMatcherService:** Calculates compatibility scores
- **CoverLetterGeneratorService:** Generates personalized letters
- **LLM Interface:** Pluggable AI providers (currently Gemini)

## API Endpoints

### Jobs

```plaintext
POST   /api/jobs           # Create job
GET    /api/jobs           # List jobs (with filtering)
GET    /api/jobs/{id}      # Get job details
PATCH  /api/jobs/{id}      # Update job
DELETE /api/jobs/{id}      # Delete job
```

### Authentication

```plaintext
POST   /api/auth/login     # Username/password login
POST   /api/auth/refresh   # Refresh token
POST   /api/auth/logout    # Logout
GET    /auth/google        # OAuth initiate
GET    /auth/google/callback # OAuth callback
```

### System

```plaintext
GET    /health             # Health check
GET    /health/ready       # Readiness probe
```

## Database Schema

**Core Entities:**

- `users` - Authentication, roles
- `companies` - Company information
- `jobs` - Job postings with AI match scores
- `profiles` - User experience, education, skills

**Relationships:**

- User → Profile (1:1)
- User → Jobs (1:many)
- Company → Jobs (1:many)

## Configuration

**Environment Variables:**

```bash
# App
SERVER_PORT=:8765
IS_DEVELOPMENT=true
LOG_LEVEL=info

# Database
DB_CONNECTION_STRING=/app/data/vega.db?_journal_mode=WAL
MIGRATIONS_DIR=migrations/sqlite

# Auth
TOKEN_SECRET=your-jwt-secret
ACCESS_TOKEN_EXPIRY=60
REFRESH_TOKEN_EXPIRY=168

# Google OAuth
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_CLIENT_REDIRECT_URL=http://localhost:8765/auth/google/callback

# AI
GEMINI_API_KEY=your-gemini-key
```

## Testing Strategy

**Unit Tests:** Services and repositories with mocks
**Integration Tests:** Full workflows with test database
**Test Framework:** testify + go-sqlmock
**Coverage:** Custom scripts for coverage tracking

```bash
make test           # Run all tests
```

## Deployment

**Development:**

```bash
docker-compose up
```

**Production:**

```bash
docker build -t vega .
docker run -p 8765:8765 -v ./data:/app/data vega
```

**Database Migrations:**

```bash
make migrate-up     # Apply migrations
make migrate-down   # Rollback migrations
```

## Key Dependencies

- **gin-gonic/gin** - HTTP framework
- **modernc.org/sqlite** - SQLite driver
- **golang-jwt/jwt** - JWT tokens
- **rs/zerolog** - Structured logging
- **golang-migrate/migrate** - Database migrations
- **google.golang.org/genai** - Gemini AI client

## Security Features

- bcrypt password hashing
- JWT with secure cookies
- CORS configuration
- GDPR-compliant logging
- Input validation
- SQL injection prevention

## AI Prompt Templates

**Job Matching:**

- User profile analysis
- Job requirements parsing
- Compatibility scoring (0-100)

**Cover Letter Generation:**

- User experience integration
- Job-specific customization
- Professional tone adaptation
