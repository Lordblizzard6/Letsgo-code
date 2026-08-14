import { Events } from "@wailsio/runtime";
import type { Session } from "../../bindings/github.com/user/go-claude-code/internal/db/models.js";
import type { Config } from "../../bindings/github.com/user/go-claude-code/internal/config/models.js";
import type {
  Agent,
  Plugin,
  Server,
  Today,
} from "../../bindings/github.com/user/go-claude-code/cmd/wails/services/models.js";

export type ChatMessage = {
  id: string;
  role: "user" | "assistant";
  content: string;
};

export type ToolActivity = {
  callId: string;
  name: string;
  success: boolean;
  error?: string;
};

const chat = $state<{
  messages: ChatMessage[];
  isStreaming: boolean;
  error: string | null;
  idle: boolean;
  streamingText: string;
}>({
  messages: [],
  isStreaming: false,
  error: null,
  idle: false,
  streamingText: "",
});

const session = $state<{
  sessions: Session[];
  activeId: string | null;
  isOpen: boolean;
}>({
  sessions: [],
  activeId: null,
  isOpen: false,
});

const themeState = $state<{ theme: "dark" | "light" }>({ theme: "dark" });

const account = $state<{
  status: "offline" | "online" | "busy";
  provider: string;
  model: string;
  today: Today | null;
}>({
  status: "offline",
  provider: "",
  model: "",
  today: null,
});

const mcp = $state<{ servers: Server[] }>({ servers: [] });
const plugins = $state<{ plugins: Plugin[] }>({ plugins: [] });
const agents = $state<{ agents: Agent[] }>({ agents: [] });
const usage = $state<{ today: Today | null }>({ today: null });

const config = $state<{ config: Config | null; isSaving: boolean }>({
  config: null,
  isSaving: false,
});

const tools = $state<{ activity: ToolActivity[]; requesting: boolean }>({
  activity: [],
  requesting: false,
});

export const useChat = {
  messages: () => chat.messages,
  isStreaming: () => chat.isStreaming,
  error: () => chat.error,
  idle: () => chat.idle,
  streamingText: () => chat.streamingText,
  setStreaming: (v: boolean) => {
    chat.isStreaming = v;
    if (!v) chat.streamingText = "";
  },
  setError: (msg: string | null) => {
    chat.error = msg;
  },
  setIdle: (v: boolean) => {
    chat.idle = v;
  },
  setMessages: (msgs: ChatMessage[]) => {
    chat.messages = msgs;
    chat.isStreaming = false;
    chat.streamingText = "";
    chat.error = null;
  },
  appendUser: (text: string) => {
    chat.messages = [
      ...chat.messages,
      { id: `user-${Date.now()}`, role: "user", content: text },
    ];
  },
  startStream: () => {
    chat.isStreaming = true;
    chat.error = null;
    chat.streamingText = "";
    chat.messages = [
      ...chat.messages,
      { id: `assistant-${Date.now()}`, role: "assistant", content: "" },
    ];
  },
  appendDelta: (text: string) => {
    if (!chat.isStreaming) return;
    chat.streamingText += text;
    const last = chat.messages[chat.messages.length - 1];
    if (last && last.role === "assistant") {
      chat.messages[chat.messages.length - 1] = {
        ...last,
        content: last.content + text,
      };
    }
  },
  endStream: (err?: string | null) => {
    chat.isStreaming = false;
    chat.streamingText = "";
    chat.error = err ?? null;
  },
  lastUserText: (): string => {
    for (let i = chat.messages.length - 1; i >= 0; i--) {
      if (chat.messages[i].role === "user") return chat.messages[i].content;
    }
    return "";
  },
  reset: () => {
    chat.messages = [];
    chat.isStreaming = false;
    chat.error = null;
    chat.idle = false;
    chat.streamingText = "";
  },
};

export const useSession = {
  sessions: () => session.sessions,
  activeId: () => session.activeId,
  hasActive: () => session.activeId !== null,
  isOpen: () => session.isOpen,
  setOpen: (v: boolean) => {
    session.isOpen = v;
  },
  setSessions: (list: Session[]) => {
    session.sessions = list;
    const active = list.find((s) => s.is_active);
    session.activeId = active ? active.id : session.activeId;
  },
  setActive: (id: string | null) => {
    session.activeId = id;
  },
};

export const useTheme = {
  theme: () => themeState.theme,
  variant: () => themeState.theme,
  setTheme: (v: "dark" | "light") => {
    themeState.theme = v;
    document.documentElement.dataset.theme = v;
  },
};

export const useAccount = {
  status: () => account.status,
  provider: () => account.provider,
  model: () => account.model,
  today: () => account.today,
  setStatus: (s: "offline" | "online" | "busy") => {
    account.status = s;
  },
  setSnapshot: (online: boolean, busy: boolean, provider: string, model: string) => {
    account.status = !online ? "offline" : busy ? "busy" : "online";
    account.provider = provider;
    account.model = model;
  },
  setToday: (t: Today | null) => {
    account.today = t;
  },
};

export const useMCP = {
  servers: () => mcp.servers,
  setServers: (list: Server[]) => {
    mcp.servers = list;
  },
};

export const usePlugins = {
  plugins: () => plugins.plugins,
  setPlugins: (list: Plugin[]) => {
    plugins.plugins = list;
  },
};

export const useAgents = {
  agents: () => agents.agents,
  setAgents: (list: Agent[]) => {
    agents.agents = list;
  },
};

export const useUsage = {
  today: () => usage.today,
  setToday: (t: Today | null) => {
    usage.today = t;
  },
};

export const useConfig = {
  config: () => config.config,
  isSaving: () => config.isSaving,
  setConfig: (c: Config | null) => {
    config.config = c;
  },
  setSaving: (v: boolean) => {
    config.isSaving = v;
  },
  hasCredentials: () => {
    const c = config.config;
    if (!c) return false;
    return Boolean(
      c.APIKey ||
        c.AnthropicAPIKey ||
        c.OpenAIAPIKey ||
        c.GroqAPIKey ||
        c.OpenRouterAPIKey,
    );
  },
};

export const useTools = {
  toolActivity: () => tools.activity,
  toolRequesting: () => tools.requesting,
  setToolRequesting: (v: boolean) => {
    tools.requesting = v;
  },
  pushActivity: (a: ToolActivity) => {
    tools.activity = [a, ...tools.activity].slice(0, 20);
  },
  upsertActivity: (a: ToolActivity) => {
    const idx = tools.activity.findIndex((t) => t.callId === a.callId);
    if (idx >= 0) {
      tools.activity[idx] = a;
      tools.activity = [...tools.activity];
    } else {
      tools.activity = [a, ...tools.activity].slice(0, 20);
    }
  },
  clearActivity: () => {
    tools.activity = [];
  },
};

export const actions = {
  setSessions: useSession.setSessions,
  setKPI: useUsage.setToday,
  setTheme: useTheme.setTheme,
  setMCP: (_started: number, _total: number, servers: Server[]) => useMCP.setServers(servers),
  setAgents: (_active: number, _total: number, rows: Agent[]) => useAgents.setAgents(rows),
  setPlugins: (_installed: number, _enabled: number, rows: Plugin[]) =>
    usePlugins.setPlugins(rows),
  setSavingConfig: useConfig.setSaving,
  setAccountStatus: useAccount.setStatus,
  setToolRequesting: useTools.setToolRequesting,
};

function wireEngineEvents() {
  Events.On("chat:start", () => {
    useChat.startStream();
  });
  Events.On("chat:user", (ev: any) => {
    if (ev.data?.text) useChat.appendUser(String(ev.data.text));
  });
  Events.On("chat:delta", (ev: any) => {
    if (ev.data?.text) useChat.appendDelta(String(ev.data.text));
  });
  Events.On("chat:end", () => {
    useChat.endStream();
  });
  Events.On("chat:cancelled", () => {
    useChat.endStream();
  });
  Events.On("chat:error", (ev: any) => {
    useChat.endStream(ev.data?.message ?? "Engine error");
  });
  Events.On("chat:cleared", () => {
    useChat.reset();
  });
  Events.On("chat:idle", () => {
    useChat.endStream();
    useChat.setIdle(true);
  });
  Events.On("tool:start", (ev: any) => {
    useTools.upsertActivity({
      callId: ev.data?.call_id ?? "",
      name: ev.data?.name ?? "",
      success: true,
    });
  });
  Events.On("tool:end", (ev: any) => {
    useTools.upsertActivity({
      callId: ev.data?.call_id ?? "",
      name: ev.data?.name ?? "",
      success: !ev.data?.is_error && !ev.data?.rejected && !ev.data?.timed_out,
      error: ev.data?.is_error ? String(ev.data?.content ?? "").slice(0, 200) : undefined,
    });
  });
  Events.On("session:list", (ev: any) => {
    if (Array.isArray(ev.data)) useSession.setSessions(ev.data as Session[]);
  });
  Events.On("config:changed", (ev: any) => {
    const payload = ev.data?.config ?? ev.data;
    if (payload) {
      useConfig.setConfig(payload as Config);
      const variant = (payload as Config).ThemeVariant;
      if (variant === "dark" || variant === "light") useTheme.setTheme(variant);
    }
  });
  Events.On("theme:changed", (ev: any) => {
    if (ev.data?.variant === "dark" || ev.data?.variant === "light") {
      useTheme.setTheme(ev.data.variant);
    }
  });
}

wireEngineEvents();
