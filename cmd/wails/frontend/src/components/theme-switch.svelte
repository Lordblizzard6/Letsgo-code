<script lang="ts">
  import { useTheme } from "../lib/store.svelte";
  import { ThemeService } from "../bindings";

  async function toggleTheme() {
    const next = useTheme.theme() === "dark" ? "light" : "dark";
    await ThemeService.Set(next);
    useTheme.setTheme(next);
  }
</script>

<button class="theme-switcher" onclick={toggleTheme} aria-label="Toggle theme">
  {#if useTheme.theme() === "dark"}
    <span class="icon">☀</span>
    <span class="label">Light</span>
  {:else}
    <span class="icon">🌙</span>
    <span class="label">Dark</span>
  {/if}
</button>

<style>
  .theme-switcher {
    position: fixed;
    top: 12px;
    right: 64px;
    z-index: 20;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 20px;
    cursor: pointer;
    font-size: 13px;
  }

  .theme-switcher:hover {
    border-color: var(--accent);
  }

  .icon {
    font-size: 14px;
  }

  .label {
    color: var(--text);
  }
</style>