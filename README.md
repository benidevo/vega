# Vega

![Build Status](https://github.com/benidevo/vega/workflows/CI/badge.svg)

## Overview

Vega is a self-hosted job prospecting and management application designed to help users organize their job search process. Built with simplicity and privacy in mind, it provides a clean web interface for tracking job applications, managing opportunities, and maintaining your job search workflow.

## ✨ Features

* **🤖 AI-Powered Matching**: Google Gemini integration for intelligent job-resume compatibility scoring
* **🔗 Browser Extension**: One-click job capture from LinkedIn with automated data extraction
* **🔐 Secure Authentication**: Google OAuth and username/password options
* **📋 Job Management**: Create, edit, and organize job postings with detailed tracking
* **🎨 Modern UI**: Responsive design with Tailwind CSS and HTMX interactivity
* **🔧 Easy Setup**: One-command deployment with Docker Compose
* **🛡️ Privacy-First**: Self-hosted with GDPR-compliant logging
* **⚡ Fast & Lightweight**: Efficient SQLite database with optimized performance
* **🔌 API-Ready**: RESTful endpoints for automation and integrations

## Technology Stack

Built with Go, SQLite, and modern web technologies. See [Technical Design Document](docs/TDD.md) for complete technology stack details.

## 🚀 Quick Start

### Prerequisites

* Docker
* Docker Compose

### Installation

1. **Clone the repository**:

   ```bash
   git clone https://github.com/benidevo/vega.git
   cd vega
   ```

2. **Configure environment**:

   ```bash
   cp .env.example .env
   # Edit .env with your settings (Google OAuth, admin credentials, etc.)
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
   * Load the extension in Chrome for one-click job capture from LinkedIn

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
* [x] Job management CRUD operations
* [x] Browser extension for LinkedIn job capture
* [x] Responsive web interface
* [x] RESTful API endpoints
* [x] Environment-based configuration
* [x] Automated migrations
* [x] GDPR-compliant logging

### Future Plans 🚧

* [ ] AI-powered job matching with Google Gemini
* [ ] Automated cover letter generation
* [ ] Support for additional job boards (Indeed, etc.)
* [ ] Email notifications and alerts
* [ ] Analytics and insights dashboard
* [ ] Advanced AI scoring and recommendations

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
