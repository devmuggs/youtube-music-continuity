# API Documentation

## Base URL
`http://localhost:8080/api`

## Authentication
All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

## Endpoints

### Authentication

#### Register User
- **POST** `/auth/register`
- **Body**: `{"email": "string", "password": "string"}`
- **Response**: `{"token": "string", "user": {...}}`

#### Login User
- **POST** `/auth/login`
- **Body**: `{"email": "string", "password": "string"}`
- **Response**: `{"token": "string", "user": {...}}`

### Session Management

#### Get Current Session
- **GET** `/session`
- **Headers**: `Authorization: Bearer <token>`
- **Response**: `{"id": int, "user_id": int, "video_id": "string", "playback_time": float, "is_playing": bool, "updated": "timestamp"}`

#### Update Session
- **PUT** `/session`
- **Headers**: `Authorization: Bearer <token>`
- **Body**: `{"video_id": "string", "playback_time": float, "is_playing": bool}`
- **Response**: `{"id": int, "user_id": int, "video_id": "string", "playback_time": float, "is_playing": bool, "updated": "timestamp"}`

### WebSocket

#### Connect to WebSocket
- **WS** `/ws?token=<jwt_token>`
- **Messages**: JSON format with `{"type": "string", "payload": {}}`

#### Message Types
- `session_update`: Update playback session
- `auth`: Authenticate WebSocket connection (if token not in query)

## Example Usage

### Register and Login
```bash
# Register
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Session Management
```bash
# Get session
curl -X GET http://localhost:8080/api/session \
  -H "Authorization: Bearer <token>"

# Update session
curl -X PUT http://localhost:8080/api/session \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"video_id":"dQw4w9WgXcQ","playback_time":120.5,"is_playing":true}'
```

### WebSocket Connection
```javascript
const ws = new WebSocket('ws://localhost:8080/api/ws?token=' + authToken);

// Send session update
ws.send(JSON.stringify({
  type: 'session_update',
  payload: {
    video_id: 'dQw4w9WgXcQ',
    playback_time: 120.5,
    is_playing: true
  }
}));
```