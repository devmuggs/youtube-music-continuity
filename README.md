# YouTube Music Continuity

A monorepo project that enables seamless music playback continuity across YouTube Music sessions using a Chrome extension and Go backend API.

## Features

- **Cross-device synchronization**: Resume your music exactly where you left off
- **Real-time updates**: Uses WebSocket connections for instant playback state synchronization
- **Secure authentication**: Email/password based user system
- **Chrome extension**: Seamlessly integrates with YouTube Music interface

## Architecture

This project consists of two main components:

### 1. API Backend (Go)
- User registration and authentication system
- WebSocket server for real-time communication
- Stores playback state (video ID, timestamp) for each user
- RESTful API endpoints for session management

### 2. Chrome Extension
- Injects into YouTube Music web interface
- Tracks play/pause events and current playback position
- Provides "Save Session" and "Resume Session" functionality
- Real-time synchronization with backend via WebSocket

## Project Structure

```
├── backend/          # Go API backend
│   ├── cmd/         # Application entrypoints
│   ├── internal/    # Private application code
│   ├── pkg/         # Public libraries
│   └── config/      # Configuration files
├── extension/       # Chrome extension
│   ├── manifest.json
│   ├── content.js   # YouTube Music injection
│   ├── popup/       # Extension popup UI
│   └── background/  # Background scripts
└── docs/           # Documentation
```

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Node.js 18+ (for extension development)
- Chrome browser for testing

### Backend Setup
```bash
cd backend
go mod tidy
go run cmd/server/main.go
```

### Extension Setup
1. Open Chrome and navigate to `chrome://extensions/`
2. Enable "Developer mode"
3. Click "Load unpacked" and select the `extension/` directory

## Development

### Run Backend
```bash
make backend-dev
```

### Build Extension
```bash
make extension-build
```

### Run All
```bash
make dev
```

## API Endpoints

- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `GET /api/session` - Get current playback session
- `PUT /api/session` - Update playback session
- `WS /api/ws` - WebSocket connection for real-time updates

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.