# Vega - Navigate Your Career Journey

![Build Status](https://github.com/benidevo/vega/workflows/CI/badge.svg)

## Overview

Like ancient navigators used the star Vega to find their way, Vega helps you navigate your career journey with confidence. Vega is an AI-powered job search platform that combines intelligent job matching, automated cover letter generation, and comprehensive application tracking to help you reach your professional destination.

## ✨ Features

* **⭐ Intelligent Career Navigation**: AI-powered job matching and personalized guidance
* **🤖 AI Cover Letter Generation**: Automatically create tailored cover letters for each opportunity
* **📊 Smart Job Matching**: AI analyzes job requirements against your profile for compatibility scoring
* **🔗 Browser Extension**: One-click job capture from LinkedIn with automated data extraction
* **🗺️ Application Pipeline Tracking**: Visualize your journey from "Interested" to "Offer Received"
* **🔐 Secure Authentication**: Google OAuth and username/password options
* **📋 Comprehensive Profile Builder**: Education, experience, skills, and certifications management
* **🎨 Modern UI**: Sleek, dark-themed interface with star navigation metaphors
* **🔧 Easy Setup**: One-command deployment with Docker Compose
* **🛡️ Privacy-First**: Self-hosted solution with GDPR-compliant logging
* **⚡ Fast & Lightweight**: Efficient SQLite database with optimized performance
* **🔌 API-Ready**: RESTful endpoints for automation and integrations

## Technology Stack

Built with Go, SQLite, and modern web technologies. See [Technical Design Document](docs/TDD.md) for complete technology stack details.

## 🚀 Quick Start

### Prerequisites

* Docker
* Docker Compose
* Google Gemini API key (for AI features)

### Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/benidevo/vega.git
   cd vega
   ```

2. **Configure environment**:

   ```bash
   cp .env.example .env
   # Edit .env with your settings:
   # - Google OAuth credentials (for social login)
   # - Google Gemini API key (for AI features)
   # - Admin credentials
   # - Other configuration options
   ```

3. **Start the application**:

   ```bash
   make run
   ```

4. **Access the application**:
   * Open <http://localhost:8000> in your browser
   * Create an admin user or sign in with Google OAuth

5. **Install the browser extension** (optional):
   * Download from the [Vega Extension repository](https://github.com/benidevo/vega-extension)
   * Load the extension in Chrome or Edge for one-click job capture from LinkedIn

### 🔑 Admin User Setup

For first-time setup, you can create an admin user automatically:

```bash
# Set environment variables in .env or export directly:
export CREATE_ADMIN_USER=true
export ADMIN_USERNAME=admin
export ADMIN_PASSWORD=your_secure_password
export ADMIN_EMAIL=admin@example.com

# Restart the application
make run
```

## 🛠️ Development Commands

| Command | Description |
|---------|-------------|
| `make run` | Build and start the application containers |
| `make restart` | Rebuild and restart containers |
| `make test` | Run the full test suite with coverage |
| `make test-verbose` | Run tests with verbose output |
| `make stop` | Stop the application containers |
| `make logs` | View container logs (last 100 lines, follow) |
| `make enter-app` | Open shell inside the running container |
| `make format` | Format Go code and run linters |

## 🗄️ Database Management

| Command | Description |
|---------|-------------|
| `make migrate-create` | Create a new migration file |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Rollback the most recent migration |
| `make migrate-reset` | Rollback all migrations |
| `make migrate-force` | Set migration to specific version |

## 📖 Configuration

Copy `.env.example` to `.env` and configure Google OAuth credentials. See [Technical Design Document](docs/TDD.md) for complete configuration reference.

## 🧪 Development & Testing

### Database Migrations

Migrations are automatically applied on startup. To work with migrations manually:

```bash
# Create new migration
make migrate-create
# Enter name when prompted

# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down
```

### Code Quality

```bash
# Set up Git hooks for code quality
make setup-hooks

# Format code and run linters
make format

# Run tests with coverage
make test
```

## 📊 API Documentation

Vega provides RESTful APIs for job management, authentication, and system health. See [Technical Design Document](docs/TDD.md) for complete API reference.

## 🔮 Roadmap

### Current Status ✅

* [x] User authentication (Google OAuth + username/password)
* [x] Job management with status tracking
* [x] AI-powered job matching with Google Gemini
* [x] Automated cover letter generation
* [x] Browser extension for LinkedIn job capture
* [x] Profile management (education, experience, skills, certifications)
* [x] Responsive, dark-themed web interface
* [x] RESTful API endpoints
* [x] Environment-based configuration
* [x] Automated database migrations
* [x] GDPR-compliant logging
* [x] Application pipeline visualization

### Future Plans 🚧

* [ ] Support for additional job boards (Indeed, AngelList, etc.)
* [ ] Email notifications and alerts
* [ ] Advanced analytics and insights dashboard
* [ ] Interview preparation tools
* [ ] Resume parsing and optimization
* [ ] Calendar integration for interview scheduling
* [ ] Mobile application

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run the test suite: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
