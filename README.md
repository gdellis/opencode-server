# OpenCode Remote Server

A secure authentication wrapper for OpenCode's web interface with project management capabilities.

**Extends the official `ghcr.io/anomalyco/opencode` container** - no separate opencode installation needed.

## Features

- **Custom Login UI** - Clean, modern authentication page
- **Session Management** - Secure HTTP-only cookies with configurable expiry
- **Project Management** - Create and switch between projects
- **Rate Limiting** - Protection against brute-force attacks
- **Single Container** - Extends official opencode image (~70MB total)
- **Easy Setup Script** - One-command configuration

## Quick Start

### Option 1: Automated Setup (Recommended)

```bash
# Run the setup script
make setup
# or: ./setup.sh

# Start the server
make up
# or: docker compose up -d

# Access at http://localhost:4096
```

The setup script will:
1. Prompt for a password (min 8 characters)
2. Generate bcrypt hash with correct format
3. Generate session secret
4. Ask for session expiry duration (default: 24h)
5. Ask for OpenCode model (default: anthropic/claude-sonnet-4-5)
6. Optionally configure API keys
7. Create `.env` file

### Option 2: Manual Setup

#### 1. Install htpasswd

```bash
# Ubuntu/Debian
sudo apt install apache2-utils

# macOS
brew install httpd

# Alpine
apk add apache2-utils
```

#### 2. Generate Password Hash

```bash
# Generate bcrypt hash (outputs $2y$ format)
htpasswd -nBC 10 "" | tr -d ':\n'

# Then convert to $2a$ format and escape for Docker Compose:
# Input:  $2y$10$hash...
# Output: $$2a$$10$$hash...
# Replace all $ with $$
```

#### 3. Generate Session Secret

```bash
openssl rand -hex 32
```

#### 4. Create .env File

```bash
cp .env.example .env
# Edit .env with:
# - AUTH_PASSWORD_HASH (from step 2, with $$ escaping)
# - SESSION_SECRET (from step 3)
# - API keys (optional)
```

#### 5. Start Server

```bash
docker compose up -d --build
```

Access at http://localhost:4096

## Available Commands

```bash
make setup    # Generate .env file with password (first-time setup)
make up       # Start containers
make down     # Stop containers
make logs     # View container logs (follow mode)
make clean    # Remove containers, volumes, and .env
make build    # Build Docker image
make restart  # Rebuild and restart containers
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AUTH_PASSWORD_HASH` | Yes | - | bcrypt hash of login password |
| `SESSION_SECRET` | Yes | - | Random secret for session cookies |
| `SESSION_EXPIRY` | No | `24h` | Session duration |
| `LISTEN_ADDR` | No | `:4096` | Server listen address |
| `INTERNAL_ADDR` | No | `127.0.0.1:4097` | Internal opencode address |
| `PROJECTS_DIR` | No | `/var/opencode/projects` | Projects storage directory |
| `OPENCODE_ARGS` | No | - | Additional opencode arguments (e.g., `--model anthropic/claude-sonnet-4-5`) |

### OpenCode Configuration

Place your `opencode.json` in `./config/` directory. It will be mounted read-only.

Example `opencode.json`:
```json
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "provider": {
    "anthropic": {
      "options": {
        "apiKey": "{env:ANTHROPIC_API_KEY}"
      }
    }
  }
}
```

## API Endpoints

| Endpoint | Method | Description | Auth |
|----------|--------|-------------|------|
| `/login` | GET | Login page | No |
| `/login` | POST | Authenticate | No |
| `/logout` | GET | End session | No |
| `/api/health` | GET | Health check | No |
| `/api/projects` | GET | List projects | Yes |
| `/api/projects` | POST | Create project | Yes |
| `/api/projects/switch` | POST | Switch project | Yes |
| `/*` | ALL | Proxied to OpenCode | Yes |

### Create Project

```bash
curl -X POST http://localhost:4096/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name": "my-project"}'
```

### Switch Project

```bash
curl -X POST http://localhost:4096/api/projects/switch \
  -H "Content-Type: application/json" \
  -d '{"name": "my-project"}'
```

## Architecture

```
┌────────────────────────────────────────┐
│         Docker Container (~70MB)       │
│                                        │
│  ┌──────────────────────────────────┐  │
│  │      Auth Proxy (Go) :4096       │  │
│  │      - Login page                │  │
│  │      - Session management        │  │
│  │      - Project API               │  │
│  └─────────┬────────────────────────┘  │
│            │ reverse proxy              │
│            ▼                            │
│  ┌──────────────────────────────────┐  │
│  │   opencode serve :4097           │  │
│  │   (from base image)              │  │
│  └──────────────────────────────────┘  │
│                                        │
│  /var/opencode/projects/               │
└────────────────────────────────────────┘
```

## HTTPS/SSL

For production, run behind a reverse proxy like Caddy or Nginx with Let's Encrypt:

### Caddy Example

```caddyfile
opencode.yourdomain.com {
    tls your-email@example.com
    reverse_proxy localhost:4096
}
```

### Nginx Example

```nginx
server {
    listen 443 ssl http2;
    server_name opencode.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/opencode.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/opencode.yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:4096;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## Security Checklist

- [ ] Use strong password (20+ characters)
- [ ] Generate strong SESSION_SECRET (32+ random bytes)
- [ ] Set appropriate SESSION_EXPIRY (e.g., 8h for workday)
- [ ] Run behind HTTPS reverse proxy
- [ ] Keep Docker image updated (`docker compose pull && docker compose up -d --build`)
- [ ] Review logs regularly for suspicious activity
- [ ] Use secrets management for API keys

## Development

### Prerequisites

- Go 1.24+
- Docker (for testing with opencode)

### Run Locally

```bash
go mod tidy
# Requires opencode binary installed
go run ./cmd/server
```

### Build Binary

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-proxy ./cmd/server
```

### Build Docker Image

```bash
docker build -t opencode-server .
```

## Updating OpenCode

The image uses `ghcr.io/anomalyco/opencode:latest` as base. To update:

```bash
docker compose pull opencode-server
docker compose up -d --build
```

Or pin to a specific version in the Dockerfile:

```dockerfile
FROM ghcr.io/anomalyco/opencode:0.0.0-beta-202603152037
```

## Troubleshooting

### Port Already in Use

If port 4096 is in use (e.g., by a running OpenCode instance), modify `docker-compose.yml`:

```yaml
ports:
  - "4098:4096"  # Use different external port
```

### Password Not Working

Ensure the bcrypt hash is correctly formatted:
- Must use `$2a$` prefix (not `$2b$` or `$2y$`)
- Must escape `$` as `$$` in `.env` file for Docker Compose

The setup script handles this automatically. For manual setup:

```bash
# Generate correct format
echo "your-password" | htpasswd -niBC 10 "" | tr -d ':\n' | sed 's/\$2y\$/\$2a\$/; s/\$/$$/g'
```

### Container Not Starting

Check logs:
```bash
docker compose logs
```

Common issues:
- Missing `AUTH_PASSWORD_HASH` or `SESSION_SECRET` in `.env`
- Invalid bcrypt hash format
- Port conflicts

## License

MIT