<script lang="ts">
  import { useUsage, useAccount } from "../lib/store.svelte";
  import { UsageService } from "../bindings";

  async function loadUsage() {
    const today = await UsageService.GetUsageToday();
    useUsage.setToday(today);
    useAccount.setToday(today);
  }

  const today = $derived(useUsage.today());

  $effect(() => {
    loadUsage();
  });
</script>

<div class="overlay">
  <div class="overlay-panel">
    <div class="panel-title">
      <h2>Usage</h2>
      <button class="close-btn" onclick={() => (window.location.hash = "")}>
        ✕
      </button>
    </div>

    {#if today}
      <div class="kpi-header">
        <span>Today's Usage</span>
        {#if today.cost_usd > today.budget_cap && today.budget_cap > 0}
          <span class="over-budget">Over Budget</span>
        {/if}
      </div>

      <div class="kpi-row">
        <div class="kpi">
          <div class="value">${today.cost_usd.toFixed(2)}</div>
          <div class="label">Cost</div>
        </div>
        <div class="kpi">
          <div class="value">{today.requests}</div>
          <div class="label">Requests</div>
        </div>
        <div class="kpi">
          <div class="value">{today.tokens}</div>
          <div class="label">Tokens</div>
        </div>
      </div>

      {#if today.budget_cap > 0}
        <div class="budget-bar">
          <div
            class="budget-fill"
            style:width={Math.min(100, (today.cost_usd / today.budget_cap) * 100) + "%"}
          ></div>
        </div>
      {/if}
    {:else}
      <div class="empty-state">
        <p>Not available yet.</p>
      </div>
    {/if}
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

  .kpi-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--border);
  }

  .over-budget {
    color: var(--err);
    font-weight: 600;
  }

  .kpi-row {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
    margin-bottom: 24px;
  }

  .kpi {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 16px;
    text-align: center;
  }

  .kpi .value {
    font-size: 32px;
    font-weight: 600;
    color: var(--accent);
  }

  .kpi .label {
    color: var(--text-dim);
    font-size: 14px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
  }

  .budget-bar {
    height: 10px;
    border-radius: 5px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    overflow: hidden;
  }

  .budget-fill {
    height: 100%;
    background: var(--accent);
  }

  .empty-state {
    text-align: center;
    color: var(--text-dim);
    padding: 40px 16px;
  }
</style>