# Development Guide

This guide covers development setup, testing, and contributing to Vega AI.

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Go 1.24+](https://golang.org/) (for local development)
- [Google Gemini API key](https://aistudio.google.com/app/apikey)

## Development Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/benidevo/vega-ai.git
   cd vega-ai
   ```

2. **Set up environment:**

   ```bash
   cp .env.example .env
   # Edit .env with your API keys and configuration
   ```

3. **Start development environment:**

   ```bash
   make run
   ```

4. **Access the application:**
   - Web interface: <http://localhost:8765>
   - The application will auto-reload on code changes

## Development Commands

| Command | Description |
|---------|-------------|
| `make run` | Start development environment |
| `make test` | Run test suite |
| `make restart` | Rebuild and restart containers |
| `make logs` | View container logs |
| `make clean` | Stop and remove containers |

## Frontend Development

### CSS Build Process

The project uses Tailwind CSS with a local build process for offline functionality:

```bash
# Rebuild CSS after template changes
./tailwindcss -i ./src/input.css -o ./static/css/tailwind.min.css --minify

# Watch for changes (development)
./tailwindcss -i ./src/input.css -o ./static/css/tailwind.min.css --watch
```

**Note:** The `tailwindcss` binary is downloaded automatically during setup and should not be committed to the repository.

### Template Structure

- `templates/layouts/base.html` - Base layout with CSS/JS includes
- `templates/partials/` - Reusable template components
- `static/` - Static assets (CSS, images, etc.)

## Testing

### Run Tests

```bash
# All tests
make test

# Specific package
go test ./internal/auth/...

# With coverage
go test -cover ./...
```

### Test Structure

- Unit tests: `*_test.go` files alongside source code
- Integration tests: `internal/*/setup_test.go`
- Test coverage reports: `coverage.out`

## Production Build

### Local Build

```bash
# Build Docker image locally
./scripts/docker-build.sh

# Build with specific tag and push
./scripts/docker-build.sh --tag v1.0.0 --push
```

### CI/CD

- Automated builds on GitHub Actions
- Images published to GitHub Container Registry
- Builds triggered on version tags and manual dispatch

## Project Structure

```
├── cmd/vega/           # Application entry point
├── internal/           # Private application code
│   ├── auth/          # Authentication & authorization
│   ├── job/           # Job management
│   ├── ai/            # AI integration (Gemini)
│   ├── settings/      # User settings & profiles
│   └── common/        # Shared utilities
├── templates/         # HTML templates
├── static/           # Static assets
├── migrations/       # Database migrations
├── docker/           # Docker configurations
├── scripts/          # Build and utility scripts
└── docs/             # Documentation
```

## API Documentation

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `TOKEN_SECRET` | Yes | JWT secret for user sessions |
| `GEMINI_API_KEY` | Yes | Google AI API key |
| `GOOGLE_CLIENT_ID` | No | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | No | Google OAuth client secret |
| `CREATE_ADMIN_USER` | No | Create admin user on startup |
| `ADMIN_USERNAME` | No | Admin username (if creating) |
| `ADMIN_PASSWORD` | No | Admin password (if creating) |

### Database

- SQLite for development and small deployments
- Automatic migrations on startup
- Schema in `migrations/sqlite/`

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Add tests for new functionality
5. Ensure tests pass: `make test`
6. Commit your changes: `git commit -m 'Add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

### Code Style

- Follow Go conventions and `gofmt`
- Use meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused
- Write tests for new functionality

### Commit Messages

- Use present tense: "Add feature" not "Added feature"
- Use imperative mood: "Move cursor to..." not "Moves cursor to..."
- Limit first line to 72 characters
- Reference issues and pull requests when applicable

## Troubleshooting

### Common Issues

1. **Port already in use:**

   ```bash
   # Stop conflicting services
   docker ps
   docker stop <container-id>
   ```

2. **Database locked:**

   ```bash
   # Restart the application
   make restart
   ```

3. **Permission denied:**

   ```bash
   # Check file permissions
   ls -la data/
   # Fix if needed
   sudo chown -R $USER:$USER data/
   ```

### Logs

```bash
# View application logs
make logs

# Follow logs in real-time
docker compose logs -f vega-ai
```
