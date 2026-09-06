// Tipos de UI de la cáscara de referencia (GUI Shell).
// Solo los tipos que necesita el render del shell; sin EngineEvent del agente.

export type EventType =
    | "text"
    | "tool_start"
    | "tool_result"
    | "tool_error"
    | "confirm"
    | "status"
    | "usage"
    | "done"
    | "error";

export interface UsageInfo {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
}

export interface ToolCallInfo {
    id: string;
    name: string;
    arguments: string;
}

export interface ConfirmInfo {
    tool_id: string;
    name: string;
    arguments: string;
}

export interface AppStatus {
    provider: string;
    model: string;
    language: string;
    theme: string;
    active_account: string;
    is_running: boolean;
    has_config: boolean;
    project_dir: string;
    work_mode: string;
    session_tokens: number;
    session_cost_usd: number;
    account_cumulative_tokens: number;
}

// Tool card state for the activity panel.
export interface ToolActivity {
    id: string;
    name: string;
    arguments: string;
    status: "running" | "done" | "error";
    result?: string;
    error?: string;
}

export interface ChatMessage {
    role: "user" | "assistant" | "tool";
    content: string;
    id: number;
    dbId?: number;
    usage?: UsageInfo | null;
    duration_ms?: number | null;
    model?: string | null;
    // Línea de actividad de herramienta incrustada en el chat.
    kind?: "tool";
    toolArgs?: string;
    toolStatus?: "running" | "done" | "error";
}

export interface Attachment {
    id: number;
    name: string;
    size: number;
    content?: string;
}

// ─── Configuración ────────────────────────────────────────────

export interface ProviderInfo {
    id: string;
    display_name: string;
    default_model: string;
    base_url: string;
}

export interface AccountView {
    name: string;
    provider: string;
    model: string;
    has_key: boolean;
    key_masked: string;
    active: boolean;
    tokens: number;
}

export interface BudgetView {
    max_tokens: number;
    max_cost_usd: number;
    warn_at: number;
}

export interface SearchView {
    enabled: boolean;
    provider: string;
    has_key: boolean;
    key_masked: string;
    key: string;
    cache_size: number;
    timeout: number;
}

export interface ConfigView {
    active_account?: string;
    language?: string;
    theme?: string;
    agent_timeout?: number;
    trusted_paths?: string[];
    project_dir?: string;
    budget?: BudgetView;
    safety_mode?: string;
    search?: SearchView;
    ui_zoom?: number;
    accounts?: AccountView[];
    config_path?: string;
    has_config?: boolean;
    // Multi-provider API keys and settings
    anthropic_api_key?: string;
    openai_api_key?: string;
    groq_api_key?: string;
    openrouter_api_key?: string;
    gemini_api_key?: string;
    deepseek_api_key?: string;
    ollama_base_url?: string;
    model?: string;
    base_url?: string;
    shell?: string;
    auto_approve?: Record<string, boolean>;
    max_tokens?: number;
    temperature?: number;
}

export type SettingsTab = "providers" | "general" | "safety" | "budget";

// ─── Proyectos y sesiones ─────────────────────────────────

export interface RecentProject {
    path: string;
    name: string;
    last_opened: string;
}

export type SpendCategory = "normal" | "web" | "test" | "skills";

export interface SpendEntry {
    ts: number;
    category: SpendCategory;
    model: string;
    tokens: number;
    cost_usd: number;
    label?: string;
}

export interface FileEntry {
    name: string;
    path: string;
    is_dir: boolean;
    size: number;
}

export interface GUISessionInfo {
    id: string;
    kind?: string;
    title: string;
    created_at: string;
    updated_at: string;
    message_count: number;
    project_path?: string;
}

// Consulta reciente del historial del Browser.
export interface BrowserHistoryEntry {
    query: string;
    at: number;
}

export interface GitDiffFile {
    path: string;
    name: string;
    dir: string;
    status: string;
    additions: number;
    deletions: number;
    staged: boolean;
}

export interface GitDiffSummary {
    is_repo: boolean;
    branch: string;
    clean: boolean;
    uncommitted_count: number;
    committed_count: number;
    files: GitDiffFile[];
    raw_diff: string;
}

export type ActivityTab = "overview" | "git" | "commands";

export interface CommandLogItem {
    id: string;
    name: string;
    kind: "tool" | "bash" | "mcp" | "skill";
    args?: string;
    status: "running" | "done" | "error";
    duration_ms?: number;
    timestamp: number;
    output?: string;
    diff?: string;
}

export interface BackgroundTask {
    id: string;
    title: string;
    status: "running" | "done" | "error";
    startTime: number;
}