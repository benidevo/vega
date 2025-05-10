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
* **UI Framework**:
  * Go templates with template inheritance
  * HTMX for interactive UI elements and form submissions
  * Tailwind CSS for responsive styling
  * Particles.js for background animations
  * Minimal JavaScript approach
* **Color Theme**:
  * Primary: Teal (#0D9488)
  * Secondary: Amber (#F59E0B)
  * Additional Accent Colors: Light Blue (#0EA5E9), Purple (#8B5CF6)
  * Background: Slate gradient (slate-900 via slate-800 to slate-900)
  * Text: White/Gray for readability
* **UI Layout**:
  * Responsive design with mobile-first approach
  * Glass-morphism effects (backdrop blur) for containers
  * Clean, minimalist interface with subtle animations
  * Multi-color particle backgrounds with interactive hover effects
* **Scheduling**: Cron (via separate container)

### 2.2 UI Design System

#### Common UI Components

* **Container Styles**:
  ```html
  <!-- Main container with glow effect -->
  <div class="w-full max-w-5xl z-10 relative">
    <!-- Glow effect -->
    <div class="absolute -inset-1 rounded-xl bg-gradient-to-r from-primary via-secondary to-primary opacity-30 blur-xl"></div>

    <!-- Content container -->
    <div class="relative p-8 md:p-12 bg-slate-900 bg-opacity-70 backdrop-blur-xl rounded-xl shadow-2xl border border-white border-opacity-10">
      <!-- Content goes here -->
    </div>
  </div>
  ```

* **Animated Logo**:
  ```html
  <div class="animate-float">
    <div class="p-6 bg-slate-800 bg-opacity-30 rounded-full border border-primary border-opacity-30 shadow-lg shadow-primary/20">
      <svg class="h-20 w-20 text-primary" viewBox="0 0 24 24"><!-- SVG content --></svg>
    </div>
    <div class="absolute -inset-1 bg-primary opacity-20 blur-xl rounded-full animate-pulse-slow"></div>
  </div>
  ```

* **Gradient Text Headings**:
  ```html
  <h1 class="text-5xl font-bold text-center mb-4 bg-clip-text text-transparent bg-gradient-to-r from-primary to-secondary">
    Heading Text
  </h1>
  ```

* **Feature Cards**:
  ```html
  <div class="group p-6 bg-slate-800 bg-opacity-50 rounded-xl border border-slate-700 hover:border-primary transition-all duration-300 transform hover:-translate-y-1 hover:shadow-lg hover:shadow-primary/20">
    <div class="flex items-center justify-center mb-4">
      <div class="p-3 bg-slate-700 bg-opacity-50 rounded-lg group-hover:bg-primary group-hover:bg-opacity-20 transition-all duration-300">
        <svg class="h-10 w-10 text-secondary group-hover:text-primary transition-colors duration-300">
          <!-- SVG content -->
        </svg>
      </div>
    </div>
    <h3 class="text-lg font-semibold text-center text-white mb-3">Feature Title</h3>
    <p class="text-gray-300 text-center">Feature description text</p>
  </div>
  ```

* **Gradient Buttons**:
  ```html
  <button class="relative inline-flex items-center justify-center overflow-hidden font-bold rounded-lg group">
    <span class="absolute w-full h-full bg-gradient-to-br from-primary to-secondary opacity-70 group-hover:opacity-100 transition-opacity duration-300"></span>
    <span class="relative px-8 py-4 transition-all ease-out bg-slate-900 rounded-md group-hover:bg-opacity-0 duration-300">
      <span class="relative text-white font-semibold text-lg">Button Text</span>
    </span>
  </button>
  ```

* **Form Fields with Icons**:
  ```html
  <div class="relative">
    <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
      <svg class="h-5 w-5 text-gray-400"><!-- SVG content --></svg>
    </div>
    <input type="text" class="w-full pl-10 pr-4 py-3 rounded-lg bg-slate-800 bg-opacity-50 border border-slate-700 text-white focus:outline-none focus:ring-2 focus:ring-primary focus:border-primary transition-colors">
  </div>
  ```

* **Particles.js Configuration**:
  * Use multi-color particles (Teal, Amber, Light Blue, Purple)
  * Implement interactive hover effects (grab, push)
  * Adjust opacity and animation settings for subtle background effect

* **Custom Animations**:
  ```css
  @keyframes float {
    0% { transform: translateY(0px); }
    50% { transform: translateY(-10px); }
    100% { transform: translateY(0px); }
  }

  .animate-float {
    animation: float 6s ease-in-out infinite;
  }

  @keyframes pulse-slow {
    0% { opacity: 0.8; }
    50% { opacity: 0.3; }
    100% { opacity: 0.8; }
  }

  .animate-pulse-slow {
    animation: pulse-slow 5s ease-in-out infinite;
  }
  ```

### 2.3 Development & Deployment

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
