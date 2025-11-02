# Telegram Message Forwarder

A Golang service that listens to your Telegram messages and forwards messages matching regex patterns or keywords to another chat via a bot.

## What It Does

1. Connects to your Telegram account (user account)
2. Listens for incoming messages
3. Matches messages against regex patterns or keywords (managed via API)
4. Forwards matching messages using a bot to a target chat
5. Provides an HTTP API to manage forwarding rules dynamically
6. Includes a web-based admin panel for easy rule management

## Setup

### 1. Get Telegram Credentials

**User Account:**
- Go to https://my.telegram.org
- Create an app to get `app_id` and `app_hash`

**Bot Account:**
- Talk to [@BotFather](https://t.me/botfather)
- Create a bot and get the token

### 2. Configure

```bash
# Copy example env file
cp .env.example .env

# Edit with your credentials
vim .env
```

Example `.env`:
```env
TG_USER_APP_ID=12345678
TG_USER_APP_HASH=your_api_hash_here
TG_USER_PHONE=+1234567890
TG_USER_SESSION=

TG_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
TG_BOT_TARGET_CHAT_ID=987654321
TG_BOT_TARGET_USERNAME=

API_PORT=8080
API_TOKEN=your-secret-api-token-here
```

**Environment Variables:**
- `TG_USER_APP_ID`: Your Telegram app ID (from https://my.telegram.org)
- `TG_USER_APP_HASH`: Your Telegram app hash
- `TG_USER_PHONE`: Your phone number (with country code)
- `TG_USER_SESSION`: Telethon StringSession (optional, see Authentication below)
- `TG_BOT_TOKEN`: Your bot token (from @BotFather)
- `TG_BOT_TARGET_CHAT_ID`: Target chat ID to forward messages to
- `TG_BOT_TARGET_USERNAME`: Alternative to chat ID, use username (e.g., `@channel`)
- `API_PORT`: HTTP API port (default: `8080`)
- `API_TOKEN`: Secret token for API authentication
- `MONGODB_URI`: MongoDB connection string (default: `mongodb://localhost:27017`)
- `MONGODB_DATABASE`: MongoDB database name (default: `tg-forward`)

### 3. Run

**Local:**
```bash
go run cmd/tg-forward/main.go
```

**Docker:**
```bash
docker build -t tg-forward .
docker run -d -p 8080:8080 \
  --env-file .env \
  tg-forward
```

**First Run:**
On first run, you'll be prompted to enter the 2FA code sent to your Telegram. After successful authentication, the session string will be automatically printed to the console.

Example output after first login:
```
================================================================================
🔑 Session authenticated successfully!
================================================================================

For cloud deployments, add this to your environment:

TG_USER_SESSION=1AsCoAAEBu2FhYWFh...YWFhYWFhYWFhYWE=

Session size: 353 characters

================================================================================
```

### Authentication

**Without Session String:**
- You'll need to enter your 2FA code **every time** the service restarts
- The session is kept in memory during runtime only

**With Session String (Recommended):**
1. Run the app locally first to authenticate with 2FA
2. Copy the `TG_USER_SESSION` value from the console output
3. Add it to your `.env` file or deployment environment variables
4. Future restarts will use the session string (no 2FA required)

**Session Format:**
- Uses Telethon's StringSession format (~350 characters)
- Contains: DC ID, IP, Port, and Auth Key
- Safe to use in environment variables
- Compatible across Telegram libraries

## Storage

Forwarding rules are stored in MongoDB and managed exclusively via the HTTP API. The service connects to MongoDB using the `MONGODB_URI` environment variable.

**MongoDB Setup:**
```bash
# Local development (using Docker)
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Or use MongoDB Atlas (free tier available)
# https://www.mongodb.com/cloud/atlas
```

Add to your `.env`:
```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=tg-forward
```

## Managing Rules

### Web Admin Panel (Recommended)

Access the admin panel at `http://localhost:8080/admin`

**Features:**
- 🎨 Modern, responsive UI built with HTMX and Tailwind CSS
- 🔒 Secure token-based authentication (uses your API_TOKEN)
- ➕ Add, edit, and delete rules with a visual interface
- 🏷️ Support for pattern-based, keyword-based, and mixed rules
- ⚡ Real-time updates without page reloads
- 📱 Mobile-friendly design

**Usage:**
1. Navigate to `http://localhost:8080/admin` in your browser
2. Enter your `API_TOKEN` from your `.env` file
3. Manage rules with the visual interface

### API Usage (Alternative)

All endpoints except `/health` and `/admin` require authentication:
```bash
Authorization: Bearer your-secret-api-token
```

**Response Format:** All successful responses are wrapped in `{"data": {...}}`.

### Get Rules
```bash
curl http://localhost:8080/rules \
  -H "Authorization: Bearer your-token"
```

Response:
```json
{
  "data": {
    "rules": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "Urgent Messages",
        "pattern": "urgent.*"
      },
      {
        "id": "650e8400-e29b-41d4-a716-446655440001",
        "name": "Important Keywords",
        "keywords": ["important", "critical"]
      }
    ]
  }
}
```

### Add Rule

**Pattern-based rule:**
```bash
curl -X POST http://localhost:8080/rules/add \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Emergency Messages", "pattern": "emergency.*"}'
```

**Keyword-based rule (all keywords must be present):**
```bash
curl -X POST http://localhost:8080/rules/add \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Payment Alert", "keywords": ["payment", "received"]}'
```

**Mixed rule (pattern OR keywords):**
```bash
curl -X POST http://localhost:8080/rules/add \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Alerts", "pattern": "alert.*", "keywords": ["urgent", "critical"]}'
```

Response:
```json
{
  "data": {
    "rule": {
      "id": "850e8400-e29b-41d4-a716-446655440003",
      "name": "Emergency Messages",
      "pattern": "emergency.*"
    }
  }
}
```

### Update All Rules
```bash
curl -X PUT http://localhost:8080/rules \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "rules": [
      {"id": "550e8400-e29b-41d4-a716-446655440000", "name": "Pattern Rule", "pattern": "urgent.*"},
      {"id": "650e8400-e29b-41d4-a716-446655440001", "name": "Keyword Rule", "keywords": ["payment", "received"]}
    ]
  }'
```

Response:
```json
{
  "data": {
    "rules": [
      {"id": "550e8400-e29b-41d4-a716-446655440000", "name": "Pattern Rule", "pattern": "urgent.*"},
      {"id": "650e8400-e29b-41d4-a716-446655440001", "name": "Keyword Rule", "keywords": ["payment", "received"]}
    ]
  }
}
```

### Remove Rule
```bash
curl -X DELETE http://localhost:8080/rules/remove \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"id": "550e8400-e29b-41d4-a716-446655440000"}'
```

Response:
```json
{
  "data": {
    "message": "rule deleted successfully"
  }
}
```

### Health Check (No Auth)
```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "data": {
    "status": "ok"
  }
}
```

## Rule Types

### Pattern-based Rules
Matches using regex patterns:
- `{"name": "Urgent", "pattern": "urgent.*"}` - matches text with regex

### Keyword-based Rules
Matches when ALL keywords are present (case-insensitive):
- `{"name": "Payment Alert", "keywords": ["payment", "received"]}` - both keywords must exist

### Mixed Rules
Matches if EITHER pattern OR keywords match:
- `{"name": "Alerts", "pattern": "alert.*", "keywords": ["urgent", "critical"]}` - pattern match OR all keywords present

```regex
[0-9]{6}              # 6-digit codes
^urgent               # Messages starting with "urgent"
(?i)important         # "important" (case-insensitive)
https?://[^\s]+       # URLs
\+?[0-9]{10,15}       # Phone numbers
```

## Development

```bash
# Run tests
make test

# Build
make build

# Docker build
make docker-build
```

## License

MIT
