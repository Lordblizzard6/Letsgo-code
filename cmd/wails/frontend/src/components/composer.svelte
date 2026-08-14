<script lang="ts">
  import { useChat } from "../lib/store.svelte";
  import { ChatService } from "../bindings";

  let message = $state("");

  async function handleSend() {
    if (!message.trim()) return;
    const text = message.trim();
    message = "";
    useChat.setStreaming(true);
    try {
      await ChatService.Start();
      await ChatService.Send(text);
    } catch (err) {
      useChat.setStreaming(false);
      useChat.setError(err instanceof Error ? err.message : String(err));
    }
  }

  async function handleCancel() {
    useChat.setStreaming(false);
    try {
      await ChatService.Cancel();
    } catch {
      useChat.setStreaming(false);
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }
</script>

<div class="composer">
  <textarea
    bind:value={message}
    onkeydown={handleKeyDown}
    placeholder="Type your message..."
    disabled={useChat.isStreaming()}
  ></textarea>

  <div class="actions">
    <button class="primary" onclick={handleSend} disabled={!message.trim() || useChat.isStreaming()}>
      Send
    </button>
    {#if useChat.isStreaming()}
      <button onclick={handleCancel}>Cancel</button>
    {/if}
  </div>
</div>

<style>
  .composer {
    display: flex;
    gap: 8px;
    padding: 12px;
    border-top: 1px solid var(--border);
    background: var(--bg-panel);
  }

  textarea {
    flex: 1;
    resize: none;
    background: var(--bg-input);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 10px;
    min-height: 44px;
    max-height: 160px;
    font-family: var(--font);
    font-size: 14px;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .actions button {
    align-self: flex-end;
  }

  button:disabled {
    opacity: 0.5;
    cursor: default;
  }
</style>