# Product Requirements Document for Ascentio (MVP)

## 1. Product Overview

Ascentio is an automated system designed to streamline the job search process. It identifies relevant job opportunities matching a user's profile, evaluates the match quality using AI, and generates tailored cover letter drafts to aid the job application process. The MVP will focus on manual job entry, core matching functionality, and basic document generation. The system presents prioritized job matches and associated materials via a dashboard, allowing users to efficiently review opportunities and apply manually.

## 2. Business Requirements

### 2.1 Key Objectives (MVP)

* Enable manual job entry through admin UI and API endpoints
* Intelligently match job requirements with user skills/experience using API-based LLMs (OpenAI, Claude, Gemini)
* Generate tailored cover letter drafts as PDFs to assist the application process
* Provide a clear dashboard for reviewing matched jobs, match quality, and generated materials
* Enable users to efficiently manage and track job opportunities they're interested in
* Operate efficiently on minimal infrastructure (single VPS) and be fully configurable through environment variables and mounted files
* Support self-hosting via Docker Compose

### 2.2 Target Users

* Job seekers looking to maximize their application efficiency
* Professionals interested in passively monitoring relevant job openings
* Career transitioners wanting to explore opportunities in new fields
* Users comfortable with self-hosting via Docker

## 3. Functional Requirements

### 3.1 Job Management

* **Manual Job Entry:**
  * Admin UI form for manually adding job listings
  * API endpoint for submitting job details programmatically
  * Support for future scraper integrations via the same API
* **Job Data Processing:**
  * Store key information (title, company, location, description, requirements, URL)
  * Detect and filter duplicate listings based on URL

### 3.2 Profile Management

* **Resume Analysis:**
  * Parse and extract skills, experience, and qualifications from an uploaded resume file
  * Support for a single primary resume
* **Preference Configuration:**
  * Target roles, locations, keywords defined via config files
  * All settings configurable through mounted files and environment variables

### 3.3 Job Matching

* **AI-Powered Evaluation:**
  * Use API-based LLMs to assess job-resume compatibility
  * Generate match scores (0-100%) with confidence ratings
  * Provide reasoning/explanation for match assessments
* **Filtering and Prioritization:**
  * Rank jobs based on match quality
  * Filter out poor matches based on a configurable threshold

### 3.4 Application Assistance

* **Cover Letter Generation:**
  * Generate tailored cover letter draft using selected LLM
  * Use user-provided writing samples for style reference
  * Save generated cover letter as downloadable PDF
  * Associate generated PDF with corresponding job match
* **Apply Links:**
  * Store and provide the original job posting URL for manual application

### 3.5 Admin Dashboard (UI)

* **Job Management:**
  * List view of all jobs with filtering options
  * Form for manually adding new jobs
  * View/edit job details
* **Match Results:**
  * Overview of matched jobs with scores and explanations
  * Access to download generated cover letter PDFs
  * Direct links to original job postings
* **Status Management:**
  * Allow marking jobs (e.g., "Interested", "Applied", "Not Interested")
  * Basic status tracking and filtering

## 4. Non-Functional Requirements

### 4.1 Performance

* Admin UI dashboard loads within 2 seconds
* Job matching and cover letter generation complete within 30 seconds
* Support for at least 1000 stored jobs with good UI performance

### 4.2 Security

* Secure storage of API keys for LLM services
* Basic authentication for admin dashboard
* HTTPS support for production deployments

### 4.3 Reliability

* High success rate for AI matching and cover letter generation tasks
* Graceful error handling with clear error messages
* Comprehensive logging for troubleshooting

### 4.4 Configuration

* All settings configurable via environment variables
* Support for mounted configuration files and assets
* Documentation for all configuration options

## 5. User Experience

### 5.1 Setup & Configuration

* Deployment via Docker Compose
* Configuration through environment variables and mounted files
* Clear documentation on required file formats

### 5.2 Admin Dashboard

* Clean, responsive web interface
* Intuitive job entry form
* Job listing with sort/filter functionality
* Match details view with explanations
* PDF preview and download functionality

### 5.3 Automation Flow

* Scheduled, configurable execution of matching and document generation
* Optional notification when new matches are found (future)

## 6. Technical Constraints

* Must operate efficiently on a single VPS (1-2 CPU, 2GB RAM)
* Containerized deployment using Docker Compose
* Persistent storage via mounted volumes
* Go-based implementation

## 7. Success Metrics

* User-perceived quality of job matches
* Quality and usefulness of generated cover letter drafts
* Time saved in finding relevant opportunities
* Successful setup rate by self-hosting users

## 8. Future Expansion (Post-MVP)

* Automated job scraping from platforms (LinkedIn, Indeed)
* Email notifications for new matches
* Enhanced analytics on job requirements and skill gaps
* Multiple resume support
* Browser extension for quick job capture
