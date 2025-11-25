# 🔧 UNG Configuration Guide

Complete guide for setting up all UNG components: CLI, API, Telegram Bot, and VSCode Extension.

---

## 📋 Table of Contents

1. [Local Development Setup](#1-local-development-setup)
2. [API Server Configuration](#2-api-server-configuration)
3. [Telegram Bot Setup](#3-telegram-bot-setup)
4. [AWS S3 Configuration (Optional)](#4-aws-s3-configuration-optional)
5. [GitHub Secrets & CI/CD](#5-github-secrets--cicd)
6. [VSCode Extension Publishing](#6-vscode-extension-publishing)
7. [Production Deployment](#7-production-deployment)

---

## 1. Local Development Setup

### CLI (UNG)

The CLI works out of the box! Just install and run:

```bash
# Build from source
go build -o ung .

# Or install via go
go install github.com/Andriiklymiuk/ung@latest
```

**Data Location:**
- macOS: `~/.ung/`
- Linux: `~/.ung/`
- Windows: `%USERPROFILE%\.ung\`

No configuration needed for CLI!

---

## 2. API Server Configuration

### Create `.env` file in `api/` directory

```bash
# Copy example
cp api/.env.example api/.env

# Edit with your values
nano api/.env
```

### Required Variables

```bash
# Server
PORT=8080
ENV=development

# Database (SQLite paths)
API_DATABASE_PATH=/home/user/.ung/api.db
USER_DATA_DIR=/home/user/.ung/users

# Security - CHANGE THIS IN PRODUCTION!
JWT_SECRET=your-super-secret-jwt-key-minimum-32-chars-recommended
```

### Optional: Email/SMTP Configuration

For automated invoice reminders and notifications:

```bash
# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-gmail-app-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME=UNG Billing
SMTP_USE_TLS=true

# Scheduler Features
ENABLE_SCHEDULER=true
SCHEDULER_INVOICE_REMINDERS=true
SCHEDULER_OVERDUE_NOTIFICATIONS=true
SCHEDULER_CONTRACT_REMINDERS=true
SCHEDULER_WEEKLY_SUMMARY=false
SCHEDULER_MONTHLY_REPORTS=false
```

**Gmail Setup:**
1. Enable 2FA on your Google Account
2. Go to: https://myaccount.google.com/apppasswords
3. Generate app password for "Mail"
4. Use that 16-character password in `SMTP_PASSWORD`

### Optional: CORS Configuration

```bash
# For web frontends (if you build one)
ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com
```

### Start API Server

```bash
cd api
go run ./cmd/server
```

Server will be available at `http://localhost:8080`

---

## 3. Telegram Bot Setup

### Step 1: Create Bot with BotFather

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot`
3. Follow prompts:
   - Choose bot name (e.g., "UNG Billing Bot")
   - Choose username (e.g., "YourUngBot")
4. Copy the **bot token** (format: `1234567890:ABCdefGHIjklMNOpqrsTUVwxyz`)

### Step 2: Configure Bot

Create `.env` file in `telegram/` directory:

```bash
# Copy example
cp telegram/.env.example telegram/.env

# Edit with your values
nano telegram/.env
```

```bash
# Bot Token from BotFather
TELEGRAM_BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# API Configuration (must match API server)
UNG_API_URL=http://localhost:8080
JWT_SECRET=your-super-secret-jwt-key-minimum-32-chars-recommended

# Optional: Web App URL (if you build web frontend)
WEB_APP_URL=https://yourdomain.com

# Debug mode
DEBUG=false
```

### Step 3: Start Bot

```bash
cd telegram
go run ./cmd/bot
```

### Step 4: Register with Bot

1. Open Telegram and find your bot
2. Send `/start`
3. Follow registration flow to create account

---

## 4. AWS S3 Configuration (Optional)

For storing encrypted tenant databases in the cloud.

### Environment Variables (add to `api/.env`)

```bash
# S3 Configuration
AWS_REGION=us-east-1
AWS_S3_BUCKET=ung-tenant-databases
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

# Optional: For S3-compatible services (MinIO, DigitalOcean Spaces, etc.)
AWS_ENDPOINT=https://nyc3.digitaloceanspaces.com
AWS_USE_PATH_STYLE=true
```

### AWS Setup (Real S3)

1. **Create IAM User:**
   - Go to AWS Console → IAM → Users → Add User
   - Name: `ung-api-s3`
   - Access: Programmatic access
   - Permissions: Attach existing policy → `AmazonS3FullAccess` (or create custom)

2. **Create S3 Bucket:**
   - Go to S3 → Create Bucket
   - Name: `ung-tenant-databases` (or your choice)
   - Region: `us-east-1` (or your choice)
   - Block all public access: ✅ ENABLED
   - Versioning: Enable (recommended)

3. **Bucket Policy (Optional - for more security):**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::YOUR_ACCOUNT_ID:user/ung-api-s3"
      },
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::ung-tenant-databases/*",
        "arn:aws:s3:::ung-tenant-databases"
      ]
    }
  ]
}
```

### DigitalOcean Spaces Alternative

```bash
AWS_REGION=nyc3
AWS_S3_BUCKET=your-space-name
AWS_ENDPOINT=https://nyc3.digitaloceanspaces.com
AWS_USE_PATH_STYLE=false
AWS_ACCESS_KEY_ID=your-spaces-key
AWS_SECRET_ACCESS_KEY=your-spaces-secret
```

---

## 5. GitHub Secrets & CI/CD

### GitHub Secrets to Configure

Go to your repo → Settings → Secrets and variables → Actions → New repository secret

#### Required Secrets

```bash
# Automatically provided by GitHub (no action needed)
GITHUB_TOKEN  # Auto-provided for releases
```

#### Optional: Homebrew Publishing

If you want to auto-publish to Homebrew tap:

```bash
# Generate GitHub Personal Access Token
# Settings → Developer settings → Personal access tokens → Tokens (classic)
# Scopes: repo (full), write:packages

GH_PAT=ghp_xxxxxxxxxxxxxxxxxxxx
```

Update `.goreleaser.yaml` brew section:

```yaml
brews:
  - name: ung
    repository:
      owner: Andriiklymiuk
      name: homebrew-tools
      token: "{{ .Env.GH_PAT }}"
```

#### Optional: VSCode Extension Publishing

If you want to publish the VSCode extension to the marketplace:

```bash
# Generate Azure DevOps Personal Access Token
# Go to: https://dev.azure.com
# User Settings → Personal Access Tokens → New Token
# Scopes: Marketplace → Manage

VSCE_PAT=your-azure-devops-token
```

**Important:** `VSCE_PAT` is an **Azure DevOps token**, NOT a GitHub token!

### GitHub Actions Workflows

Current setup (`.github/workflows/release.yml`):

```yaml
# Triggers on version tags (e.g., v1.0.0)
on:
  push:
    tags:
      - 'v*'
```

**To create a release:**

```bash
# Tag your version
git tag v1.0.0

# Push tag
git push origin v1.0.0

# GitHub Actions will automatically:
# 1. Build for Linux/macOS (amd64/arm64)
# 2. Create GitHub Release
# 3. Upload binaries
# 4. Publish to Homebrew tap (if configured)
```

### Additional Workflows You Might Want

Create `.github/workflows/test.yml`:

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - name: Run tests
        run: |
          go test ./...
          cd api && go test ./...
          cd ../telegram && go test ./...
```

---

## 6. VSCode Extension Publishing

### Prerequisites

1. **Create Azure DevOps Account**
   - Go to: https://dev.azure.com
   - Sign in with Microsoft/GitHub account

2. **Create Personal Access Token (PAT)**
   - Click User Settings (top right) → Personal Access Tokens
   - New Token
   - Name: `VSCode Marketplace - UNG`
   - Organization: All accessible organizations
   - Scopes: **Marketplace** → **Manage**
   - Create and **SAVE THE TOKEN** (you can't see it again!)

3. **Create Publisher**
   - Go to: https://marketplace.visualstudio.com/manage
   - Create Publisher
   - ID: `andriiklymiuk` (or your choice - must match package.json)
   - Display Name: Your name
   - Upload icon/logo

### Update package.json

Edit `vscode-ung/package.json`:

```json
{
  "name": "ung",
  "publisher": "andriiklymiuk",  // ← Your publisher ID
  "version": "1.0.0",
  "repository": {
    "type": "git",
    "url": "https://github.com/Andriiklymiuk/ung"
  },
  "bugs": {
    "url": "https://github.com/Andriiklymiuk/ung/issues"
  },
  "homepage": "https://github.com/Andriiklymiuk/ung#readme",
  "license": "MIT"
}
```

### Install VSCE (Publishing Tool)

```bash
npm install -g @vscode/vsce
```

### Publish Extension

```bash
cd vscode-ung

# Build TypeScript
npm run compile

# Package extension (creates .vsix file)
vsce package

# Login (first time only)
vsce login andriiklymiuk
# Enter your PAT token

# Publish
vsce publish

# Or publish with version bump
vsce publish minor  # 1.0.0 → 1.1.0
vsce publish patch  # 1.0.0 → 1.0.1
vsce publish major  # 1.0.0 → 2.0.0
```

### Automate with GitHub Actions

Create `.github/workflows/publish-vscode.yml`:

```yaml
name: Publish VSCode Extension

on:
  push:
    tags:
      - 'vscode-v*'  # e.g., vscode-v1.0.0

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '18'

      - name: Install dependencies
        working-directory: vscode-ung
        run: npm ci

      - name: Build
        working-directory: vscode-ung
        run: npm run compile

      - name: Publish to VSCode Marketplace
        working-directory: vscode-ung
        run: |
          npm install -g @vscode/vsce
          vsce publish -p ${{ secrets.VSCE_PAT }}
        env:
          VSCE_PAT: ${{ secrets.VSCE_PAT }}
```

**Add GitHub Secret:**
- Name: `VSCE_PAT`
- Value: Your Azure DevOps PAT token

**To publish:**
```bash
git tag vscode-v1.0.0
git push origin vscode-v1.0.0
```

---

## 7. Production Deployment

### Option 1: VPS (DigitalOcean, Linode, AWS EC2)

#### API Server

```bash
# Install Go
wget https://go.dev/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone https://github.com/Andriiklymiuk/ung.git
cd ung/api
go build -o ung-api ./cmd/server

# Create .env file with production values
nano .env

# Run with systemd
sudo nano /etc/systemd/system/ung-api.service
```

**systemd service file:**
```ini
[Unit]
Description=UNG API Server
After=network.target

[Service]
Type=simple
User=ung
WorkingDirectory=/home/ung/api
ExecStart=/home/ung/api/ung-api
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable ung-api
sudo systemctl start ung-api
```

#### Telegram Bot

```bash
cd /home/ung/telegram
go build -o ung-telegram ./cmd/bot

# Create systemd service
sudo nano /etc/systemd/system/ung-telegram.service
```

```ini
[Unit]
Description=UNG Telegram Bot
After=network.target

[Service]
Type=simple
User=ung
WorkingDirectory=/home/ung/telegram
ExecStart=/home/ung/telegram/ung-telegram
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable ung-telegram
sudo systemctl start ung-telegram
```

### Option 2: Docker

Create `docker-compose.yml` in root:

```yaml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: api/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - ENV=production
      - JWT_SECRET=${JWT_SECRET}
      - API_DATABASE_PATH=/data/api.db
      - USER_DATA_DIR=/data/users
    volumes:
      - ung-data:/data
    restart: unless-stopped

  telegram:
    build:
      context: .
      dockerfile: telegram/Dockerfile
    environment:
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - UNG_API_URL=http://api:8080
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - api
    restart: unless-stopped

volumes:
  ung-data:
```

**.env for Docker:**
```bash
JWT_SECRET=your-production-secret-min-32-chars
TELEGRAM_BOT_TOKEN=your-bot-token
```

**Run:**
```bash
docker-compose up -d
```

### Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name api.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Enable HTTPS with Let's Encrypt:**
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d api.yourdomain.com
```

---

## 🔐 Security Checklist

### Before Going to Production:

- [ ] Change `JWT_SECRET` to random 64+ character string
- [ ] Use environment variables, never commit secrets
- [ ] Enable HTTPS/TLS for API
- [ ] Set strong database encryption key
- [ ] Configure firewall (UFW/iptables)
- [ ] Regular database backups to S3
- [ ] Monitor logs for suspicious activity
- [ ] Rate limiting on API endpoints
- [ ] Keep dependencies updated

### Generate Strong JWT Secret:

```bash
# Linux/macOS
openssl rand -base64 64

# Or
head -c 64 /dev/urandom | base64
```

---

## 📊 Monitoring & Logging

### Application Logs

```bash
# API logs
sudo journalctl -u ung-api -f

# Telegram bot logs
sudo journalctl -u ung-telegram -f
```

### Optional: Advanced Monitoring

- **Prometheus** + **Grafana** for metrics
- **Loki** for log aggregation
- **Sentry** for error tracking
- **Uptime Kuma** for uptime monitoring

---

## 🆘 Troubleshooting

### API won't start
```bash
# Check port availability
sudo lsof -i :8080

# Check environment variables
env | grep -E "(JWT|PORT|DATABASE)"

# Check database permissions
ls -la ~/.ung/
```

### Telegram bot not responding
```bash
# Verify bot token
curl https://api.telegram.org/bot<YOUR_TOKEN>/getMe

# Check API connection
curl http://localhost:8080/health

# Check logs
journalctl -u ung-telegram --since "10 minutes ago"
```

### VSCode extension not working
```bash
# Rebuild extension
cd vscode-ung
npm run compile

# Check UNG CLI is in PATH
which ung
ung --version
```

---

## 📝 Quick Reference

### All Environment Variables

| Variable | Location | Required | Default | Description |
|----------|----------|----------|---------|-------------|
| `PORT` | api/.env | No | 8080 | API server port |
| `JWT_SECRET` | api/.env, telegram/.env | **YES** | - | JWT signing key (32+ chars) |
| `API_DATABASE_PATH` | api/.env | No | ~/.ung/api.db | Main API database |
| `USER_DATA_DIR` | api/.env | No | ~/.ung/users | Tenant databases directory |
| `TELEGRAM_BOT_TOKEN` | telegram/.env | **YES** | - | From @BotFather |
| `UNG_API_URL` | telegram/.env | No | http://localhost:8080 | API server URL |
| `SMTP_HOST` | api/.env | No | - | SMTP server |
| `SMTP_USERNAME` | api/.env | No | - | Email username |
| `SMTP_PASSWORD` | api/.env | No | - | Email password |
| `AWS_S3_BUCKET` | api/.env | No | - | S3 bucket name |
| `AWS_ACCESS_KEY_ID` | api/.env | No | - | AWS credentials |
| `AWS_SECRET_ACCESS_KEY` | api/.env | No | - | AWS credentials |

### GitHub Secrets (for CI/CD)

| Secret | Purpose | Where to Get | Required For |
|--------|---------|--------------|--------------|
| `GITHUB_TOKEN` | GitHub releases | Auto-provided by GitHub Actions | CLI releases |
| `GH_PAT` | Homebrew tap publishing | GitHub → Settings → Tokens (classic)<br/>Scopes: `repo` | Homebrew formula |
| `VSCE_PAT` | VSCode Marketplace publishing | Azure DevOps → PAT<br/>Scopes: `Marketplace: Manage` | VSCode extension |

**Important:** `GH_PAT` and `VSCE_PAT` are **different tokens** from **different services**!

---

## 🎯 Next Steps

1. **Local Development**: Start with API and Telegram bot locally
2. **Test Features**: Create invoices, track time, test workflows
3. **Prepare Production**: Set up VPS/cloud hosting
4. **Deploy**: Use Docker or systemd services
5. **Publish Extension**: Release to VSCode Marketplace
6. **Monitor**: Set up logging and monitoring
7. **Scale**: Add S3 storage and backups as needed

---

## 🤝 Support

- **Issues**: https://github.com/Andriiklymiuk/ung/issues
- **Discussions**: https://github.com/Andriiklymiuk/ung/discussions
- **Security**: Report security issues privately

---

**Last Updated**: 2025-01-25
**Version**: 1.0.0
