<script lang="ts">
  import { useConfig } from "../lib/store.svelte";
  import { SettingsService } from "../bindings";

  const { ondone }: { ondone: () => void } = $props();

  const PROVIDERS = [
    { id: "AnthropicAPIKey", label: "Anthropic", model: "claude-sonnet-4-5" },
    { id: "OpenAIAPIKey", label: "OpenAI", model: "gpt-4o" },
    { id: "GroqAPIKey", label: "Groq", model: "llama-3.3-70b-versatile" },
    { id: "OpenRouterAPIKey", label: "OpenRouter", model: "anthropic/claude-sonnet-4" },
  ];

  let provider = $state("AnthropicAPIKey");
  let apiKey = $state("");
  let model = $state(PROVIDERS[0].model);
  let saving = $state(false);
  let error = $state<string | null>(null);

  function onProviderChange() {
    model = PROVIDERS.find((p) => p.id === provider)?.model ?? "";
  }

  async function handleSave() {
    error = null;
    if (!apiKey.trim()) {
      error = "Please enter an API key.";
      return;
    }
    saving = true;
    const partial: Record<string, unknown> = {
      [provider]: apiKey.trim(),
      APIKey: apiKey.trim(),
    };
    if (model.trim()) partial.Model = model.trim();
    try {
      const cfg = await SettingsService.SaveConfig(partial);
      useConfig.setConfig(cfg);
      ondone();
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      saving = false;
    }
  }
</script>

<div class="onboarding">
  <div class="onboarding-card">
    <h1>Welcome to LetsGO</h1>
    <p class="lead">
      Configure your AI provider to get started. Your key is stored locally in
      your configuration file.
    </p>

    <form onsubmit={(e) => {
      e.preventDefault();
      handleSave();
    }}>
      <div class="field">
        <label for="provider">Provider</label>
        <select id="provider" bind:value={provider} onchange={onProviderChange}>
          {#each PROVIDERS as p (p.id)}
            <option value={p.id}>{p.label}</option>
          {/each}
        </select>
      </div>

      <div class="field">
        <label for="apikey">API Key</label>
        <input
          id="apikey"
          type="password"
          bind:value={apiKey}
          placeholder="sk-..."
          autocomplete="off"
        />
      </div>

      <div class="field">
        <label for="model">Model</label>
        <input id="model" type="text" bind:value={model} placeholder="Model ID" />
      </div>

      {#if error}
        <div class="error-banner">{error}</div>
      {/if}

      <button type="submit" class="primary" disabled={saving}>
        {saving ? "Saving..." : "Save & start chatting"}
      </button>
    </form>
  </div>
</div>

<style>
  .onboarding {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg);
  }

  .onboarding-card {
    width: 420px;
    max-width: 92vw;
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 32px;
  }

  .onboarding-card h1 {
    margin: 0 0 8px;
    font-size: 24px;
    font-weight: 600;
  }

  .lead {
    color: var(--text-dim);
    font-size: 14px;
    line-height: 1.5;
    margin: 0 0 24px;
  }

  .field {
    margin-bottom: 16px;
  }

  .field label {
    display: block;
    font-weight: 600;
    font-size: 14px;
    margin-bottom: 6px;
  }

  .field input,
  .field select {
    width: 100%;
    padding: 10px 12px;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 6px;
  }

  .error-banner {
    border: 1px solid var(--err);
    background: color-mix(in srgb, var(--err) 12%, transparent);
    border-radius: 6px;
    padding: 10px 12px;
    margin-bottom: 16px;
    font-size: 13px;
  }

  button[type="submit"] {
    width: 100%;
    padding: 12px;
  }
</style>