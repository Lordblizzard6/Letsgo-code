<script lang="ts">
  import { useAgents } from "../lib/store.svelte";
  import { TasksService } from "../bindings";

  function isActive(status: string) {
    return status === "active" || status === "running";
  }

  async function loadAgents() {
    const rows = await TasksService.List();
    useAgents.setAgents(rows);
  }

  $effect(() => {
    loadAgents();
  });
</script>

<aside class="pane">
  <div class="panel-header">
    <h2>Tasks</h2>
    <span class="count">
      {useAgents.agents().filter((a) => isActive(a.status)).length}/{useAgents.agents().length} active
    </span>
  </div>

  {#if useAgents.agents().length === 0}
    <div class="empty-state">
      <p>No agents configured.</p>
    </div>
  {:else}
    <ul class="agent-list">
      {#each useAgents.agents() as agent (agent.id)}
        <li class="agent-item" class:active={isActive(agent.status)}>
          <div class="agent-header">
            <span class="agent-name">{agent.id}</span>
            <span class="status" class:active={isActive(agent.status)}>{agent.status}</span>
          </div>
          {#if agent.goal}
            <div class="agent-goal">{agent.goal}</div>
          {/if}
          {#if agent.result}
            <div class="agent-result">{agent.result}</div>
          {/if}
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

  .agent-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .agent-item {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
  }

  .agent-item.active {
    border-color: var(--accent);
    background: color-mix(in srgb, var(--accent) 8%, transparent);
  }

  .agent-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .agent-name {
    font-weight: 600;
    font-size: 14px;
  }

  .agent-goal,
  .agent-result {
    color: var(--text-dim);
    font-size: 12px;
    margin-top: 4px;
  }

  .status {
    display: inline-block;
    font-size: 11px;
    padding: 2px 6px;
    border-radius: 4px;
    background: var(--bg-input);
  }

  .status.active {
    background: var(--ok);
    color: var(--bg);
  }
</style>