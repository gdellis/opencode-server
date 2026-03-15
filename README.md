# OpenCode Remote Server

A secure authentication wrapper for OpenCode's web interface with project management capabilities.

## Features

- **Custom Login UI** - Clean, modern authentication page
- **Session Management** - Secure HTTP-only cookies with configurable expiry
- **Project Management** - Create and switch between projects
- **Rate Limiting** - Protection against brute-force attacks
- **Docker Ready** - Production-ready containerized deployment

## Quick Start

### 1. Generate Password Hash

```bash
# Install htpasswd (apache2-utils)
# On Ubuntu/Debian:
sudo apt install apache2-utils

# Generate bcrypt hash
htpasswd -nBC 10 "" | tr -d ':\n' | sed 's/$2y/$2a/'
# Enter your password when prompted
# Copy the output (starts with $2a$10$...)
```

### 2. Generate Session Secret

```bash
openssl rand -hex 32
```

### 3. Configure Environment

```bash
cp .env.example .env
# Edit .env and set:
# - AUTH_PASSWORD_HASH (from step 1)
# - SESSION_SECRET (from step 2)
# - Your API keys (ANTHROPIC_API_KEY, etc.)
```

### 4. Build and Run

```bash
docker compose up -d --build
```

Access at http://localhost:4096

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
| `OPENCODE_PATH` | No | `opencode` | Path to opencode binary |
| `OPENCODE_ARGS` | No | - | Additional opencode arguments |

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
- [ ] Keep Docker image updated
- [ ] Review logs regularly for suspicious activity
- [ ] Use secrets management for API keys

## Development

### Prerequisites

- Go 1.24+
- OpenCode installed

### Run Locally

```bash
go mod tidy
go run ./cmd/server
```

### Build Binary

```bash
CGO_ENABLED=0 go build -ldflags="-s -w" -o auth-proxy ./cmd/server
```

## License

MIT