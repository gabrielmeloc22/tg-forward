# Telegram Message Forwarder

A Golang service that listens to your Telegram messages and forwards messages matching regex patterns to another chat via a bot.

## What It Does

1. Connects to your Telegram account (user account)
2. Listens for incoming messages
3. Matches messages against regex patterns (managed via API or `rules.json`)
4. Forwards matching messages using a bot to a target chat
5. Provides an HTTP API to manage forwarding rules dynamically

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
TG_USER_SESSION_FILE=session.json
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
- `TG_USER_SESSION_FILE`: Path to session file (default: `session.json`)
- `TG_USER_SESSION`: Base64-encoded session string (optional, see Session Management below)
- `TG_BOT_TOKEN`: Your bot token (from @BotFather)
- `TG_BOT_TARGET_CHAT_ID`: Target chat ID to forward messages to
- `TG_BOT_TARGET_USERNAME`: Alternative to chat ID, use username (e.g., `@channel`)
- `API_PORT`: HTTP API port (default: `8080`)
- `API_TOKEN`: Secret token for API authentication

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
  -v $(pwd)/configs:/app/configs \
  tg-forward
```

**First Run:**
On first run, you'll be prompted to enter the 2FA code sent to your Telegram. After successful authentication, the session is saved to `session.json` (or the path specified in `TG_USER_SESSION_FILE`). Subsequent runs will use this session file without requiring 2FA.

### Session Management

**For Fly.io or other cloud deployments**, you can use the `TG_USER_SESSION` environment variable instead of a session file:

1. Run the app locally first to authenticate and generate `session.json`
2. Convert the session file to base64:
   ```bash
   base64 -i session.json
   ```
3. Copy the output and set it as `TG_USER_SESSION` in your deployment environment
4. The app will use the session from the environment variable instead of requiring terminal input

**Note:** Forwarding rules are NOT stored in `.env`. Use the API to manage rules (stored in `rules.json`) or manually create a `rules.json` file with your initial rules:

```json
{
  "rules": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Urgent Messages",
      "pattern": "urgent.*"
    },
    {
      "id": "650e8400-e29b-41d4-a716-446655440001",
      "name": "Important",
      "pattern": "important"
    },
    {
      "id": "750e8400-e29b-41d4-a716-446655440002",
      "name": "6-digit codes",
      "pattern": "[0-9]{6}"
    }
  ]
}
```

## API Usage

All endpoints except `/health` require authentication:
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
        "name": "Important",
        "pattern": "important"
      }
    ]
  }
}
```

### Add Rule
```bash
curl -X POST http://localhost:8080/rules/add \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"name": "Emergency Messages", "pattern": "emergency.*"}'
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
      {"id": "550e8400-e29b-41d4-a716-446655440000", "name": "New Pattern", "pattern": "new.*"},
      {"id": "650e8400-e29b-41d4-a716-446655440001", "name": "Test", "pattern": "test"}
    ]
  }'
```

Response:
```json
{
  "data": {
    "rules": [
      {"id": "550e8400-e29b-41d4-a716-446655440000", "name": "New Pattern", "pattern": "new.*"},
      {"id": "650e8400-e29b-41d4-a716-446655440001", "name": "Test", "pattern": "test"}
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
    "message": "rule removed"
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
    "status": "healthy"
  }
}
```

## Common Regex Patterns

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
