# Web Admin Panel

A single-file, self-contained admin interface for managing TG Forward rules.

## Features

- **Single HTML File**: All HTML, CSS, and JavaScript in one file
- **HTMX**: Reactive updates without page reloads
- **Tailwind CSS**: Modern, responsive design via CDN
- **Secure**: Uses your existing API token for authentication
- **Full CRUD**: Create, read, update, and delete rules
- **Rule Types**: Support for pattern-based, keyword-based, and mixed rules

## Access

Navigate to: `http://localhost:8080/admin`

Enter your `API_TOKEN` from your `.env` file to login.

## Technology Stack

- **HTMX 1.9.10**: For reactive, server-driven interactions
- **Tailwind CSS**: For styling (loaded via CDN)
- **Vanilla JavaScript**: No framework dependencies
- **Session Storage**: For token persistence (cleared on logout)

## Security

- Token stored in browser's sessionStorage (cleared on logout)
- All API calls use Bearer token authentication
- No direct MongoDB access - uses your existing API
- Admin panel route (`/admin`) is public, but all data operations require authentication

## How It Works

1. **Authentication**: User enters API token → stored in sessionStorage
2. **Load Rules**: Fetches from `/rules` endpoint with Bearer token
3. **Add/Edit**: POST to `/rules/add` or PUT to `/rules` with token
4. **Delete**: DELETE to `/rules/remove` with token
5. **Reactive UI**: HTMX handles DOM updates without page reloads

## File Structure

```
web/
  admin.html   # Single-file admin panel (15KB)
  README.md    # This file
```

## Development

The admin panel uses your existing API endpoints:
- `GET /rules` - List all rules
- `POST /rules/add` - Add new rule
- `PUT /rules` - Update all rules (used for editing)
- `DELETE /rules/remove` - Remove rule

All responses follow the format: `{"data": {...}}`
