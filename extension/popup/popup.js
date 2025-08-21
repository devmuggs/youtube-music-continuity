// Popup script for YouTube Music Continuity
(function() {
  'use strict';

  const API_BASE_URL = 'http://localhost:8080/api';
  let isLoginMode = true;
  let currentUser = null;

  // DOM elements
  const authContainer = document.getElementById('authContainer');
  const userContainer = document.getElementById('userContainer');
  const authForm = document.getElementById('authForm');
  const emailInput = document.getElementById('email');
  const passwordInput = document.getElementById('password');
  const authButton = document.getElementById('authButton');
  const toggleAuth = document.getElementById('toggleAuth');
  const errorMessage = document.getElementById('errorMessage');
  const successMessage = document.getElementById('successMessage');
  const userEmail = document.getElementById('userEmail');
  const statusIndicator = document.getElementById('statusIndicator');
  const connectionStatus = document.getElementById('connectionStatus');
  const sessionInfo = document.getElementById('sessionInfo');
  const sessionDetails = document.getElementById('sessionDetails');
  const saveSessionBtn = document.getElementById('saveSessionBtn');
  const resumeSessionBtn = document.getElementById('resumeSessionBtn');
  const signOutBtn = document.getElementById('signOutBtn');

  // Initialize popup
  async function init() {
    setupEventListeners();
    await checkAuthStatus();
  }

  // Setup event listeners
  function setupEventListeners() {
    authForm.addEventListener('submit', handleAuthSubmit);
    toggleAuth.addEventListener('click', toggleAuthMode);
    saveSessionBtn.addEventListener('click', saveSession);
    resumeSessionBtn.addEventListener('click', resumeSession);
    signOutBtn.addEventListener('click', signOut);
  }

  // Check authentication status
  async function checkAuthStatus() {
    try {
      const result = await chrome.storage.local.get(['authToken', 'user']);
      if (result.authToken && result.user) {
        currentUser = result.user;
        showUserContainer();
        checkConnectionStatus();
        loadSessionInfo();
      } else {
        showAuthContainer();
      }
    } catch (error) {
      console.error('Error checking auth status:', error);
      showAuthContainer();
    }
  }

  // Handle auth form submission
  async function handleAuthSubmit(e) {
    e.preventDefault();
    
    const email = emailInput.value.trim();
    const password = passwordInput.value;

    if (!email || !password) {
      showError('Please fill in all fields');
      return;
    }

    setLoading(true);
    hideMessages();

    try {
      const endpoint = isLoginMode ? '/auth/login' : '/auth/register';
      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password })
      });

      const data = await response.json();

      if (response.ok) {
        // Store auth data
        await chrome.storage.local.set({
          authToken: data.token,
          user: data.user
        });

        currentUser = data.user;
        showSuccess(isLoginMode ? 'Signed in successfully!' : 'Account created successfully!');
        
        // Notify content script
        notifyContentScript();
        
        setTimeout(() => {
          showUserContainer();
          checkConnectionStatus();
          loadSessionInfo();
        }, 1000);
      } else {
        showError(data.message || 'Authentication failed');
      }
    } catch (error) {
      console.error('Auth error:', error);
      showError('Network error. Please check if the server is running.');
    } finally {
      setLoading(false);
    }
  }

  // Toggle between login and register
  function toggleAuthMode(e) {
    e.preventDefault();
    isLoginMode = !isLoginMode;
    
    if (isLoginMode) {
      authButton.textContent = 'Sign In';
      toggleAuth.textContent = "Don't have an account? Register";
    } else {
      authButton.textContent = 'Register';
      toggleAuth.textContent = 'Already have an account? Sign In';
    }
    
    hideMessages();
  }

  // Check connection status
  async function checkConnectionStatus() {
    try {
      const result = await chrome.storage.local.get(['authToken']);
      const response = await fetch(`${API_BASE_URL}/session`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${result.authToken}`
        }
      });

      if (response.ok) {
        setConnectionStatus(true);
      } else {
        setConnectionStatus(false);
      }
    } catch (error) {
      console.error('Connection check error:', error);
      setConnectionStatus(false);
    }
  }

  // Load session information
  async function loadSessionInfo() {
    try {
      const result = await chrome.storage.local.get(['authToken']);
      const response = await fetch(`${API_BASE_URL}/session`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${result.authToken}`
        }
      });

      if (response.ok) {
        const session = await response.json();
        if (session && session.video_id) {
          showSessionInfo(session);
        } else {
          hideSessionInfo();
        }
      }
    } catch (error) {
      console.error('Error loading session info:', error);
      hideSessionInfo();
    }
  }

  // Save current session
  async function saveSession() {
    try {
      // Get current tab
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      
      if (!tab.url.includes('music.youtube.com')) {
        showError('Please navigate to YouTube Music to save a session');
        return;
      }

      // Send message to content script to save session
      chrome.tabs.sendMessage(tab.id, { type: 'SAVE_SESSION' });
      showSuccess('Session save requested');
      
      // Refresh session info after a delay
      setTimeout(loadSessionInfo, 2000);
    } catch (error) {
      console.error('Error saving session:', error);
      showError('Failed to save session');
    }
  }

  // Resume session
  async function resumeSession() {
    try {
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      
      // Send message to content script to resume session
      chrome.tabs.sendMessage(tab.id, { type: 'RESUME_SESSION' });
    } catch (error) {
      console.error('Error resuming session:', error);
      showError('Failed to resume session');
    }
  }

  // Sign out
  async function signOut() {
    try {
      await chrome.storage.local.remove(['authToken', 'user']);
      currentUser = null;
      
      // Notify content script
      notifyContentScript();
      
      showAuthContainer();
      clearForm();
      showSuccess('Signed out successfully');
    } catch (error) {
      console.error('Error signing out:', error);
    }
  }

  // Notify content script of auth state change
  function notifyContentScript() {
    chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
      if (tabs[0] && tabs[0].url.includes('music.youtube.com')) {
        chrome.tabs.sendMessage(tabs[0].id, {
          type: 'AUTH_STATUS_CHANGED',
          isAuthenticated: !!currentUser,
          user: currentUser
        });
      }
    });
  }

  // UI helper functions
  function showAuthContainer() {
    authContainer.style.display = 'block';
    userContainer.style.display = 'none';
  }

  function showUserContainer() {
    authContainer.style.display = 'none';
    userContainer.style.display = 'block';
    userEmail.textContent = currentUser?.email || '';
  }

  function setConnectionStatus(connected) {
    if (connected) {
      statusIndicator.className = 'status-indicator status-connected';
      connectionStatus.textContent = 'Connected to server';
    } else {
      statusIndicator.className = 'status-indicator status-disconnected';
      connectionStatus.textContent = 'Server unavailable';
    }
  }

  function showSessionInfo(session) {
    const timeStr = Math.floor(session.playback_time / 60) + ':' + 
                   String(Math.floor(session.playback_time % 60)).padStart(2, '0');
    
    sessionDetails.innerHTML = `
      <div><strong>Video ID:</strong> ${session.video_id}</div>
      <div><strong>Time:</strong> ${timeStr}</div>
      <div><strong>Status:</strong> ${session.is_playing ? 'Playing' : 'Paused'}</div>
      <div><strong>Updated:</strong> ${new Date(session.updated).toLocaleTimeString()}</div>
    `;
    sessionInfo.style.display = 'block';
    resumeSessionBtn.style.display = 'block';
  }

  function hideSessionInfo() {
    sessionInfo.style.display = 'none';
    resumeSessionBtn.style.display = 'none';
  }

  function setLoading(loading) {
    authButton.disabled = loading;
    authButton.textContent = loading ? 'Please wait...' : (isLoginMode ? 'Sign In' : 'Register');
  }

  function showError(message) {
    errorMessage.textContent = message;
    errorMessage.style.display = 'block';
    successMessage.style.display = 'none';
  }

  function showSuccess(message) {
    successMessage.textContent = message;
    successMessage.style.display = 'block';
    errorMessage.style.display = 'none';
  }

  function hideMessages() {
    errorMessage.style.display = 'none';
    successMessage.style.display = 'none';
  }

  function clearForm() {
    emailInput.value = '';
    passwordInput.value = '';
    hideMessages();
  }

  // Initialize when DOM is loaded
  document.addEventListener('DOMContentLoaded', init);
})();