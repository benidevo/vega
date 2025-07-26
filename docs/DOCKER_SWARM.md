# Docker Swarm Deployment Guide

This guide explains how to deploy Vega AI using Docker Swarm mode.

## Prerequisites

- Docker Engine with Swarm mode enabled
- A `docker-compose.yml` file configured for Vega AI
- Gemini API key from [Google AI Studio](https://aistudio.google.com/app/apikey)

## Environment Variables and Docker Swarm

Docker Stack Deploy has a [known limitation](https://github.com/moby/moby/issues/29133) where it doesn't automatically read `.env` files like `docker-compose up` does. This means environment variables must be handled differently when deploying to a Swarm.

## Deployment Methods

### Method 1: Using docker-compose config (Recommended)

This method processes your `.env` file and creates a fully resolved configuration.

1. Create your `.env` file:

```bash
GEMINI_API_KEY=your-gemini-api-key
ADMIN_USERNAME=your-admin-username  # optional, defaults to admin
ADMIN_PASSWORD=your-secure-password  # optional, defaults to VegaAdmin
TOKEN_SECRET=your-secret-token  # optional
```

2. Deploy using docker-compose config:

```bash
docker-compose config | docker stack deploy -c - vega-stack
```

This command:

- Reads your `.env` file
- Substitutes all variables in your `docker-compose.yml`
- Pipes the processed configuration to `docker stack deploy`

### Method 2: Inline Environment Variables

For simpler deployments, you can specify environment variables directly in your `docker-compose.yml`:

```yaml
services:
  vega-ai:
    image: ghcr.io/benidevo/vega-ai:latest
    environment:
      - GEMINI_API_KEY=your-gemini-api-key
      - ADMIN_USERNAME=your-admin-username  # optional, defaults to admin
      - ADMIN_PASSWORD=your-secure-password  # optional, defaults to VegaAdmin
      - TOKEN_SECRET=your-secret-token  # optional
    ports:
      - "8765:8765"
    volumes:
      - vega-data:/app/data
    deploy:
      replicas: 1
      restart_policy:
        condition: on-failure

volumes:
  vega-data:
```

Deploy with:

```bash
docker stack deploy -c docker-compose.yml vega-stack
```

## Troubleshooting

### Environment Variables Not Loading

If your environment variables aren't being recognized:

1. Verify your `.env` file exists and is readable
2. Check for syntax errors in your `.env` file
3. Use `docker service logs vega-stack_vega-ai` to check for errors
4. Ensure variables are properly exported in shell scripts

### Default Credentials Warning

If you see warnings about default credentials:

- The container is using default values because environment variables weren't passed correctly
- Use Method 1 or 2 above to ensure variables are loaded properly
- After deployment, change credentials via the web interface at Settings → Account

## Managing Your Deployment

```bash
# View stack services
docker stack services vega-stack

# View service logs
docker service logs vega-stack_vega-ai

# Scale the service
docker service scale vega-stack_vega-ai=3

# Update the service
docker service update --image ghcr.io/benidevo/vega-ai:latest vega-stack_vega-ai

# Remove the stack
docker stack rm vega-stack
```
