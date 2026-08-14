<script lang="ts">
  import { useTheme, useConfig, useAccount, useChat } from "./lib/store.svelte";
  import { AccountService, SettingsService, ThemeService } from "./bindings";
  import Chat from "./components/chat.svelte";
  import SessionsPanel from "./components/sessions-panel.svelte";
  import GitPanel from "./components/git-panel.svelte";
  import TasksPanel from "./components/tasks-panel.svelte";
  import MCPPanel from "./components/mcp-panel.svelte";
  import PluginsPanel from "./components/plugins-panel.svelte";
  import SettingsOverlay from "./components/settings-overlay.svelte";
  import UsageOverlay from "./components/usage-overlay.svelte";
  import HelpOverlay from "./components/help-overlay.svelte";
  import ThemeSwitch from "./components/theme-switch.svelte";
  import AccountFlyout from "./components/account-flyout.svelte";
  import Palette, { type PaletteItem } from "./components/palette.svelte";
  import ProviderSetup from "./components/ProviderSetup.svelte";

  type Tab =
    | "chat"
    | "sessions"
    | "git"
    | "tasks"
    | "mcp"
    | "plugins"
    | "settings"
    | "usage"
    | "help";

  const TABS: Tab[] = [
    "chat",
    "sessions",
    "git",
    "tasks",
    "mcp",
    "plugins",
    "settings",
    "usage",
    "help",
  ];
  const WORK: Tab[] = ["chat", "sessions", "git", "tasks"];
  const TOOLS: Tab[] = ["mcp", "plugins"];
  const SYSTEM: Tab[] = ["settings", "usage", "help"];
  const ICONS: Record<Tab, string> = {
    chat: "💬",
    sessions: "📁",
    git: "🌿",
    tasks: "✅",
    mcp: "🔌",
    plugins: "🧩",
    settings: "⚙️",
    usage: "📊",
    help: "❓",
  };

  function navigate(tab: Tab) {
    window.location.hash = tab === "chat" ? "" : `#${tab}`;
  }

  const paletteItems: PaletteItem[] = [
    ...TABS.map((tab) => ({
      id: tab,
      label: tab.charAt(0).toUpperCase() + tab.slice(1),
      icon: ICONS[tab],
      hint: tab === "chat" ? "Alt+1" : `Alt+${TABS.indexOf(tab) + 1}`,
    })),
    {
      id: "__theme__",
      label: "Toggle theme",
      icon: "🎨",
      hint: "Dark / Light",
    },
  ];

  let currentTab = $state<Tab>("chat");
  let accountOpen = $state(false);
  let paletteOpen = $state(false);
  let configLoaded = $state(false);
  let hasKey = $state(false);

  const gated = $derived(configLoaded && !hasKey);

  $effect(() => {
    const read = () => {
      const hash = window.location.hash.replace("#", "");
      currentTab = (TABS as string[]).includes(hash) ? (hash as Tab) : "chat";
    };
    read();
    window.addEventListener("hashchange", read);
    return () => window.removeEventListener("hashchange", read);
  });

  $effect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.key.toLowerCase() === "k") {
        e.preventDefault();
        paletteOpen = true;
        return;
      }
      if (e.altKey && e.key >= "1" && e.key <= "9") {
        e.preventDefault();
        const idx = Number(e.key) - 1;
        if (idx >= 0 && idx < TABS.length) {
          paletteOpen = false;
          accountOpen = false;
          navigate(TABS[idx]);
        }
        return;
      }
      if (e.key === "Escape") {
        if (paletteOpen) {
          paletteOpen = false;
          refocusComposer();
        } else if (accountOpen) {
          accountOpen = false;
        }
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  });

  function refocusComposer() {
    document.querySelector<HTMLTextAreaElement>(".composer textarea")?.focus();
  }

  function handlePaletteSelect(item: PaletteItem) {
    paletteOpen = false;
    if (item.id === "__theme__") {
      const next = useTheme.theme() === "dark" ? "light" : "dark";
      useTheme.setTheme(next);
      ThemeService.Set(next);
      return;
    }
    if ((TABS as string[]).includes(item.id)) {
      navigate(item.id as Tab);
    }
  }

  $effect(() => {
    ThemeService.Get().then((v) => {
      if (v === "dark" || v === "light") useTheme.setTheme(v);
    });
    SettingsService.GetConfig().then((c) => {
      useConfig.setConfig(c);
      configLoaded = true;
      hasKey = useConfig.hasCredentials();
    });
    AccountService.GetStatus().then((s) => {
      useAccount.setSnapshot(s.online, s.busy, s.provider, s.model);
    });
  });
</script>

{#if gated}
  <ProviderSetup ondone={() => (hasKey = useConfig.hasCredentials())} />
{:else}
<div class="app">
  <nav class="rail" aria-label="Main">
    <div class="rail-section">Work</div>
    {#each WORK as tab}
      <button
        class="rail-item {currentTab === tab ? 'active' : ''}"
        class:dot-busy={tab === "chat" && useChat.isStreaming()}
        onclick={() => navigate(tab)}
        title={tab}
      >
        {ICONS[tab]}
      </button>
    {/each}

    <div class="rail-section">Tools</div>
    {#each TOOLS as tab}
      <button
        class="rail-item {currentTab === tab ? 'active' : ''}"
        onclick={() => navigate(tab)}
        title={tab}
      >
        {ICONS[tab]}
      </button>
    {/each}

    <div class="rail-section">System</div>
    {#each SYSTEM as tab}
      <button
        class="rail-item {currentTab === tab ? 'active' : ''}"
        onclick={() => navigate(tab)}
        title={tab}
      >
        {ICONS[tab]}
      </button>
    {/each}
  </nav>

  <main class="main">
    {#if currentTab === "chat"}
      <Chat />
    {:else if currentTab === "sessions"}
      <SessionsPanel />
    {:else if currentTab === "git"}
      <GitPanel />
    {:else if currentTab === "tasks"}
      <TasksPanel />
    {:else if currentTab === "mcp"}
      <MCPPanel />
    {:else if currentTab === "plugins"}
      <PluginsPanel />
    {:else if currentTab === "settings"}
      <SettingsOverlay />
    {:else if currentTab === "usage"}
      <UsageOverlay />
    {:else}
      <HelpOverlay />
    {/if}
  </main>

  <ThemeSwitch />

  {#if paletteOpen}
    <Palette
      items={paletteItems}
      onClose={() => (paletteOpen = false)}
      onSelect={handlePaletteSelect}
    />
  {/if}

  <div class="account-area">
    <button
      class="avatar"
      class:offline={useAccount.status() === "offline"}
      onclick={() => (accountOpen = !accountOpen)}
      aria-label="Account"
    >
      {useAccount.status() === "offline" ? "○" : "●"}
    </button>
    {#if accountOpen}
      <AccountFlyout onclose={() => (accountOpen = false)} />
    {/if}
  </div>
</div>
{/if}

<style>
  .app {
    display: grid;
    grid-template-columns: auto 1fr;
    height: 100vh;
    background: var(--bg);
  }

  .main {
    overflow: hidden;
  }

  .account-area {
    position: fixed;
    top: 12px;
    right: 16px;
    z-index: 20;
  }

  .avatar {
    width: 36px;
    height: 36px;
    border-radius: 50%;
    background: var(--bg-input);
    border: 1px solid var(--border);
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--ok);
    font-size: 16px;
  }

  .avatar.offline {
    color: var(--text-dim);
  }
</style>