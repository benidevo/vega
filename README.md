# Vega AI

Vega AI helps you navigate your career journey with AI-powered job search tools. Track job applications, generate tailored cover letters and documents using AI, get intelligent job matching based on your profile, and capture opportunities from LinkedIn with the browser extension. Self-hosted for complete data privacy.

## Quick Start

**1. Create environment file (`.env.prod`):**

```bash
# Required
TOKEN_SECRET=your-secure-jwt-secret-here
GEMINI_API_KEY=your-gemini-api-key

# Optional - Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Recommended - First-time admin setup  
CREATE_ADMIN_USER=true
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your-admin-password
ADMIN_EMAIL=admin@yourdomain.com
```

**2. Run with Docker:**

```bash
docker run -d \
  --name vega-ai \
  -p 8080:8080 \
  -v vega-data:/app/data \
  --env-file .env.prod \
  ghcr.io/benidevo/vega-ai:latest
```

**3. Access:** <http://localhost:8080>

That's it! 🚀

## Features

* **🤖 AI Document Generation**: Automatically create tailored application documents
* **📊 Smart Job Matching**: AI analyzes job requirements vs your profile
* **🗺️ Application Tracking**: Visualize your pipeline from "Interested" to "Offer"
* **🔗 Browser Extension**: One-click job capture from LinkedIn
* **🔐 Secure Auth**: Google OAuth + local accounts
* **⚡ Self-Hosted**: Your data stays with you

## Development

**Prerequisites:** Docker, Docker Compose, [Gemini API key](https://ai.google.dev/)

```bash
git clone https://github.com/benidevo/vega-ai.git
cd vega-ai
cp .env.example .env
# Edit .env with your API keys
make run
```

Access at <http://localhost:8000>

## Browser Extension

Install the [Vega AI Extension](https://github.com/benidevo/vega-ai-extension) for one-click job capture from LinkedIn.

## Development Commands

| Command | Description |
|---------|-------------|
| `make run` | Start development environment |
| `make test` | Run test suite |
| `make restart` | Rebuild and restart |
| `make logs` | View container logs |

## Production Build

```bash
# Build locally
./scripts/docker-build.sh

# Build and push
./scripts/docker-build.sh --tag v1.0.0 --push
```

Images are built automatically on version tags and can be manually triggered via GitHub Actions.

## License

MIT License - see [LICENSE](LICENSE) file.
