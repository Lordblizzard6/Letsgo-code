<script lang="ts">
  let helpVisible = $state(true);

  function closeHelp() {
    helpVisible = false;
    window.location.hash = "";
  }
</script>

{#if helpVisible}
  <div class="overlay">
    <div class="overlay-panel">
      <div class="panel-title">
        <h2>Help</h2>
        <button class="close-btn" onclick={closeHelp}>
          ✕
        </button>
      </div>

      <div class="help-content">
        <h3>Keyboard Shortcuts</h3>
        <ul class="shortcut-list">
          <li>
            <kbd>Alt + 1</kbd> to Alt + 8
            <span>Chat</span>
          </li>
          <li>
            <kbd>Alt + 9</kbd>
            <span>Help</span>
          </li>
          <li>
            <kbd>Esc</kbd>
            <span>Close help</span>
          </li>
          <li>
            <kbd>Ctrl + K</kbd>
            <span>Show palette</span>
          </li>
        </ul>

        <h3>Tools</h3>
        <ul class="tool-list">
          <li>
            <strong>Tool Approval</strong>
            <p>
              Tools are displayed in the status bar. If auto-approve is disabled,
              you'll see pending tool calls that you need to approve or reject.
            </p>
          </li>
          <li>
            <strong>Tool Retry</strong>
            <p>
              If a tool call fails, you can retry it using the tool retry button.
            </p>
          </li>
        </ul>

        <h3>Sessions</h3>
        <ul class="session-list">
          <li>
            <strong>Creating a Session</strong>
            <p>
              Click "New" to create a new session. You can then enter a name for the session
              and optionally specify the project path.
            </p>
          </li>
          <li>
            <strong>Multiple Sessions</strong>
            <p>
              You can create multiple sessions and switch between them using the sessions panel.
            </p>
          </li>
        </ul>

        <h3>Git</h3>
        <ul class="git-list">
          <li>
            <strong>Git Operations</strong>
            <p>
              The Git panel allows you to commit, reset, and create WIP commits.
              You can also view the markup of your changes.
            </p>
          </li>
        </ul>

        <h3>Settings</h3>
        <ul class="settings-list">
          <li>
            <strong>API Key</strong>
            <p>
              Your API key is stored locally in your configuration file. It is never sent
              to any other server except the AI service provider.
            </p>
          </li>
          <li>
            <strong>Theme</strong>
            <p>
              You can switch between dark and light themes. The theme is saved in your configuration.
            </p>
          </li>
          <li>
            <strong>Auto Approve</strong>
            <p>
              When auto-approve is enabled, all tool calls will be automatically approved.
              When disabled, you'll need to approve each tool call manually.
            </p>
          </li>
        </ul>
      </div>
    </div>
  </div>
{/if}

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

  .help-content h3 {
    margin: 24px 0 12px;
    font-size: 16px;
    font-weight: 600;
  }

  .shortcut-list,
  .tool-list,
  .session-list,
  .git-list,
  .settings-list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .shortcut-list li,
  .tool-list li,
  .session-list li,
  .git-list li,
  .settings-list li {
    margin-bottom: 12px;
  }

  .shortcut-list span,
  .tool-list p,
  .session-list p,
  .git-list p,
  .settings-list p {
    color: var(--text-dim);
    font-size: 13px;
  }

  .shortcut-list kbd {
    display: inline-block;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 4px 8px;
    font-family: var(--mono);
    font-size: 12px;
    margin-right: 8px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  }
</style>