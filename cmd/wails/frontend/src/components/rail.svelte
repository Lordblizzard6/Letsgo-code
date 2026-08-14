<script lang="ts">
  export let selected = false;
</script>

<nav class="rail">
  <button
    class="rail-item {selected ? 'active' : ''}"
    title="Chat"
    onclick={() => {
      window.location.hash = "#";
    }}
  >
    💬
  </button>

  <button class="rail-item" title="Sessions" onclick={() => window.location.hash = "#sessions"}>
    📁
  </button>

  <button class="rail-item" title="Git" onclick={() => window.location.hash = "#git"}>
    🌿
  </button>

  <button class="rail-item" title="Tasks" onclick={() => window.location.hash = "#tasks"}>
    ✅
  </button>

  <button class="rail-item" title="MCP Servers" onclick={() => window.location.hash = "#mcp"}>
    🔌
  </button>

  <button class="rail-item" title="Plugins" onclick={() => window.location.hash = "#plugins"}>
    🔌
  </button>

  <button class="rail-item" title="Settings" onclick={() => window.location.hash = "#settings"}>
    ⚙️
  </button>

  <button class="rail-item" title="Usage" onclick={() => window.location.hash = "#usage"}>
    📊
  </button>

  <button class="rail-item" title="Help" onclick={() => window.location.hash = "#help"}>
    ❓
  </button>
</nav>

<style>
  .rail {
    background: var(--bg-panel);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    width: 64px;
    padding: 8px 4px;
    gap: 2px;
    flex-shrink: 0;
  }

  .rail-item {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 40px;
    border: none;
    background: transparent;
    color: var(--text-dim);
    border-radius: 6px;
  }

  .rail-item:hover {
    color: var(--text);
    background: var(--bg-input);
  }

  .rail-item.active {
    color: var(--accent);
    background: var(--bg-input);
  }

  .rail-item.active::before {
    content: "";
    position: absolute;
    left: -4px;
    top: 10px;
    bottom: 10px;
    width: 3px;
    border-radius: 2px;
    background: var(--accent);
  }
</style>