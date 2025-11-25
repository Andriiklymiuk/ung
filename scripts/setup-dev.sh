#!/bin/bash

# UNG Development Environment Setup Script
# This script helps you set up all .env files for local development

set -e

echo "🐾 UNG Development Setup"
echo "========================"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to generate random secret
generate_secret() {
    openssl rand -base64 64 | tr -d '\n'
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the UNG repository root"
    exit 1
fi

echo -e "${BLUE}📝 Setting up environment files...${NC}"
echo ""

# ==========================================
# API .env setup
# ==========================================

if [ -f "api/.env" ]; then
    echo -e "${YELLOW}⚠️  api/.env already exists. Skipping...${NC}"
else
    echo -e "${GREEN}Creating api/.env...${NC}"

    # Generate JWT secret
    JWT_SECRET=$(generate_secret)

    cat > api/.env << EOF
# API Configuration
PORT=8080
ENV=development

# Database
API_DATABASE_PATH=$HOME/.ung/api.db
USER_DATA_DIR=$HOME/.ung/users

# Security
JWT_SECRET=$JWT_SECRET

# CORS
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080

# Email/SMTP Configuration (Optional - configure when needed)
# SMTP_HOST=smtp.gmail.com
# SMTP_PORT=587
# SMTP_USERNAME=your-email@gmail.com
# SMTP_PASSWORD=your-app-password
# SMTP_FROM_EMAIL=noreply@ung.app
# SMTP_FROM_NAME=UNG Billing
# SMTP_USE_TLS=true

# Scheduler Configuration (Optional)
ENABLE_SCHEDULER=false
SCHEDULER_INVOICE_REMINDERS=false
SCHEDULER_OVERDUE_NOTIFICATIONS=false
SCHEDULER_CONTRACT_REMINDERS=false
SCHEDULER_WEEKLY_SUMMARY=false
SCHEDULER_MONTHLY_REPORTS=false
EOF

    echo -e "${GREEN}✅ Created api/.env${NC}"
    echo -e "   JWT_SECRET: ${YELLOW}[Auto-generated]${NC}"
fi

echo ""

# ==========================================
# Telegram .env setup
# ==========================================

if [ -f "telegram/.env" ]; then
    echo -e "${YELLOW}⚠️  telegram/.env already exists. Skipping...${NC}"
else
    echo -e "${GREEN}Creating telegram/.env...${NC}"

    # Read JWT secret from API .env if it exists
    if [ -f "api/.env" ]; then
        JWT_SECRET=$(grep "^JWT_SECRET=" api/.env | cut -d '=' -f2)
    else
        JWT_SECRET=$(generate_secret)
    fi

    cat > telegram/.env << EOF
# Telegram Bot Configuration
# Get your token from https://t.me/botfather
TELEGRAM_BOT_TOKEN=YOUR_BOT_TOKEN_HERE

# UNG API Configuration
UNG_API_URL=http://localhost:8080
WEB_APP_URL=https://ung.app

# JWT Secret (must match API)
JWT_SECRET=$JWT_SECRET

# Debug mode
DEBUG=false
EOF

    echo -e "${GREEN}✅ Created telegram/.env${NC}"
    echo -e "   ${YELLOW}⚠️  Remember to update TELEGRAM_BOT_TOKEN${NC}"
fi

echo ""
echo "================================"
echo -e "${GREEN}✅ Setup complete!${NC}"
echo "================================"
echo ""
echo -e "${BLUE}📋 Next steps:${NC}"
echo ""
echo "1. ${YELLOW}Create Telegram Bot:${NC}"
echo "   • Open Telegram and search for @BotFather"
echo "   • Send: /newbot"
echo "   • Follow the prompts"
echo "   • Copy the bot token"
echo "   • Update telegram/.env with your token"
echo ""
echo "2. ${YELLOW}Start the API server:${NC}"
echo "   cd api && go run ./cmd/server"
echo ""
echo "3. ${YELLOW}Start the Telegram bot:${NC}"
echo "   cd telegram && go run ./cmd/bot"
echo ""
echo "4. ${YELLOW}Test the CLI:${NC}"
echo "   go run . create"
echo ""
echo -e "${BLUE}📚 Full documentation:${NC} CONFIGURATION_GUIDE.md"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"
