<script lang="ts">
  import { useMCP } from "../lib/store.svelte";
  import { MCPService } from "../bindings";

  async function loadServers() {
    const servers = await MCPService.List();
    useMCP.setServers(servers);
  }

  async function toggleServer(name: string, running: boolean) {
    if (running) {
      await MCPService.Stop(name);
    } else {
      await MCPService.Start(name);
    }
    await loadServers();
  }

  $effect(() => {
    loadServers();
  });
</script>

<aside class="pane">
  <div class="panel-header">
    <h2>MCP Servers</h2>
    <span class="count">
      {useMCP.servers().filter((s) => s.running).length}/{useMCP.servers().length} running
    </span>
  </div>

  {#if useMCP.servers().length === 0}
    <div class="empty-state">
      <p>No MCP servers configured.</p>
    </div>
  {:else}
    <ul class="server-list">
      {#each useMCP.servers() as server (server.name)}
        <li class="server-item">
          <div class="server-info">
            <strong>{server.name}</strong>
            {#if server.command}
              <code class="command">{server.command}</code>
            {/if}
          </div>
          <div class="server-actions">
            <span class="status" class:started={server.running} class:stopped={!server.running}>
              {server.running ? "running" : "stopped"}
            </span>
            <button class="toggle-btn" onclick={() => toggleServer(server.name, server.running)}>
              {server.running ? "Stop" : "Start"}
            </button>
          </div>
        </li>
      {/each}
    </ul>
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

  .count {
    font-size: 12px;
    color: var(--text-dim);
  }

  .empty-state {
    text-align: center;
    color: var(--text-dim);
    padding: 40px 16px;
  }

  .server-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .server-item {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .server-info {
    display: flex;
    gap: 8px;
    align-items: center;
    min-width: 0;
  }

  .server-info strong {
    font-weight: 600;
    font-size: 14px;
  }

  .command {
    color: var(--text-dim);
    font-family: var(--mono);
    font-size: 12px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .server-actions {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status {
    font-size: 11px;
    padding: 2px 6px;
    border-radius: 4px;
    background: var(--bg-input);
  }

  .status.started {
    background: var(--ok);
    color: var(--bg);
  }

  .status.stopped {
    background: var(--err);
    color: var(--bg);
  }

  .toggle-btn:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>