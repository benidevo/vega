# Technical Design Document for ProspecTor (MVP)

## 1. System Architecture Overview

ProspecTor MVP uses a **monolithic architecture** implemented entirely in Go, with a SQLite database for persistence. The application is containerized using Docker and configured via environment variables and mounted files.

### 1.1 High-Level Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                  VPS Host                                    │
│                            (via Docker Compose)                              │
│                                                                              │
│  ┌─────────────────────────────┐                                             │
│  │      Go Application         │                                             │
│  │---------------------------  │                                             │
│  │ - Admin UI & API            │                                             │
│  │ - Job Management            │                                             │
│  │ - LLM Client(s)             │                                             │
│  │ - Job Matcher               │                                             │
│  │ - Cover Letter Generator    │                                             │
│  │ - PDF Creator               │                                             │
│  │ - Authentication            │                                             │
│  └─────────────┬───────────────┘                                             │
│                │                                                             │
│                ▼                                                             │
│  ┌─────────────┴───────────────┐      ┌─────────────────────────────────┐    │
│  │     SQLite (Data)           │      │     Scheduled Processes (Cron)  │    │
│  │                             │      │                                 │    │
│  └─────────────────────────────┘      └─────────────────────────────────┘    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Component Responsibilities

1. **Admin UI & API**
   * Web interface for job management and results viewing
   * REST API endpoints for programmatic job submission
   * Authentication and session management

2. **Job Management**
   * Job CRUD operations
   * Deduplication logic
   * Status tracking

3. **LLM Clients**
   * Interface for LLM provider interactions
   * Implementations for OpenAI, Claude, and Gemini
   * Error handling and retries

4. **Job Matcher**
   * Resume parsing and skill extraction
   * Job description analysis
   * Match score calculation
   * Match explanation generation

5. **Cover Letter Generator**
   * Content generation based on job and resume
   * PDF formatting and creation
   * Storage and retrieval

6. **Scheduled Processes**
   * Job matching execution
   * Document generation execution
   * Run via cron as configured

## 2. Technology Stack

### 2.1 Core Technologies

* **Language**: Go 1.23+
* **Web Framework**: Gin
* **Database**: SQLite3
* **AI/ML**: API clients for OpenAI, Claude, and Gemini
* **PDF Generation**: `jung-kurt/gofpdf` or similar
* **UI**: Go templates + HTMX + Tailwind CSS (minimal JavaScript)
* **Scheduling**: Cron (via separate container)

### 2.2 Development & Deployment

* **Containerization**: Docker
* **Orchestration**: Docker Compose
* **Configuration**: Environment variables + mounted files

## 3. API Endpoints

### 3.1 Job Management API

```
POST /api/jobs
- Add a new job posting
- Body: Job details (title, company, description, etc.)
- Returns: Created job with ID

GET /api/jobs
- Get list of jobs with optional filters
- Query params: status, search, page, limit
- Returns: Paginated job list

GET /api/jobs/{id}
- Get details for a specific job
- Returns: Complete job details

PATCH /api/jobs/{id}
- Update job details or status
- Body: Fields to update
- Returns: Updated job

DELETE /api/jobs/{id}
- Remove a job from the system
- Returns: Success confirmation
```

### 3.2 Match Management API

```
GET /api/matches
- List job matches with optional filters
- Query params: minScore, status, jobId
- Returns: Paginated matches with scores

GET /api/matches/{id}
- Get details for a specific match
- Returns: Match details with explanation

GET /api/matches/{id}/cover-letter
- Download generated cover letter PDF
- Returns: PDF file
```

### 3.3 Process Trigger API

```
POST /api/processes/match
- Trigger job matching process
- Query params: jobId (optional)
- Returns: Process run ID and status

POST /api/processes/generate
- Trigger cover letter generation
- Query params: matchId (optional)
- Returns: Process run ID and status
```

## 4. Configuration Architecture

### 4.1 Environment Variables

```
# Database
DB_PATH=/app/data/prospecTor.db

# Authentication
ADMIN_USERNAME=admin
ADMIN_PASSWORD=secure_password
JWT_SECRET=random_secure_string

# LLM Configuration
DEFAULT_LLM_PROVIDER=openai
OPENAI_API_KEY=sk-...
CLAUDE_API_KEY=sk-...
GEMINI_API_KEY=...

# Application Settings
PORT=8080
LOG_LEVEL=info
BASE_URL=http://localhost:8080
```

### 4.2 Mounted Volumes & Files

```
/app/config/
  ├── profile.json       # User profile configuration
  ├── search_settings.json  # Job search preferences
  └── llm_settings.json  # LLM-specific configuration

/app/assets/
  ├── resume.pdf         # User's resume
  └── writing_samples/   # Cover letter samples

/app/data/
  ├── generated/         # Generated cover letters
  └── uploads/           # Temporary upload storage
```

## 5. Scheduled Processes

The scheduler container will run these processes based on the crontab generated from configuration:

1. **Job Matcher** (`/app/matcher`)
   * Find unmatched jobs and compare to resume
   * Generate match scores and explanations
   * Update database with results

2. **Cover Letter Generator** (`/app/generator`)
   * Find matched jobs without cover letters
   * Generate cover letter content using LLM
   * Create PDF files and store them
   * Update database with file paths
