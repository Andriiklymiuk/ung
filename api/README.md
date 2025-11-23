# UNG REST API

Multi-tenant REST API for UNG (Universal Next-Gen Billing & Tracking). Built with Go, Chi router, and SQLite.

## Features

- 🔐 JWT-based authentication
- 👥 Multi-tenant architecture (each user has isolated database)
- 🚀 Fast and lightweight (Chi router)
- 💾 SQLite for data persistence
- 🐳 Docker support
- 📝 Standard REST API design

## Quick Start

### Local Development

```bash
# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

The API will be available at `http://localhost:8080`

### Using Docker

```bash
# Build and run
docker-compose up --build

# Or build manually
docker build -t ung-api .
docker run -p 8080:8080 ung-api
```

## API Endpoints

### Health Check

```bash
GET /health
```

### Authentication

```bash
# Register
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "John Doe"
}

# Login
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}

# Refresh Token
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}

# Get Profile
GET /api/v1/auth/me
Authorization: Bearer {access_token}
```

### Invoices (Protected)

```bash
# List invoices
GET /api/v1/invoices
Authorization: Bearer {access_token}

# Get specific invoice
GET /api/v1/invoices/{id}
Authorization: Bearer {access_token}
```

## Architecture

### Multi-Tenant Design

Each user has their own isolated SQLite database:

```
~/.ung/
├── api.db              # API metadata (users, auth)
└── users/
    ├── user_1/
    │   └── ung.db      # User 1's data
    ├── user_2/
    │   └── ung.db      # User 2's data
    └── ...
```

### Request Flow

1. Client sends request with JWT token
2. `AuthMiddleware` validates token and loads user
3. `TenantMiddleware` opens user's database
4. Controller processes request with user's data
5. Response sent back to client

### Project Structure

```
api/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration
│   ├── middleware/
│   │   ├── auth.go           # JWT authentication
│   │   └── tenant.go         # Multi-tenant DB switching
│   ├── controllers/
│   │   ├── auth_controller.go
│   │   └── invoice_controller.go
│   ├── services/
│   │   └── auth_service.go   # Business logic
│   ├── repository/
│   │   └── user_repository.go
│   ├── models/
│   │   └── models.go         # Data models
│   ├── database/
│   │   └── database.go       # DB initialization
│   └── router/
│       └── router.go         # Route definitions
├── pkg/
│   └── utils/
│       ├── jwt.go            # JWT utilities
│       └── password.go       # Password hashing
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

## Configuration

Environment variables:

- `PORT` - Server port (default: 8080)
- `ENV` - Environment (development/production)
- `API_DATABASE_PATH` - Path to API database
- `USER_DATA_DIR` - Directory for user databases
- `JWT_SECRET` - Secret key for JWT signing (⚠️ change in production!)

## Security

- Passwords hashed with bcrypt (cost factor 12)
- JWT tokens with 15-minute expiry (access) and 7-day expiry (refresh)
- CORS configured for specific origins
- Input validation on all endpoints
- SQL injection prevented by GORM

## Development

### Adding New Endpoints

1. Create controller method in `internal/controllers/`
2. Add route in `internal/router/router.go`
3. Add business logic in `internal/services/` if needed

Example:

```go
// Controller
func (c *InvoiceController) Create(w http.ResponseWriter, r *http.Request) {
    db := middleware.GetTenantDB(r)
    // ... implementation
}

// Router
r.Post("/invoices", invoiceController.Create)
```

### Testing

```bash
# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Use the access_token from login response
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer {access_token}"
```

## Deployment

### Using Docker

1. Build image:
   ```bash
   docker build -t ung-api:latest .
   ```

2. Run with environment variables:
   ```bash
   docker run -d \
     -p 8080:8080 \
     -v /path/to/data:/root/.ung \
     -e JWT_SECRET=your-secret-key \
     ung-api:latest
   ```

### Using systemd

```ini
[Unit]
Description=UNG API Service
After=network.target

[Service]
Type=simple
User=ung
WorkingDirectory=/opt/ung-api
ExecStart=/opt/ung-api/ung-api
Restart=on-failure
EnvironmentFile=/opt/ung-api/.env

[Install]
WantedBy=multi-user.target
```

## License

MIT

## Links

- CLI Tool: `../` (parent directory)
- Documentation: `../docs/GO_API.md`
