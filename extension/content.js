// Content script for YouTube Music Continuity
(function () {
	"use strict";

	const API_BASE_URL = "http://localhost:8080/api";
	let wsConnection = null;
	let currentUser = null;
	let sessionSaveButton = null;
	let sessionResumeButton = null;

	let lastVideoId = "";

	let lastKnownState = {
		videoId: "",
		currentTime: 0,
		isPlaying: false
	};

	// Initialize the extension
	function init() {
		checkAuthStatus();
		createUI();
		setupPlayerListeners();
		connectWebSocket();
	}

	// Check if user is authenticated
	async function checkAuthStatus() {
		try {
			const result = await chrome.storage.local.get(["authToken", "user"]);
			if (result.authToken && result.user) {
				currentUser = result.user;
				console.log("User authenticated:", currentUser.email);
				checkForResumableSession();
			}
		} catch (error) {
			console.error("Error checking auth status:", error);
		}
	}

	// Create UI elements
	function createUI() {
		// Wait for the page to load
		setTimeout(() => {
			createSessionButtons();
		}, 2000);
	}

	// Create session save/resume buttons
	function createSessionButtons() {
		const rightControls = document.querySelector(".right-controls-buttons");
		if (!rightControls) {
			console.log("Right controls not found, retrying...");
			setTimeout(createSessionButtons, 1000);
			return;
		}

		// Create save session button
		if (!sessionSaveButton) {
			sessionSaveButton = document.createElement("button");
			sessionSaveButton.innerHTML = "💾";
			sessionSaveButton.title = "Save Session";
			sessionSaveButton.style.cssText = `
        background: none;
        border: none;
        color: var(--yt-spec-text-primary);
        font-size: 20px;
        padding: 8px;
        cursor: pointer;
        margin-left: 8px;
        border-radius: 4px;
        display: ${currentUser ? "block" : "none"};
      `;
			sessionSaveButton.addEventListener("click", () => saveSession());
			rightControls.appendChild(sessionSaveButton);
		}

		// Create resume session button (only show when resumable session exists)
		if (!sessionResumeButton && currentUser) {
			sessionResumeButton = document.createElement("button");
			sessionResumeButton.innerHTML = "▶️";
			sessionResumeButton.title = "Resume Session";
			sessionResumeButton.style.cssText = `
        background: #ff0000;
        border: none;
        color: white;
        font-size: 16px;
        padding: 8px 12px;
        cursor: pointer;
        margin-left: 8px;
        border-radius: 4px;
        display: none;
      `;
			sessionResumeButton.addEventListener("click", resumeSession);
			rightControls.appendChild(sessionResumeButton);
		}
	}

	// Setup player event listeners
	function setupPlayerListeners() {
		// Monitor for player state changes
		setInterval(checkPlayerState, 1000);
	}

	// Check current player state
	function checkPlayerState() {
		if (!currentUser) return;

		const video = document.querySelector("video");
		if (!video) return;

		const videoId = extractVideoId();
		const currentTime = video.currentTime;
		const isPlaying = !video.paused;

		// Only update if state has changed significantly
		if (
			videoId !== lastKnownState.videoId ||
			Math.abs(currentTime - lastKnownState.currentTime) > 2 ||
			isPlaying !== lastKnownState.isPlaying
		) {
			lastKnownState = { videoId, currentTime, isPlaying };

			if (videoId) {
				// Send update via WebSocket if connected
				if (wsConnection && wsConnection.readyState === WebSocket.OPEN) {
					sendSessionUpdate(videoId, currentTime, isPlaying);
				}
			}
		}
	}

	// Extract video ID from current URL
	function extractVideoId() {
		// Try URL first
		const url = window.location.href;
		const match = url.match(/[?&]v=([^&]+)/);
		if (match) return match[1];

		// Try to get from player bar (YouTube Music often stores it here)
		const playerBar = document.querySelector("ytmusic-player-bar");
		if (playerBar && playerBar.hasAttribute("video-id")) {
			return playerBar.getAttribute("video-id");
		}

		// Try to get from video element src
		const video = document.querySelector("video");
		console.log({ videoSrc: video.src });
		if (video && video.src) {
			const srcMatch = video.src.match(/\/([a-zA-Z0-9_-]{11})/);
			if (srcMatch) return srcMatch[1];
		}

		// Try to get from other data attributes
		const dataVideoIdElem = document.querySelector("[data-video-id]");
		if (dataVideoIdElem) return dataVideoIdElem.getAttribute("data-video-id");

		return "";
	}

	// Connect to WebSocket
	async function connectWebSocket() {
		if (!currentUser) return;

		try {
			const result = await chrome.storage.local.get(["authToken"]);
			if (!result.authToken) return;

			wsConnection = new WebSocket(`ws://localhost:8080/api/ws?token=${result.authToken}`);

			wsConnection.onopen = () => {
				console.log("WebSocket connected");
			};

			wsConnection.onmessage = (event) => {
				const message = JSON.parse(event.data);
				handleWebSocketMessage(message);
			};

			wsConnection.onclose = () => {
				console.log("WebSocket disconnected");
				// Reconnect after delay
				setTimeout(connectWebSocket, 5000);
			};

			wsConnection.onerror = (error) => {
				console.error("WebSocket error:", error);
			};
		} catch (error) {
			console.error("Error connecting WebSocket:", error);
		}
	}

	// Handle WebSocket messages
	function handleWebSocketMessage(message) {
		if (message.type === "session_update") {
			// Handle session updates from other devices
			console.log("Session update received:", message.payload);
		}
	}

	// Send session update via WebSocket
	function sendSessionUpdate(videoId, playbackTime, isPlaying) {
		if (wsConnection && wsConnection.readyState === WebSocket.OPEN) {
			console.debug(`Sending session update: ${videoId}, ${playbackTime}, ${isPlaying}`);

			if (!videoId || playbackTime < 1) {
				console.debug("Invalid session data, not sending update.");
				return;
			}

			const message = {
				type: "session_update",
				payload: {
					video_id: videoId,
					playback_time: playbackTime,
					is_playing: isPlaying
				}
			};

			wsConnection.send(JSON.stringify(message));
		}
	}

	// Save current session
	async function saveSession(videoId) {
		if (!currentUser) {
			alert("Please sign in to save your session");
			return;
		}

		if (!videoId) videoId = extractVideoId();
		const video = document.querySelector("video");

		if (!video || !videoId) {
			alert("No video currently playing");
			return;
		}

		const sessionData = {
			video_id: videoId,
			playback_time: video.currentTime,
			is_playing: !video.paused
		};

		if (!sessionData.is_playing || sessionData.playback_time < 1 || !sessionData.video_id) {
			return;
		}

		try {
			const result = await chrome.storage.local.get(["authToken"]);
			const response = await fetch(`${API_BASE_URL}/session`, {
				method: "PUT",
				headers: {
					"Content-Type": "application/json",
					Authorization: `Bearer ${result.authToken}`
				},
				body: JSON.stringify(sessionData)
			});

			if (response.ok) {
				// Show success feedback
				sessionSaveButton.innerHTML = "✅";
				setTimeout(() => {
					sessionSaveButton.innerHTML = "💾";
				}, 2000);
			} else {
				throw new Error("Failed to save session");
			}
		} catch (error) {
			console.error("Error saving session:", error);
			alert("Failed to save session");
		}
	}

	// Check for resumable session on page load
	async function checkForResumableSession() {
		try {
			const result = await chrome.storage.local.get(["authToken"]);
			const response = await fetch(`${API_BASE_URL}/session`, {
				method: "GET",
				headers: {
					Authorization: `Bearer ${result.authToken}`
				}
			});

			if (response.ok) {
				const session = await response.json();
				if (session && session.video_id && session.video_id !== extractVideoId()) {
					showResumeButton(session);
				}
			}
		} catch (error) {
			console.error("Error checking for resumable session:", error);
		}
	}

	// Show resume button
	function showResumeButton(session) {
		if (sessionResumeButton) {
			sessionResumeButton.style.display = "block";
			sessionResumeButton.onclick = () => resumeSession(session);
		}
	}

	// Resume session
	function resumeSession(session) {
		if (!session) {
			// Fetch latest session
			checkForResumableSession();
			return;
		}

		const videoUrl = `https://music.youtube.com/watch?v=${session.video_id}&t=${Math.floor(
			session.playback_time
		)}s`;
		window.location.href = videoUrl;
	}

	// Listen for auth state changes from popup
	chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
		if (request.type === "AUTH_STATUS_CHANGED") {
			if (request.isAuthenticated) {
				currentUser = request.user;
				connectWebSocket();
				if (sessionSaveButton) {
					sessionSaveButton.style.display = "block";
				}
				checkForResumableSession();
			} else {
				currentUser = null;
				if (wsConnection) {
					wsConnection.close();
				}
				if (sessionSaveButton) {
					sessionSaveButton.style.display = "none";
				}
				if (sessionResumeButton) {
					sessionResumeButton.style.display = "none";
				}
			}
		}
	});

	// Initialize when DOM is ready
	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", init);
	} else {
		init();
	}
})();
