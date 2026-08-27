/**
 * Free Gemini API Sync — Background Service Worker
 * Proactively keeps Google Gemini cookies fresh and auto-syncs them on change.
 */

const LOCAL_WS_URL = 'ws://127.0.0.1:9226';
let ws = null;
let lastSyncTime = null;
let hasSyncedOnce = false;
let syncDebounceTimeout = null;
let isRefreshingTab = false;

chrome.runtime.onInstalled.addListener(init);
chrome.runtime.onStartup.addListener(init);

chrome.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name === 'reconnect') connectToBackend();
  if (alarm.name === 'keepAlive') keepAlive();
  if (alarm.name === 'sessionKeepAlive') {
    console.log('[Gemini Sync] Running periodic session keep-alive refresh...');
    ensureGeminiTabAndSync(true); // Quietly refresh session
  }
});

async function init() {
  connectToBackend();
  // Keep-alive ping every 25 seconds (just ping, no cookies)
  chrome.alarms.create('keepAlive', { periodInMinutes: 0.4 });
  // Proactively refresh Gemini session tab every 15 minutes to rotate cookies
  chrome.alarms.create('sessionKeepAlive', { periodInMinutes: 15 });
  
  const data = await chrome.storage.local.get(['lastSyncTime']);
  if (data.lastSyncTime) lastSyncTime = data.lastSyncTime;
}

function connectToBackend() {
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) {
    return;
  }

  console.log('[Gemini Sync] Connecting to local backend at:', LOCAL_WS_URL);
  hasSyncedOnce = false;

  try {
    ws = new WebSocket(LOCAL_WS_URL);
  } catch (e) {
    console.error('[Gemini Sync] WS Connection Error:', e);
    scheduleReconnect();
    return;
  }

  ws.onopen = () => {
    console.log('[Gemini Sync] Connected to Go Backend!');
    chrome.alarms.clear('reconnect');
    // Sync ONCE on connect
    performSync();
  };

  ws.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data);
      if (msg.type === 'trigger_sync') {
        console.log('[Gemini Sync] Backend requested fresh cookies. Activating session refresh...');
        ensureGeminiTabAndSync(false);
      }
    } catch (e) {
      console.error(e);
    }
  };

  ws.onclose = () => {
    console.log('[Gemini Sync] Connection closed. Reconnecting...');
    scheduleReconnect();
  };

  ws.onerror = (err) => {
    console.error('[Gemini Sync] WebSocket Error:', err);
  };
}

function scheduleReconnect() {
  chrome.alarms.create('reconnect', { delayInMinutes: 0.083 }); // ~5s
}

function keepAlive() {
  if (ws?.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'ping' }));
  } else {
    connectToBackend();
  }
}

const ALLOWED_COOKIE_NAMES = new Set([
  '__Secure-1PSID',
  '__Secure-3PSID',
  '__Secure-1PAPISID',
  '__Secure-3PAPISID',
  '__Secure-1PSIDTS',
  '__Secure-3PSIDTS',
  '__Secure-1PSIDCC',
  '__Secure-3PSIDCC',
  'SID',
  'HSID',
  'SSID',
  'APISID',
  'SAPISID',
  'SIDCC',
  'OSID',
  '__Secure-OSID'
]);

// Performs the actual extraction and WebSocket transfer
function performSync() {
  if (!ws || ws.readyState !== WebSocket.OPEN) return;

  chrome.cookies.getAll({}, (cookies) => {
    const googleCookies = cookies.filter(c => 
      (c.domain === '.google.com' || c.domain === '.google.co' || c.domain === 'gemini.google.com' || c.domain.includes('googleusercontent.com')) &&
      (ALLOWED_COOKIE_NAMES.has(c.name) || c.domain.includes('googleusercontent.com'))
    );

    const formatted = googleCookies.map(c => {
      let exp = c.expirationDate;
      if (!exp || c.session) {
        exp = Math.floor(Date.now() / 1000) + 31536000; // Default 1 year
      }

      let sameSite = c.sameSite || 'unspecified';
      if (sameSite === 'no_restriction') sameSite = 'none';

      return {
        domain: c.domain,
        expirationDate: exp,
        hostOnly: c.hostOnly,
        httpOnly: c.httpOnly,
        name: c.name,
        path: c.path,
        sameSite: sameSite,
        secure: c.secure,
        session: c.session,
        storeId: c.storeId || '0',
        value: c.value
      };
    });

    console.log(`[Gemini Sync] Syncing ${formatted.length} essential cookies to backend`);
    ws.send(JSON.stringify({
      type: 'cookies_payload',
      cookies: formatted
    }));

    hasSyncedOnce = true;
    lastSyncTime = Date.now();
    chrome.storage.local.set({ lastSyncTime });
    chrome.runtime.sendMessage({ type: 'SYNC_UPDATE', success: true, count: formatted.length }).catch(() => {});
  });
}

// Proactive refresh mechanism: opens or reloads Gemini tab in background to force cookie rotation
async function ensureGeminiTabAndSync(quietMode = false) {
  if (isRefreshingTab) return;
  isRefreshingTab = true;

  try {
    const tabs = await chrome.tabs.query({ url: '*://gemini.google.com/*' });
    
    if (tabs.length > 0) {
      console.log('[Gemini Sync] Gemini tab exists. Reloading to rotate cookies...');
      await chrome.tabs.reload(tabs[0].id);
      if (!quietMode) {
        // Force focus on the existing tab to wake it up and ensure quick reload/sync
        await chrome.tabs.update(tabs[0].id, { active: true });
        
        // Also bring the window to the front if minimized
        const tab = tabs[0];
        if (tab.windowId) {
          await chrome.windows.update(tab.windowId, { focused: true });
        }
      }
    } else {
      console.log('[Gemini Sync] No Gemini tab found. Launching a session...');
      // Open in foreground (active: true) if not quietMode to guarantee immediate load
      await chrome.tabs.create({ url: 'https://gemini.google.com', active: !quietMode });
    }

    // Safety timeout to reset the refreshing flag if something hangs
    setTimeout(() => {
      isRefreshingTab = false;
    }, 15000);

  } catch (e) {
    console.error('[Gemini Sync] Error during session tab refresh:', e);
    isRefreshingTab = false;
    performSync(); // Fallback sync
  }
}

// ─── Tab Load Listener ───────────────────────────────────────────────
// Listen for Gemini tab loads/reloads to capture fresh cookies automatically
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete' && tab.url && tab.url.includes('gemini.google.com')) {
    console.log('[Gemini Sync] Gemini tab loaded completely. Performing sync...');
    performSync();
    isRefreshingTab = false;
  }
});


// ─── Real-Time Cookie Listener ──────────────────────────────────────
// Auto-detect when Google changes or rotates session cookies and push them instantly!
chrome.cookies.onChanged.addListener((changeInfo) => {
  const cookie = changeInfo.cookie;
  
  const isTargetDomain = cookie.domain === '.google.com' || 
                         cookie.domain === '.google.co' || 
                         cookie.domain === 'gemini.google.com' ||
                         cookie.domain.includes('googleusercontent.com');

  if (isTargetDomain && ALLOWED_COOKIE_NAMES.has(cookie.name)) {
    // Skip if it was deleted (unless it is a known rotation)
    if (changeInfo.removed) return;

    console.log(`[Gemini Sync] Real-time cookie updated: ${cookie.name}. Scheduling sync...`);

    // Debounce multiple fast updates (since Google updates multiple cookies together)
    if (syncDebounceTimeout) clearTimeout(syncDebounceTimeout);
    syncDebounceTimeout = setTimeout(() => {
      console.log('[Gemini Sync] Running debounced real-time cookie sync...');
      performSync();
    }, 1500);
  }
});

// Receive message from Popup
chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === 'GET_STATUS') {
    sendResponse({
      connected: ws && ws.readyState === WebSocket.OPEN,
      lastSyncTime,
      hasSyncedOnce
    });
  }
  if (msg.type === 'FORCE_SYNC') {
    ensureGeminiTabAndSync(false);
    sendResponse({ ok: true });
  }
  return true;
});
