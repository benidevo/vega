# 🌟 Vega AI

> Navigate your career journey with AI-powered precision

Just as ancient navigators used the star Vega to find their way, Vega AI helps you navigate your career journey with intelligent job search tools. Track applications, generate tailored documents using AI, get smart job matching based on your profile, and capture opportunities from LinkedIn with our browser extension. Self-hosted for complete data privacy.

---

## 🚀 Quick Start

### Option 1: Docker Run

**1. Create environment file (`.env`):**

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
  -p 8765:8765 \
  -v vega-data:/app/data \
  --env-file .env \
  ghcr.io/benidevo/vega-ai:latest
```

### Option 2: Docker Compose

**1. Create `docker-compose.yml`:**

```yaml
services:
  vega-ai:
    image: ghcr.io/benidevo/vega-ai:latest
    container_name: vega-ai
    restart: unless-stopped
    ports:
      - "8765:8765"
    volumes:
      - vega-data:/app/data
    env_file:
      - .env
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8765/health"]
      interval: 2m
      timeout: 15s
      retries: 3
      start_period: 30s

volumes:
  vega-data:
```

**2. Start the application:**

```bash
docker compose up -d
```

**3. Access:** <http://localhost:8765> 🎉

---

## ✨ Features

* **🤖 AI Document Generation**: Automatically create tailored application documents
* **📊 Smart Job Matching**: AI analyzes job requirements vs your profile
* **🗺️ Application Tracking**: Visualize your pipeline from "Interested" to "Offer"
* **🔗 Browser Extension**: One-click job capture from LinkedIn
* **🔐 Secure Auth**: Google OAuth + local accounts
* **⚡ Self-Hosted**: Your data stays with you

---

## 🛠️ Development

**Prerequisites:** Docker, Docker Compose, [Gemini API key](https://ai.google.dev/)

```bash
git clone https://github.com/benidevo/vega-ai.git
cd vega-ai
cp .env.example .env
# Edit .env with your API keys
make run
```

**Access:** <http://localhost:8765>

| Command | Description |
|---------|-------------|
| `make run` | Start development environment |
| `make test` | Run test suite |
| `make restart` | Rebuild and restart |
| `make logs` | View container logs |

---

## 🔗 Extensions

Install the [**Vega AI Extension**](https://github.com/benidevo/vega-ai-extension) for one-click job capture from LinkedIn.

---

## 🏗️ Production Build

```bash
# Build locally
./scripts/docker-build.sh

# Build and push
./scripts/docker-build.sh --tag v1.0.0 --push
```

Images are built automatically on version tags and can be manually triggered via GitHub Actions.

---

## 📄 License

MIT License - see [LICENSE](LICENSE) file.
