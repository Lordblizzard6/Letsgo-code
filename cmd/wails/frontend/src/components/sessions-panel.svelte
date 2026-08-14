<script lang="ts">
  import { useSession, useChat } from "../lib/store.svelte";
  import { SessionsService } from "../bindings";

  let tempName = $state("");

  async function loadSessions() {
    const list = await SessionsService.List();
    useSession.setSessions(list);
  }

  async function handleCreate() {
    if (!tempName.trim()) return;
    await SessionsService.Create(tempName.trim(), "");
    tempName = "";
    await loadSessions();
  }

  async function handleOpen(id: string) {
    const messages = await SessionsService.Open(id);
    useSession.setActive(id);
    useChat.setMessages(
      messages.map((m) => ({
        id: String(m.id),
        role: m.role === "user" ? ("user" as const) : ("assistant" as const),
        content: typeof m.content === "string" ? m.content : JSON.stringify(m.content ?? ""),
      })),
    );
    window.location.hash = "";
  }

  $effect(() => {
    loadSessions();
  });
</script>

<aside class="pane">
  <div class="panel-header">
    <h2>Sessions</h2>
  </div>

  <div class="session-form">
    <input
      type="text"
      bind:value={tempName}
      placeholder="New session name"
      onkeydown={(e) => {
        if (e.key === "Enter") handleCreate();
      }}
    />
    <button class="primary" onclick={handleCreate} disabled={!tempName.trim()}>
      Create
    </button>
  </div>

  {#if useSession.sessions().length === 0}
    <div class="empty-state">
      <p>No sessions yet. Create one to start chatting.</p>
    </div>
  {:else}
    <ul class="session-list">
      {#each useSession.sessions() as s (s.id)}
        <li
          class={useSession.activeId() === s.id ? "active" : ""}
          onclick={() => handleOpen(s.id)}
        >
          <strong>{s.name}</strong>
          {#if s.project_path}
            <small>({s.project_path})</small>
          {/if}
          <span class="status" class:busy={s.is_active}></span>
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

  .empty-state {
    text-align: center;
    color: var(--text-dim);
    padding: 40px 16px;
  }

  .session-form {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }

  .session-form input {
    flex: 1;
    padding: 8px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .session-list {
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .session-list li {
    padding: 8px;
    border-radius: 6px;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .session-list li:hover {
    background: var(--bg-input);
  }

  .session-list li.active {
    background: var(--bg-input);
    color: var(--accent);
  }

  .session-list li small {
    color: var(--text-dim);
  }

  .status {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--text-dim);
    margin-left: 8px;
  }

  .status.busy {
    background: var(--ok);
  }
</style>