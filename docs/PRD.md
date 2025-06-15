# Product Requirements Document for Ascentio (MVP)

## 1. Product Overview

Ascentio is a job prospecting and management application designed to help users organize their job search process. The current implementation provides a solid foundation with user authentication, job management capabilities, and a modern web interface. The application emphasizes simplicity, self-hosting capabilities, and clean architecture principles. Users can manage job postings, track application status, and maintain their job search workflow through an intuitive web dashboard.

## 2. Business Requirements

### 2.1 Key Objectives (Current State)

* Provide secure user authentication with multiple options (Google OAuth, username/password)
* Enable job management through intuitive web UI and REST API endpoints
* Offer a responsive, modern web interface for job tracking and management
* Support efficient job organization and application status tracking
* Operate efficiently on minimal infrastructure (single VPS) with environment-based configuration
* Maintain clean, maintainable codebase with comprehensive testing
* Support self-hosting via Docker Compose with minimal setup requirements
* Provide foundation for future AI-powered features and automation

### 2.2 Target Users

* Job seekers who want to organize and track their job applications systematically
* Professionals who prefer self-hosted solutions for privacy and control
* Developers and technical users comfortable with Docker-based deployments
* Users who value clean, simple interfaces over complex feature sets
* Career changers who need to manage multiple job opportunities across different fields

## 3. Functional Requirements

### 3.1 Authentication & User Management

* **Google OAuth Integration:**
  * Seamless registration and login via Google accounts
  * Secure token-based authentication
* **Username/Password Authentication:**
  * Traditional login support for users who prefer it
  * Secure password hashing and session management
* **Environment-based Admin Creation:**
  * Automatic admin user creation on first startup
  * Configurable via environment variables

### 3.2 Job Management

* **Job CRUD Operations:**
  * Create, read, update, and delete job postings
  * Web form interface for manual job entry
  * REST API endpoints for programmatic access
* **Job Data Storage:**
  * Store essential information (title, company, location, description, URL, salary)
  * Application status tracking (applied, interested, rejected, etc.)
  * Notes and custom fields for additional context

### 3.3 User Interface

* **Responsive Web Dashboard:**
  * Modern, mobile-friendly interface built with Tailwind CSS
  * HTMX-powered interactive elements for smooth user experience
  * Clean, intuitive navigation and layout
* **Job Listing Views:**
  * Sortable and filterable job tables
  * Detailed job view with all information
  * Status management and bulk operations

### 3.4 Settings & Configuration

* **User Preferences:**
  * Customizable application settings
  * User profile management
* **System Configuration:**
  * Environment-based configuration system
  * Database connection and authentication settings

### 3.5 API Endpoints

* **Job Management API:**
  * GET /api/jobs - List all jobs with filtering
  * POST /api/jobs - Create new job posting
  * GET /api/jobs/{id} - Get specific job details
  * PATCH /api/jobs/{id} - Update job information
  * DELETE /api/jobs/{id} - Remove job posting
* **Authentication API:**
  * POST /api/auth/login - Username/password authentication
  * POST /api/auth/refresh - Token refresh
  * GET /auth/google - Google OAuth initiation
  * GET /auth/google/callback - OAuth callback handling

## 4. Non-Functional Requirements

### 4.1 Performance

* Web dashboard loads within 2 seconds on typical hardware
* API responses complete within 1 second for standard operations
* Support for thousands of stored jobs with efficient pagination
* Optimized SQLite queries with proper indexing

### 4.2 Security

* Secure JWT token-based authentication
* bcrypt password hashing with appropriate cost factors
* GDPR-compliant logging with hashed user identifiers
* Environment-based secret management
* HTTPS support for production deployments

### 4.3 Reliability

* Comprehensive error handling with user-friendly messages
* Automatic database migrations on startup
* Graceful degradation when external services are unavailable
* Extensive test coverage for critical functionality

### 4.4 Configuration

* Complete environment variable-based configuration
* Secure secret management
* Docker Compose deployment with minimal configuration
* Clear documentation for all settings

## 5. User Experience

### 5.1 Setup & Configuration

* One-command deployment via Docker Compose (`make run`)
* Environment variable-based configuration with sensible defaults
* Automatic admin user creation on first startup
* Clear documentation and examples provided

### 5.2 Web Interface

* Clean, modern interface with Tailwind CSS styling
* Responsive design that works on desktop and mobile
* Intuitive navigation and user-friendly forms
* Interactive elements powered by HTMX for smooth experience
* Accessible design following web standards

### 5.3 Authentication Flow

* Multiple authentication options (Google OAuth recommended)
* Seamless registration process for new users
* Secure session management with automatic token refresh
* Clear error messages and user feedback

## 6. Technical Constraints

* Must operate efficiently on a single VPS (1-2 CPU, 2GB RAM)
* Containerized deployment using Docker Compose
* SQLite database for simplicity and ease of backup
* Go-based implementation with minimal external dependencies
* Stateless application design for easy scaling

## 7. Success Metrics

* User satisfaction with the job tracking and management interface
* Time saved in organizing job search activities
* Successful deployment rate by self-hosting users
* Application stability and uptime
* Code quality and maintainability metrics

## 8. Current Architecture Status

### 8.1 Implemented Features ✅

* Complete authentication system with Google OAuth and username/password
* Job management CRUD operations via web UI and API
* Responsive web interface with modern design
* Environment-based configuration system
* Automated database migrations
* Comprehensive test coverage
* GDPR-compliant logging system
* Health monitoring endpoints

### 8.2 Foundation for Future Features

* Modular architecture ready for AI integration
* RESTful API design for external tool integration
* Extensible job data model
* Plugin-ready configuration system

## 9. Future Expansion Roadmap

### Phase 1: AI Integration

* LLM-based job matching and scoring
* Automated cover letter generation
* Resume parsing and skill extraction

### Phase 2: Automation

* Browser extension for job capture
* Email notifications and alerts
* Scheduled processing workflows

### Phase 3: Advanced Features

* Analytics dashboard with insights
