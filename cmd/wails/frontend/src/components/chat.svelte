<script lang="ts">
  import { useChat } from "../lib/store.svelte";
  import { ChatService } from "../bindings";
  import Message from "./message.svelte";
  import Composer from "./composer.svelte";
  import ToolActivity from "./tool-activity.svelte";
async function handleRetry() {
    useChat.setError(null);
    useChat.setStreaming(true);
    try {
      await ChatService.Start();
    } catch (err) {
      useChat.setStreaming(false);
      useChat.setError(err instanceof Error ? err.message : String(err));
    }
  }
</script>

<main class="pane chat-col">
  {#if useChat.error()}
    <div class="error-banner">
      <strong>Error:</strong> {useChat.error()}
      <p>Check your API key, connection and model, then retry.</p>
      <button onclick={handleRetry}>Retry</button>
    </div>
  {/if}

  {#if useChat.messages().length === 0}
    <div class="empty-state">
      <p>No messages yet. Send a message to start a chat.</p>
    </div>
  {:else}
    {#each useChat.messages() as message (message.id)}
      <Message {message} />
    {/each}

    <div class="stream-indicator">
      {#if useChat.isStreaming()}
        <div class="spinner"></div>
        <span>Streaming...</span>
      {/if}
      {#if useChat.idle()}
        <span>Ready</span>
      {/if}
    </div>
  {/if}

  <ToolActivity />
  <Composer />
</main>

<style>
  .pane.chat-col {
    max-width: 720px;
    margin: 0 auto;
  }

  .error-banner {
    border: 1px solid var(--err);
    background: color-mix(in srgb, var(--err) 12%, transparent);
    border-radius: 6px;
    padding: 12px;
    margin-bottom: 12px;
  }

  .error-banner p {
    margin: 6px 0;
    font-size: 13px;
  }

  .empty-state {
    text-align: center;
    color: var(--text-dim);
    padding: 40px 16px;
  }

  .stream-indicator {
    display: flex;
    align-items: center;
    gap: 8px;
    color: var(--text-dim);
    font-size: 12px;
    padding: 4px 2px 12px;
  }

  .spinner {
    width: 12px;
    height: 12px;
    border: 2px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>