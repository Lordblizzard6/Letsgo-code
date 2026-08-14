<script lang="ts">
  export type PaletteItem = {
    id: string;
    label: string;
    icon?: string;
    hint?: string;
  };

  const {
    items,
    onClose,
    onSelect,
  }: {
    items: PaletteItem[];
    onClose: () => void;
    onSelect: (item: PaletteItem) => void;
  } = $props();

  let query = $state("");
  let inputEl = $state<HTMLInputElement | null>(null);

  const filtered = $derived(
    query
      ? items.filter((i) =>
          i.label.toLowerCase().includes(query.toLowerCase()),
        )
      : items,
  );

  $effect(() => {
    inputEl?.focus();
  });
</script>

<div class="palette-overlay" onclick={onClose}>
  <div
    class="palette"
    role="dialog"
    aria-label="Command palette"
    onclick={(e) => e.stopPropagation()}
  >
    <input
      bind:this={inputEl}
      bind:value={query}
      placeholder="Type a command or destination..."
      onkeydown={(e) => {
        if (e.key === "Enter" && filtered[0]) onSelect(filtered[0]);
        if (e.key === "Escape") onClose();
      }}
    />
    {#if filtered.length === 0}
      <div class="palette-empty">No matches</div>
    {:else}
      <ul class="palette-list">
        {#each filtered as item (item.id)}
          <li onmousedown={() => onSelect(item)}>
            <span class="palette-icon">{item.icon ?? "›"}</span>
            <span class="palette-label">{item.label}</span>
            {#if item.hint}
              <span class="palette-hint">{item.hint}</span>
            {/if}
          </li>
        {/each}
      </ul>
    {/if}
  </div>
</div>

<style>
  .palette-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.45);
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding-top: 12vh;
    z-index: 60;
  }

  .palette {
    width: 520px;
    max-width: 90vw;
    max-height: 60vh;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 8px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
  }

  .palette input {
    border: none;
    border-bottom: 1px solid var(--border);
    background: var(--bg-input);
    padding: 14px 16px;
    font-size: 15px;
    border-radius: 8px 8px 0 0;
  }

  .palette-list {
    list-style: none;
    margin: 0;
    padding: 6px;
    overflow: auto;
  }

  .palette-list li {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    border-radius: 6px;
    cursor: pointer;
  }

  .palette-list li:hover,
  .palette-list li:focus-within {
    background: var(--bg-input);
  }

  .palette-icon {
    color: var(--text-dim);
    width: 18px;
    text-align: center;
  }

  .palette-label {
    flex: 1;
  }

  .palette-hint {
    color: var(--text-dim);
    font-size: 12px;
  }

  .palette-empty {
    padding: 16px;
    color: var(--text-dim);
    text-align: center;
  }
</style>