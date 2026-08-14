<script lang="ts">
  import { useAccount } from "../lib/store.svelte";
  import { AccountService } from "../bindings";

  const { onclose = undefined }: { onclose?: () => void } = $props();

  async function handleSignOut() {
    await AccountService.SignOut();
    useAccount.setStatus("offline");
    useAccount.setSnapshot(false, false, "", "");
    onclose?.();
  }

  const today = $derived(useAccount.today());
</script>

<div class="account-flyout" role="dialog">
  <div class="flyout-header">
    <span class="status-dot" class:online={useAccount.status() === "online"} class:busy={useAccount.status() === "busy"}></span>
    <span class="status-text"><strong>{useAccount.status()}</strong></span>
  </div>

  <div class="flyout-content">
    <div class="info-row">
      <span class="label">Provider</span>
      <span class="value">{useAccount.provider() || "—"}</span>
    </div>

    <div class="info-row">
      <span class="label">Model</span>
      <span class="value">{useAccount.model() || "—"}</span>
    </div>

    {#if today}
      <div class="info-row">
        <span class="label">Today's Usage</span>
        <span class="value">${today.cost_usd.toFixed(2)}</span>
      </div>
    {/if}

    <button class="sign-out-btn" onclick={handleSignOut}>
      Sign Out
    </button>
  </div>
</div>

<style>
  .account-flyout {
    position: absolute;
    top: 44px;
    right: 0;
    min-width: 220px;
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 6px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    z-index: 40;
  }

  .flyout-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px;
    border-bottom: 1px solid var(--border);
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--text-dim);
  }

  .status-dot.online {
    background: var(--ok);
  }

  .status-dot.busy {
    background: var(--warn);
  }

  .status-text {
    font-size: 13px;
    text-transform: capitalize;
  }

  .flyout-content {
    padding: 12px;
  }

  .info-row {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 12px;
  }

  .info-row:last-child {
    margin-bottom: 0;
  }

  .label {
    font-size: 12px;
    color: var(--text-dim);
  }

  .value {
    font-size: 13px;
    font-weight: 500;
  }

  .sign-out-btn {
    width: 100%;
    margin-top: 8px;
    padding: 6px 12px;
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
  }

  .sign-out-btn:hover {
    background: var(--bg-input);
    border-color: var(--err);
  }
</style>