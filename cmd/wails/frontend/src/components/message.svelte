<script lang="ts">
  import type { ChatMessage } from "../lib/store.svelte";
  import { renderMarkdown } from "../lib/markdown";

  export let message: ChatMessage;
</script>

<article class="msg">
  <div class="role">{message.role === "user" ? "You" : "Assistant"}</div>

  <div class="content">
    {@html renderMarkdown(message.content)}
  </div>
</article>

<style>
  .msg {
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 16px;
    background: var(--bg-panel);
    margin-bottom: 12px;
  }

  .role {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-dim);
    margin-bottom: 6px;
  }

  .content {
    display: block;
  }

  .content :global(pre) {
    background: var(--bg-elevated);
    border: 1px solid var(--border);
    border-radius: 6px;
    padding: 8px;
    overflow-x: auto;
  }

  .content :global(code) {
    font-family: "Cascadia Mono", Consolas, monospace;
    font-size: 13px;
  }
</style>