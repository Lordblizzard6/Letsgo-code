<script lang="ts">
  import { GitService } from "../bindings";

  let cwd = $state("");
  let branch = $state("");
  let status = $state("");
  let commitMsg = $state("");

  async function loadGitInfo() {
    const [dir, currentBranch, gitStatus] = await Promise.all([
      GitService.Dir(),
      GitService.CurrentBranch(),
      GitService.Status(),
    ]);
    cwd = dir;
    branch = currentBranch;
    status = gitStatus;
  }

  async function handleCommit() {
    if (!commitMsg.trim()) return;
    await GitService.Commit(commitMsg.trim());
    commitMsg = "";
    await loadGitInfo();
  }

  async function handleUndo() {
    await GitService.Undo();
    await loadGitInfo();
  }

  async function handlePush() {
    await GitService.Push();
    await loadGitInfo();
  }

  $effect(() => {
    loadGitInfo();
  });
</script>

<aside class="pane">
  <div class="panel-header">
    <h2>Git</h2>
    <button class="primary" onclick={loadGitInfo}>Refresh</button>
  </div>

  {#if !cwd}
    <div class="empty-state">
      <p>No git repository loaded.</p>
    </div>
  {:else}
    <div class="git-info">
      <div class="info-row">
        <span class="label">Repository</span>
        <code>{cwd}</code>
      </div>
      <div class="info-row">
        <span class="label">Branch</span>
        <span class="branch">{branch || "(detached)"}</span>
      </div>
    </div>

    <div class="git-actions">
      <input
        type="text"
        bind:value={commitMsg}
        placeholder="Commit message"
        onkeydown={(e) => {
          if (e.key === "Enter") handleCommit();
        }}
      />
      <button class="primary" onclick={handleCommit} disabled={!commitMsg.trim()}>
        Commit
      </button>
      <button onclick={handleUndo}>Undo</button>
      <button onclick={handlePush}>Push</button>
    </div>

    {#if status}
      <div class="git-status">
        <h3>Status</h3>
        <pre>{status}</pre>
      </div>
    {/if}
  {/if}
</aside>

<style>
  .pane {
    overflow: auto;
    padding: 16px;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .panel-header h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .empty-state {
    text-align: center;
    color: var(--text-dim);
    padding: 40px 16px;
  }

  .git-info {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 16px;
  }

  .info-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .git-info .label {
    font-size: 12px;
    color: var(--text-dim);
    font-weight: 600;
  }

  .git-info code {
    color: var(--accent);
    font-family: var(--mono);
    font-size: 13px;
    word-break: break-all;
  }

  .branch {
    color: var(--ok);
  }

  .git-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 16px;
  }

  .git-actions input {
    flex: 1;
    min-width: 180px;
    padding: 6px 10px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .git-status h3 {
    margin: 0 0 8px;
    font-size: 14px;
    font-weight: 600;
  }

  .git-status pre {
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    font-size: 12px;
    overflow-x: auto;
    font-family: var(--mono);
    white-space: pre-wrap;
    margin: 0;
  }
</style>