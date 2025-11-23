# UNG Telegram Bot

Conversational billing assistant for Telegram. Create invoices, track time, and manage clients directly from Telegram.

## Features

- 🤖 **Conversational Interface** - Natural chat-based interactions
- 📄 **Invoice Management** - Create and view invoices with guided flows
- 👥 **Client Management** - Add and manage clients
- ⏱️ **Time Tracking** - Log hours with simple commands
- 📊 **Reports** - View revenue, overdue, and client reports
- 🔐 **Secure** - JWT authentication with UNG API
- 💬 **Interactive** - Inline keyboards for quick actions

## Quick Start

### Prerequisites

1. Create a Telegram bot with [@BotFather](https://t.me/botfather)
2. Get your bot token
3. Have UNG API running (see `../api/`)

### Local Development

```bash
# Create .env file
cp .env.example .env

# Edit .env and add your bot token
nano .env

# Install dependencies
go mod download

# Run the bot
go run cmd/bot/main.go
```

### Using Docker

```bash
# Build and run
docker-compose up --build

# Or standalone
docker build -t ung-telegram .
docker run -e TELEGRAM_BOT_TOKEN=your-token ung-telegram
```

## Configuration

Environment variables (see `.env.example`):

- `TELEGRAM_BOT_TOKEN` - Your bot token from BotFather (required)
- `UNG_API_URL` - URL of UNG API (default: http://localhost:8080)
- `WEB_APP_URL` - URL of web app for registration (default: https://ung.app)
- `DEBUG` - Enable debug logging (default: false)

## Bot Commands

### Getting Started

```
/start - Start the bot and see main menu
/help - Show available commands
```

### Invoices

```
/invoice - Create new invoice (interactive)
/invoices - List all invoices
/unpaid - View unpaid invoices
```

### Clients

```
/client - Manage clients
/clients - List all clients
```

### Time Tracking

```
/track - Log time
/today - Today's tracked time
/week - This week's summary
```

### Reports

```
/report - View reports
/revenue - Revenue summary
/overdue - Overdue invoices
```

### Other

```
/status - Account status
/settings - Bot settings
```

## Usage Examples

### Creating an Invoice

1. Send `/invoice`
2. Select a client from the list
3. Enter the amount (e.g., "1500")
4. Add a description
5. Choose due date (7, 14, or 30 days)
6. Done! 🎉

### Quick Time Logging

Just send a message like:
- "3h website design"
- "2.5 hours client meeting"
- "30m email"

## Architecture

### Project Structure

```
telegram/
├── cmd/
│   └── bot/
│       └── main.go           # Bot entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration
│   ├── handlers/
│   │   ├── start.go          # /start command
│   │   ├── help.go           # /help command
│   │   └── invoice.go        # Invoice handlers
│   ├── services/
│   │   ├── api_client.go     # UNG API client
│   │   └── session_manager.go # Session state
│   └── models/
│       └── models.go         # Data models
├── pkg/
│   └── utils/                # Utilities
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```

### How It Works

1. **User sends command** → Bot receives update
2. **Handler processes** → Routes to appropriate handler
3. **API call** → Communicates with UNG API
4. **State management** → Tracks conversation flow
5. **Response** → Sends message/keyboard to user

### Conversation Flow Example

```
User: /invoice
Bot: Select a client [Inline keyboard]
User: [Clicks "Acme Corp"]
Bot: Enter the amount
User: 1500
Bot: Add a description
User: Website development
Bot: When is this due? [7/14/30 days buttons]
User: [Clicks "30 days"]
Bot: ✅ Invoice created! [View PDF] [Send Email] buttons
```

## Authentication

The bot requires users to authenticate with their UNG account:

1. User sends `/start`
2. Bot shows "Create Account" or "I have an account" buttons
3. User clicks "I have an account"
4. Bot provides authentication link
5. User logs in on web app
6. Web app notifies bot via webhook
7. Bot confirms authentication
8. User can now use all features

## API Integration

The bot communicates with UNG REST API:

- **Authentication**: JWT tokens
- **Endpoints**:
  - `POST /api/v1/auth/login`
  - `GET /api/v1/invoices`
  - `POST /api/v1/invoices`
  - `GET /api/v1/clients`
  - etc.

See `internal/services/api_client.go` for implementation.

## Session Management

User conversation states are stored in memory:

```go
type Session struct {
    TelegramID int64
    State      string                      // Current conversation state
    Data       map[string]interface{}      // Session data (client_id, amount, etc.)
    UpdatedAt  time.Time
}
```

States:
- `invoice_select_client` - Waiting for client selection
- `invoice_amount` - Waiting for amount input
- `invoice_description` - Waiting for description
- `invoice_due_date` - Waiting for due date selection

## Development

### Adding New Commands

1. Create handler in `internal/handlers/`:

```go
type MyHandler struct {
    bot *tgbotapi.BotAPI
}

func (h *MyHandler) Handle(message *tgbotapi.Message) error {
    // Implementation
}
```

2. Register in `cmd/bot/main.go`:

```go
case "mycommand":
    return myHandler.Handle(message)
```

### Adding New Conversation States

1. Add state constant in `internal/models/models.go`:

```go
const (
    StateMyNewState SessionState = "my_new_state"
)
```

2. Handle state in message handler:

```go
case models.StateMyNewState:
    return handler.HandleMyState(message)
```

## Deployment

### Production Deployment

1. **Get bot token** from BotFather
2. **Set environment variables**
3. **Deploy with Docker**:

```bash
docker build -t ung-telegram:latest .
docker run -d \
  -e TELEGRAM_BOT_TOKEN=your-token \
  -e UNG_API_URL=https://api.ung.app \
  --restart unless-stopped \
  ung-telegram:latest
```

### With systemd

```ini
[Unit]
Description=UNG Telegram Bot
After=network.target

[Service]
Type=simple
User=ung
WorkingDirectory=/opt/ung-telegram
ExecStart=/opt/ung-telegram/ung-telegram
Restart=on-failure
EnvironmentFile=/opt/ung-telegram/.env

[Install]
WantedBy=multi-user.target
```

## Monitoring

Monitor bot health by checking:
- Bot is receiving updates (check logs)
- API connectivity (test endpoints)
- Memory usage (sessions are in-memory)
- Error rates in logs

## Security

- ✅ JWT authentication with API
- ✅ User-specific sessions
- ✅ Secure token storage
- ✅ HTTPS for API communication
- ✅ Environment-based configuration

## Future Enhancements

- [ ] Voice message support for time tracking
- [ ] Subscription management via RevenueCat
- [ ] Push notifications for overdue invoices
- [ ] Multi-language support
- [ ] Redis for session persistence
- [ ] Webhook mode (instead of polling)
- [ ] Admin commands
- [ ] Group chat support

## Troubleshooting

**Bot not responding:**
- Check bot token is correct
- Verify bot is running (`docker ps` or check process)
- Check logs for errors

**API errors:**
- Verify UNG_API_URL is correct
- Test API health: `curl http://localhost:8080/health`
- Check JWT token is valid

**Session issues:**
- Sessions are in-memory, restart clears all sessions
- Consider using Redis for persistence in production

## Links

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- UNG API: `../api/`
- UNG CLI: `../`

## License

MIT
