# 🌟 Vega AI

[![CI](https://github.com/benidevo/vega-ai/workflows/CI/badge.svg)](https://github.com/benidevo/vega-ai/actions/workflows/ci.yaml)
[![Docker](https://github.com/benidevo/vega-ai/workflows/Build%20and%20Push%20Docker%20Image/badge.svg)](https://github.com/benidevo/vega-ai/actions/workflows/docker-build.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![GitHub Container Registry](https://img.shields.io/badge/ghcr.io-vega--ai-blue)](https://github.com/benidevo/vega-ai/pkgs/container/vega-ai)

> Navigate your career journey with AI-powered precision

Just as ancient navigators used the star Vega to find their way, Vega AI helps you navigate your career journey with intelligent job search tools. Track applications, generate tailored documents using AI, get smart job matching based on your profile, and capture opportunities from LinkedIn with the browser extension.

**🚀 Try it now:** Visit [vega.benidevo.com](https://vega.benidevo.com) for the cloud version, or self-host for complete data privacy.

## 🚀 Self-Hosted Quick Start

Self-hosting Vega AI gives you complete control over your data. You only need a Gemini API key to get started.

### 1. Get Your API Key

Get your [free Gemini API key](https://aistudio.google.com/app/apikey) from Google AI Studio.

### 2. Create Configuration

```bash
# Create a directory for Vega AI
mkdir vega-ai && cd vega-ai

# Create a config file with your Gemini API key
echo "GEMINI_API_KEY=your-gemini-api-key" > config
```

That's it! No complex setup required.

### 3. Run with Docker

Start Vega AI with a single command:

```bash
docker run --pull always -d \
  --name vega-ai \
  -p 8765:8765 \
  -v vega-data:/app/data \
  --env-file config \
  ghcr.io/benidevo/vega-ai:latest
```

### 4. Access Vega AI

1. Visit <http://localhost:8765>
2. Log in with default credentials:
   - Username: `admin`
   - Password: `VegaAdmin`
3. **Important:** Change your password after first login via Settings → Account

## ✨ Features

- **🤖 AI Document Generation**: Automatically create tailored application documents
- **📊 Smart Job Matching**: AI analyzes job requirements vs your profile
- **🗺️ Application Tracking**: Visualize your pipeline from "Interested" to "Offer"
- **🔗 Browser Extension**: One-click job capture from LinkedIn
- **⚡ Self-Hosted**: Your data stays with you

## 🔗 Browser Extension

For one-click job capture from LinkedIn, checkout the [**Browser Extension**](https://github.com/benidevo/vega-ai-extension).

## 🐳 Docker Options

### ARM64 Support (Apple Silicon)

The Docker images support both AMD64 and ARM64 architectures:

```bash
# Works on both Intel/AMD and ARM processors
docker pull ghcr.io/benidevo/vega-ai:latest
```

### Docker Compose

For easier management, use Docker Compose:

```yaml
# docker-compose.yml
services:
  vega-ai:
    image: ghcr.io/benidevo/vega-ai:latest
    ports:
      - "8765:8765"
    volumes:
      - vega-data:/app/data
    env_file:
      - config
    restart: unless-stopped

volumes:
  vega-data:
```

Then run: `docker compose up -d`

### Advanced Configuration

For advanced settings (custom ports, SSL, external databases), see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

## 🛠️ Development

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for development setup, testing, and contributing guidelines.

## 📝 License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).

What this means:

- ✅ You can use, study, modify, and distribute the code
- ✅ If you run this software on a server, you must make your source code available to users
- ✅ Any modifications must also be released under AGPL-3.0

**Commercial licensing:** For commercial use without AGPL restrictions, contact <benjaminidewor@gmail.com> for licensing options.
