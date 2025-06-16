# Roadmap

## Release Timeline

### Phase 1: Docker Foundation (v0.1.x)

**June - July 2025**

- Docker Compose deployment for technical users
- Core features: Job management, authentication, web UI
- RESTful APIs and SQLite database
- Browser extension for job capture
- Documentation and community building

### Phase 2: Native Desktop (v0.2.x)

**August - September 2025**

- Native apps for Windows, macOS, Linux using Wails
- Single executable (~40-50MB), no Docker required
- System tray, native menus, embedded web UI
- Migration tool for Docker users

### Phase 3: Production Ready (v1.0.0)

**October - November 2025**

- Professional installers with code signing
- Auto-update system
- First-run setup wizard
- Performance optimizations and polish

### Phase 4: Local AI (v1.1.x)

**December 2025 - January 2026**

- Ollama integration for offline AI
- Mistral 7B as default local model
- Seamless switching between Gemini and local models
- Resource monitoring and management

## Key Benefits

**For Users:**

- Evolution from technical-only (Docker) to one-click installation
- Maintains privacy-first approach throughout
- Progressive feature additions without breaking changes

**For Development:**

- Single codebase across all platforms
- Existing Go backend and web UI remain unchanged
- Dual support: Docker for development, native for distribution
