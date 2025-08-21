# Chrome Extension Setup Guide

## Installation (Development)

1. **Clone the repository**:
   ```bash
   git clone https://github.com/devmuggs/youtube-music-continuity.git
   cd youtube-music-continuity
   ```

2. **Start the backend server**:
   ```bash
   make backend-dev
   # OR
   cd backend && go run cmd/server/main.go
   ```

3. **Load the extension in Chrome**:
   - Open Chrome and navigate to `chrome://extensions/`
   - Enable "Developer mode" (toggle in top right)
   - Click "Load unpacked"
   - Select the `extension/` directory from this repository

## Usage

### First Time Setup

1. **Navigate to YouTube Music**:
   - Go to [https://music.youtube.com](https://music.youtube.com)

2. **Open the extension popup**:
   - Click the extension icon in the Chrome toolbar
   - You'll see the login/register form

3. **Create an account or sign in**:
   - Register with email/password or login with existing credentials
   - The extension will connect to the backend server

### Saving Sessions

1. **Play any song on YouTube Music**
2. **Click the 💾 button** that appears in the player controls
3. **Your session is saved** - video ID, timestamp, and play state

### Resuming Sessions

1. **Open a new tab** and go to YouTube Music
2. **If you have a saved session**, you'll see a **▶️ Resume Session** button
3. **Click to resume** - you'll be redirected to the exact song and timestamp

### Extension Features

- **Automatic session tracking**: The extension monitors your playback automatically
- **Cross-tab synchronization**: Sessions sync across all your YouTube Music tabs
- **Real-time updates**: Uses WebSocket for instant synchronization
- **Secure authentication**: JWT-based login system

## Troubleshooting

### Backend Connection Issues

If you see "Server unavailable" in the extension popup:

1. **Check if the backend is running**:
   ```bash
   curl http://localhost:8080/api/session
   # Should return 401 Unauthorized (which is expected without auth)
   ```

2. **Restart the backend**:
   ```bash
   make backend-dev
   ```

3. **Check firewall settings**: Ensure port 8080 is not blocked

### Extension Not Working

1. **Reload the extension**:
   - Go to `chrome://extensions/`
   - Click the reload button on the YouTube Music Continuity extension

2. **Check the console**:
   - Right-click on YouTube Music page → Inspect → Console
   - Look for any error messages from the extension

3. **Verify permissions**:
   - The extension needs access to `music.youtube.com`
   - Check that permissions are granted in `chrome://extensions/`

### WebSocket Connection Issues

1. **Check browser console** for WebSocket errors
2. **Verify JWT token** is valid (try logging out and back in)
3. **Restart the backend** server

## Development

### Building for Distribution

```bash
make extension-build
```

This creates `youtube-music-continuity-extension.zip` ready for Chrome Web Store.

### Testing Changes

1. **Make your changes** to the extension files
2. **Reload the extension** in `chrome://extensions/`
3. **Refresh YouTube Music** tab to test changes

### Extension Architecture

- **`manifest.json`**: Extension configuration
- **`content.js`**: Injected into YouTube Music pages
- **`popup/`**: Extension popup UI (HTML/CSS/JS)
- **`background/`**: Background service worker
- **`icons/`**: Extension icons (16px, 32px, 48px, 128px)

## Security Notes

- **JWT tokens** are stored securely in Chrome's local storage
- **Passwords** are hashed with bcrypt on the backend
- **WebSocket connections** use token-based authentication
- **CORS** is configured to only allow Chrome extension origins