<script lang="ts">
  import { useTools } from "../lib/store.svelte";
</script>

<div class="tool-list">
  {#each useTools.toolActivity() as activity (activity.callId)}
    <div class="tool-row" class:ok={activity.success} class:err={!activity.success}>
      <span>{activity.name || "tool"}</span>
      {#if activity.error}
        <span class="error">: {activity.error}</span>
      {/if}
    </div>
  {/each}
</div>

<style>
  .tool-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-top: 12px;
  }

  .tool-row {
    font-size: 12px;
    color: var(--text-dim);
    border-left: 2px solid var(--border);
    padding: 2px 8px;
  }

  .tool-row.ok {
    color: var(--ok);
    border-left-color: var(--ok);
  }

  .tool-row.err {
    color: var(--err);
    border-left-color: var(--err);
  }

  .error {
    color: var(--err);
  }
</style>