// Background script for YouTube Music Continuity
(function() {
  'use strict';

  // Listen for extension installation
  chrome.runtime.onInstalled.addListener(() => {
    console.log('YouTube Music Continuity extension installed');
  });

  // Listen for tab updates to detect YouTube Music navigation
  chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    if (changeInfo.status === 'complete' && tab.url && tab.url.includes('music.youtube.com')) {
      // Check if user is authenticated and inject content script if needed
      checkAuthAndInject(tabId);
    }
  });

  // Listen for messages from content script and popup
  chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.type === 'BACKGROUND_PING') {
      sendResponse({ status: 'ok' });
    }
  });

  // Check authentication status and inject content script
  async function checkAuthAndInject(tabId) {
    try {
      const result = await chrome.storage.local.get(['authToken', 'user']);
      if (result.authToken && result.user) {
        // User is authenticated, make sure content script is aware
        chrome.tabs.sendMessage(tabId, {
          type: 'AUTH_STATUS_CHANGED',
          isAuthenticated: true,
          user: result.user
        }, (response) => {
          // Ignore errors if content script is not ready yet
          if (chrome.runtime.lastError) {
            console.log('Content script not ready yet');
          }
        });
      }
    } catch (error) {
      console.error('Error checking auth in background:', error);
    }
  }

  // Handle browser action click (if needed for debugging)
  chrome.action.onClicked.addListener((tab) => {
    if (tab.url && tab.url.includes('music.youtube.com')) {
      console.log('Extension clicked on YouTube Music tab');
    }
  });

  // Cleanup on extension unload
  chrome.runtime.onSuspend.addListener(() => {
    console.log('YouTube Music Continuity extension suspended');
  });
})();