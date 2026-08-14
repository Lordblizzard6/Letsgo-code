<script lang="ts">
  import { usePlugins } from "../lib/store.svelte";
  import { PluginsService } from "../bindings";

  async function loadPlugins() {
    const rows = await PluginsService.List();
    usePlugins.setPlugins(rows);
  }

  async function togglePlugin(name: string, enabled: boolean) {
    if (enabled) {
      await PluginsService.Disable(name);
    } else {
      await PluginsService.Enable(name);
    }
    await loadPlugins();
  }

  $effect(() => {
    loadPlugins();
  });
</script>

<aside class="pane">
  <div class="panel-header">
    <h2>Plugins</h2>
    <span class="count">
      {usePlugins.plugins().filter((p) => p.enabled).length}/{usePlugins.plugins().length} enabled
    </span>
  </div>

  {#if usePlugins.plugins().length === 0}
    <div class="empty-state">
      <p>No plugins installed.</p>
    </div>
  {:else}
    <ul class="plugin-list">
      {#each usePlugins.plugins() as plugin (plugin.name)}
        <li class="plugin-item">
          <div class="plugin-info">
            <strong>{plugin.name}</strong>
            {#if plugin.version}
              <span class="version">v{plugin.version}</span>
            {/if}
            {#if plugin.description}
              <span class="description">{plugin.description}</span>
            {/if}
          </div>
          <button class="toggle-btn" onclick={() => togglePlugin(plugin.name, plugin.enabled)}>
            {plugin.enabled ? "Disable" : "Enable"}
          </button>
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

  .plugin-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .plugin-item {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 12px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .plugin-info {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }

  .plugin-info strong {
    font-weight: 600;
    font-size: 14px;
  }

  .version {
    color: var(--text-dim);
    font-size: 12px;
  }

  .description {
    color: var(--text-dim);
    font-size: 12px;
  }
</style>