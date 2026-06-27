document.addEventListener('DOMContentLoaded', () => {
  const statusBadge = document.getElementById('status-badge');
  const statusText = document.getElementById('status-text');
  const lastSync = document.getElementById('last-sync');
  const syncBtn = document.getElementById('sync-btn');

  function updateUI() {
    chrome.runtime.sendMessage({ type: 'GET_STATUS' }, (response) => {
      if (chrome.runtime.lastError) return;
      if (!response) return;

      if (response.connected) {
        statusBadge.className = 'badge connected';
        statusText.innerText = 'Connected';
        statusText.style.color = '#10b981';
      } else {
        statusBadge.className = 'badge disconnected';
        statusText.innerText = 'Disconnected';
        statusText.style.color = '#ef4444';
      }

      if (response.lastSyncTime) {
        const date = new Date(response.lastSyncTime);
        lastSync.innerText = date.toLocaleTimeString();
      } else {
        lastSync.innerText = 'Never';
      }
    });
  }

  // Initial update
  updateUI();

  // Listen for real-time updates from background service worker
  chrome.runtime.onMessage.addListener((msg) => {
    if (msg.type === 'SYNC_UPDATE') {
      updateUI();
    }
  });

  syncBtn.addEventListener('click', () => {
    syncBtn.disabled = true;
    syncBtn.innerText = 'Syncing...';
    chrome.runtime.sendMessage({ type: 'FORCE_SYNC' }, () => {
      setTimeout(() => {
        syncBtn.disabled = false;
        syncBtn.innerText = 'Force Sync Cookies';
        updateUI();
      }, 800);
    });
  });
});
