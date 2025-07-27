# Secure Configuration Guide

This guide explains how to securely configure Vega AI using the `_FILE` environment variable pattern, which is particularly useful when deploying with Docker Secrets or Kubernetes Secrets.

## Overview

Vega AI supports reading any environment variable from a file by appending `_FILE` to the variable name. This allows you to:

- Store sensitive data in secure secret management systems
- Avoid exposing secrets in environment variables
- Comply with security best practices
- Use Docker Secrets and Kubernetes Secrets

## How It Works

For any environment variable `KEY`, you can instead use `KEY_FILE` pointing to a file containing the value:

```bash
# Traditional method (less secure)
GEMINI_API_KEY=your-api-key

# File-based method (more secure)
GEMINI_API_KEY_FILE=/run/secrets/gemini_api_key
```

The application will:
1. First check if `KEY_FILE` exists
2. If yes, read the value from that file
3. If no, fall back to checking `KEY`
4. If neither exists, use the default value

## Docker Secrets Example

### 1. Create Secrets

Using our helper script:
```bash
./scripts/create-docker-secrets.sh
```

Or manually:
```bash
# Create individual secrets
echo -n "your-gemini-api-key" | docker secret create gemini_api_key -
echo -n "your-secure-token" | docker secret create token_secret -
echo -n "your-admin-password" | docker secret create admin_password -
```

### 2. Deploy with Secrets

```yaml
version: '3.8'

services:
  vega-ai:
    image: ghcr.io/benidevo/vega-ai:latest
    environment:
      # All of these read from files
      - GEMINI_API_KEY_FILE=/run/secrets/gemini_api_key
      - TOKEN_SECRET_FILE=/run/secrets/token_secret
      - ADMIN_PASSWORD_FILE=/run/secrets/admin_password
    secrets:
      - gemini_api_key
      - token_secret 
      - admin_password

secrets:
  gemini_api_key:
    external: true
  token_secret:
    external: true
  admin_password:
    external: true
```

## Kubernetes Secrets Example

### 1. Create Secrets

```bash
# Create a Kubernetes secret
kubectl create secret generic vega-secrets \
  --from-literal=gemini-api-key=your-api-key \
  --from-literal=token-secret=your-token \
  --from-literal=admin-password=your-password
```

### 2. Deploy with Secrets

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vega-ai
spec:
  template:
    spec:
      containers:
      - name: vega-ai
        image: ghcr.io/benidevo/vega-ai:latest
        env:
        - name: GEMINI_API_KEY_FILE
          value: /secrets/gemini-api-key
        - name: TOKEN_SECRET_FILE
          value: /secrets/token-secret
        - name: ADMIN_PASSWORD_FILE
          value: /secrets/admin-password
        volumeMounts:
        - name: secrets
          mountPath: /secrets
          readOnly: true
      volumes:
      - name: secrets
        secret:
          secretName: vega-secrets
          items:
          - key: gemini-api-key
            path: gemini-api-key
          - key: token-secret
            path: token-secret
          - key: admin-password
            path: admin-password
```

## Local Development with Files

For local development, you can also use files:

```bash
# Create a secrets directory
mkdir -p .secrets

# Create secret files
echo -n "your-api-key" > .secrets/gemini_api_key
echo -n "your-token" > .secrets/token_secret
echo -n "your-password" > .secrets/admin_password

# Set file permissions
chmod 600 .secrets/*

# Run with file-based secrets
export GEMINI_API_KEY_FILE=.secrets/gemini_api_key
export TOKEN_SECRET_FILE=.secrets/token_secret
export ADMIN_PASSWORD_FILE=.secrets/admin_password

# Don't forget to add .secrets to .gitignore!
echo ".secrets/" >> .gitignore
```

## Available Variables with _FILE Support

ALL environment variables support the `_FILE` pattern. Here are some commonly secured ones:

| Variable | File Variable | Description |
|----------|--------------|-------------|
| `GEMINI_API_KEY` | `GEMINI_API_KEY_FILE` | Google Gemini API key |
| `TOKEN_SECRET` | `TOKEN_SECRET_FILE` | JWT signing secret |
| `ADMIN_PASSWORD` | `ADMIN_PASSWORD_FILE` | Initial admin password |
| `DB_CONNECTION_STRING` | `DB_CONNECTION_STRING_FILE` | Database connection string |
| `GOOGLE_CLIENT_SECRET` | `GOOGLE_CLIENT_SECRET_FILE` | OAuth client secret |

## Best Practices

1. **Use _FILE for all sensitive data**: API keys, passwords, tokens, connection strings
2. **Set appropriate file permissions**: Files should be readable only by the application user
3. **Rotate secrets regularly**: Update secret files without changing application configuration
4. **Never commit secrets**: Add secret files to `.gitignore`
5. **Use secret management tools**: Docker Secrets, Kubernetes Secrets, HashiCorp Vault, etc.

## Troubleshooting

### Warning Messages

If you see warnings like:
```
Warning: Failed to read KEY_FILE from /path/to/file: no such file or directory
```

This means:
- The file path specified in `KEY_FILE` doesn't exist
- The application will fall back to using `KEY` or the default value
- Check your file paths and secret mounts

### File Format

- Files should contain only the secret value
- Trailing newlines are automatically trimmed
- The entire file content is read as the value
- Empty files result in empty string values

### Migration from Environment Variables

To migrate existing deployments:

1. Create secrets in your platform (Docker/Kubernetes)
2. Update environment variables to use `_FILE` suffix
3. Deploy the updated configuration
4. Remove old environment variables

## Security Benefits

Using file-based secrets provides:

- **No exposure in process listings**: `ps aux` won't show secrets
- **No exposure in logs**: Environment dumps won't include secrets  
- **Encrypted at rest**: When using Docker/Kubernetes secrets
- **Access control**: Only authorized services can read secrets
- **Audit trail**: Secret access can be logged and monitored
- **Easy rotation**: Update secrets without redeploying