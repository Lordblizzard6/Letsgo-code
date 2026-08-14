<script lang="ts">
  import { useConfig, useTheme } from "../lib/store.svelte";
  import { SettingsService } from "../bindings";

  let tempApiKey = $state("");
  let tempModel = $state("");
  let tempAutoApprove = $state(false);

  $effect(() => {
    const cfg = useConfig.config();
    if (cfg) {
      tempApiKey = cfg.APIKey;
      tempModel = cfg.Model;
      tempAutoApprove = Boolean(cfg.AutoApprove["bash"]);
    }
  });

  function saveConfig() {
    useConfig.setSaving(true);
    const partial: Record<string, unknown> = { Model: tempModel.trim() };
    if (tempApiKey.trim()) partial.APIKey = tempApiKey.trim();
    partial.AutoApprove = { bash: tempAutoApprove };
    SettingsService.SaveConfig(partial)
      .then((c) => {
        useConfig.setConfig(c);
        useConfig.setSaving(false);
      })
      .catch(() => useConfig.setSaving(false));
  }
</script>

<div class="overlay">
  <div class="overlay-panel">
    <div class="panel-title">
      <h2>Settings</h2>
      <button class="close-btn" onclick={() => (window.location.hash = "")}>
        ✕
      </button>
    </div>

    <div class="settings-content">
      {#if useConfig.isSaving()}
        <div class="saving-indicator">
          <div class="spinner"></div>
          <span>Saving...</span>
        </div>
      {:else}
        <form onsubmit={(e) => {
          e.preventDefault();
          saveConfig();
        }}>
          <div class="setting-group">
            <label for="apiKey">API Key</label>
            <input
              id="apiKey"
              type="password"
              bind:value={tempApiKey}
              placeholder="Enter your API key"
            />
            <small>
              {#if useConfig.config()?.APIKey}
                Your API key is set.
              {:else}
                Your API key is not set. It will be stored locally.
              {/if}
            </small>
          </div>

          <div class="setting-group">
            <label for="model">AI Model</label>
            <input
              id="model"
              type="text"
              bind:value={tempModel}
              placeholder="e.g., claude-sonnet-4"
            />
          </div>

          <div class="setting-group">
            <label>Theme</label>
            <div class="theme-buttons">
              <button
                type="button"
                class="theme-btn {useTheme.theme() === 'dark' ? 'active' : ''}"
                onclick={() => useTheme.setTheme("dark")}
              >
                Dark
              </button>
              <button
                type="button"
                class="theme-btn {useTheme.theme() === 'light' ? 'active' : ''}"
                onclick={() => useTheme.setTheme("light")}
              >
                Light
              </button>
            </div>
          </div>

          <div class="setting-group">
            <label class="checkbox">
              <input type="checkbox" bind:checked={tempAutoApprove} />
              Automatically approve bash tool calls
            </label>
          </div>

          <div class="settings-actions">
            <button type="submit" class="primary">Save</button>
          </div>
        </form>
      {/if}
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.45);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
  }

  .overlay-panel {
    min-width: 520px;
    max-width: 720px;
    max-height: 80vh;
    overflow: auto;
  }

  .panel-title {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .panel-title h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }

  .close-btn {
    border-radius: 50%;
    width: 32px;
    height: 32px;
    border: none;
    font-size: 16px;
  }

  .settings-content {
    padding: 16px 0;
  }

  .setting-group {
    margin-bottom: 16px;
  }

  .setting-group label {
    display: block;
    font-weight: 600;
    font-size: 14px;
    margin-bottom: 6px;
  }

  .setting-group input[type="text"],
  .setting-group input[type="password"] {
    width: 100%;
    padding: 12px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .setting-group small {
    color: var(--text-dim);
    font-size: 12px;
    margin-top: 4px;
    display: block;
  }

  .checkbox {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
  }

  .checkbox input[type="checkbox"] {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }

  .theme-buttons {
    display: flex;
    gap: 8px;
  }

  .theme-btn {
    padding: 8px 16px;
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 4px;
    cursor: pointer;
  }

  .theme-btn:hover {
    background: var(--bg-elevated);
  }

  .theme-btn.active {
    background: var(--accent);
    border-color: var(--accent);
    color: var(--accent-text);
  }

  .settings-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    margin-top: 24px;
  }

  .saving-indicator {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 24px 0;
  }

  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>