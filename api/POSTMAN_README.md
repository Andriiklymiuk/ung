# UNG API - Postman Collection

Complete Postman collection for testing and exploring the UNG API.

## Files

- **UNG_API.postman_collection.json** - Main API collection with all endpoints
- **UNG_API.postman_environment.json** - Environment variables for local development

## Quick Start

### 1. Import Collection

1. Open Postman
2. Click **Import** button
3. Drag and drop `UNG_API.postman_collection.json`
4. Import `UNG_API.postman_environment.json` as well

### 2. Select Environment

1. Click the environment dropdown (top right)
2. Select **UNG API - Local Development**

### 3. Start API Server

```bash
cd api
cp .env.example .env
# Edit .env with your configuration
go run cmd/server/main.go
```

Server should start on `http://localhost:3000`

### 4. Test Health Check

Run the **Health Check** request to verify the server is running.

## Authentication Flow

### Register New User

1. Run **Authentication → Register User**
2. Tokens are automatically saved to environment variables
3. You're ready to make authenticated requests!

### Login Existing User

1. Run **Authentication → Login**
2. Tokens are automatically saved to environment variables

### Token Refresh

Access tokens expire after 15 minutes. To refresh:

1. Run **Authentication → Refresh Token**
2. New access token is automatically saved

## Environment Variables

The collection uses the following environment variables:

| Variable | Description | Auto-Set |
|----------|-------------|----------|
| `base_url` | API base URL | Manual |
| `api_version` | API version (v1) | Manual |
| `access_token` | JWT access token | Auto |
| `refresh_token` | JWT refresh token | Auto |
| `tenant_id` | Current tenant ID | Auto |

## Endpoints

### Authentication
- **POST** `/auth/register` - Register new user
- **POST** `/auth/login` - Login user
- **POST** `/auth/refresh` - Refresh access token
- **GET** `/auth/me` - Get user profile

### Invoices
- **GET** `/invoices` - List invoices (with filters)
- **GET** `/invoices/:id` - Get invoice by ID
- **POST** `/invoices` - Create invoice
- **PUT** `/invoices/:id` - Update invoice
- **DELETE** `/invoices/:id` - Delete invoice
- **POST** `/invoices/:id/email` - Email invoice

### Health
- **GET** `/health` - Health check (no auth required)

## Features

### Automatic Token Management

The collection includes scripts that automatically:
- Save access and refresh tokens after registration/login
- Include Bearer token in authenticated requests
- Validate response times and content types

### Query Parameters

Many endpoints support filtering and pagination:

```
GET /invoices?status=pending&limit=10&offset=0
GET /invoices?client_id=5&from_date=2024-01-01
```

### Example Requests

Each endpoint includes:
- Detailed descriptions
- Request body examples
- Response examples
- Parameter documentation

## Testing

The collection includes global test scripts that run for every request:

```javascript
// Response time check
pm.test('Response time is acceptable', function () {
    pm.expect(pm.response.responseTime).to.be.below(2000);
});

// Content-Type validation
pm.test('Response has correct content type', function () {
    pm.expect(pm.response.headers.get('Content-Type')).to.include('application/json');
});
```

## Tips

### Multi-Environment Setup

Create multiple environments for different setups:

1. **Local Development** - `http://localhost:3000`
2. **Staging** - `https://staging-api.ung.app`
3. **Production** - `https://api.ung.app`

### Collection Variables

The collection includes these default variables:
- `base_url`: http://localhost:3000
- `api_version`: v1

You can override these in your environment.

### SMTP Configuration

To test email endpoints, configure SMTP in your `.env` file:

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM_EMAIL=noreply@ung.app
SMTP_USE_TLS=true
```

For Gmail, you'll need to generate an [App Password](https://support.google.com/accounts/answer/185833).

## Troubleshooting

### 401 Unauthorized

- Check if access token is set: `{{access_token}}`
- Try refreshing token with **Refresh Token** endpoint
- Re-login if refresh token expired

### 500 Internal Server Error

- Check API server logs
- Verify database path in `.env`
- Ensure all required environment variables are set

### Connection Refused

- Verify API server is running
- Check `base_url` matches server port
- Confirm firewall settings

## Development

### Adding New Endpoints

1. Create new request in appropriate folder
2. Add authentication if needed (Bearer token)
3. Include example request body
4. Document parameters in description
5. Add test scripts if needed

### Testing Automated Workflows

Use **Collection Runner** to test complete workflows:

1. Click **Run Collection**
2. Select requests to run
3. Configure iterations and delays
4. View test results

## Support

For issues or questions:
- API Documentation: `/docs` endpoint (coming soon)
- GitHub Issues: [ung repository](https://github.com/Andriiklymiuk/ung/issues)
- Email: support@ung.app

## License

MIT License - See LICENSE file for details
