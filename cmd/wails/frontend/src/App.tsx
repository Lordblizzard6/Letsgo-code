import {Component, lazy, memo, Suspense, useCallback, useEffect, useMemo, useRef, useState} from "react";
import {Events} from "@wailsio/runtime";
import {t, type Lang} from "./i18n";
import type {
    AccountView,
    ActivityTab,
    AppStatus,
    Attachment,
    BackgroundTask,
    BrowserHistoryEntry,
    ChatMessage,
    CommandLogItem,
    ConfigView,
    FileEntry,
    GitDiffFile,
    GitDiffSummary,
    GUISessionInfo,
    ProviderInfo,
    RecentProject,
    SettingsTab,
    ToolActivity,
    UsageInfo,
} from "./types";
import SettingsView from "./SettingsView";
import {SpendView} from "./SpendView";
import {OnboardingView} from "./OnboardingView";
import {ActivityInspector} from "./ActivityInspector";
import "./App.css";
import "highlight.js/styles/github-dark.css";
import goulmLogo from "./assets/images/goulm.png";
import {ChatService, GitService, SessionsService, SettingsService, ThemeService} from "./bindings";
import {ProviderIcon} from "./ProviderIcons";

// ─────────────────────────────────────────────────────────────
// LetsGo GUI Shell.
//
// Replica la estructura visual completa (header, sidebar, tabs,
// chat, composer, panel de actividad, modales) con datos de
// ejemplo y UI local navegable. NO tiene integración con ningún
// backend: es solo el esqueleto visual para reutilizar en otros
// proyectos. Sustituye App.tsx / SettingsView / SpendView del
// frontend original por versiones autocontenidas.
// ─────────────────────────────────────────────────────────────

// react-markdown + rehype-highlight se cargan bajo demanda (lazy).
const Markdown = lazy(() => import("./Markdown"));

// Row de mensaje memoizada, idéntica a la del frontend original.
const MessageRow = memo(function MessageRow({
    m,
    lang,
    isStreaming,
    copied,
    onCopyContent,
    onCopied,
    onRetry,
    onRollback,
}: {
    m: ChatMessage;
    lang: Lang;
    isStreaming: boolean;
    copied: boolean;
    onCopyContent: (content: string) => void;
    onCopied: (id: number) => void;
    onRetry: (m: ChatMessage) => void;
    onRollback?: (m: ChatMessage) => void;
}) {
    if (m.kind === "tool") {
        const st = m.toolStatus ?? "done";
        return (
            <div className={`tool-line ${st}`}>
                <span className={`tool-line-icon ${st}`}>
                    {st === "running" ? <SpinnerIcon /> : st === "error" ? <CrossIcon /> : <CheckIcon />}
                </span>
                <span className="tool-line-text">{m.content}</span>
            </div>
        );
    }
    const isError = m.role === "assistant" && m.content.startsWith("✖");
    const isUser = m.role === "user";
    return (
        <div className={`msg ${m.role}${isError ? " error" : ""}`}>
            {!isUser && (
                <div className="msg-head">
                    <span className="msg-avatar">LG</span>
                    <span className="msg-author">{t(lang, "chat.letsgo")}</span>
                </div>
            )}
            <ItemBoundary>
                {isUser ? (
                    <div className="msg-body">{m.content || "…"}</div>
                ) : (
                    <Suspense fallback={<div className="msg-body markdown">…</div>}>
                        <div className={`msg-body markdown${isStreaming ? " streaming" : ""}`}>
                            <Markdown>{m.content || "…"}</Markdown>
                        </div>
                    </Suspense>
                )}
            </ItemBoundary>
            <div className="msg-footer">
                {!isUser && (m.usage || m.duration_ms || m.model) && (
                    <div className="msg-meta">
                        {m.model && (
                            <span className="meta-item powered">
                                {t(lang, "app.powered_by")} <strong>{m.model}</strong>
                            </span>
                        )}
                        {m.duration_ms != null && <span className="meta-item">{fmtDuration(m.duration_ms)}</span>}
                        {m.usage?.total_tokens != null && (
                            <span className="meta-item">{t(lang, "msg.tokens", {n: m.usage.total_tokens})}</span>
                        )}
                    </div>
                )}
                <div className="msg-actions">
                    <button type="button"
                        className={`msg-action icon ${copied ? "copied" : ""}`}
                        onClick={() => {
                            onCopyContent(m.content);
                            onCopied(m.id);
                        }}
                        title={copied ? t(lang, "msg.copied") : t(lang, "msg.copy")}
                    >
                        {copied ? <CheckIcon /> : <CopyIcon />}
                    </button>
                    {isUser && (
                        <>
                            <button type="button" className="msg-action icon" onClick={() => onRetry(m)} title={t(lang, "msg.retry")}>
                                <RetryIcon />
                            </button>
                            {onRollback && (
                                <button type="button" className="msg-action icon rollback-btn" onClick={() => onRollback(m)} title="Rebobinar conversación hasta este mensaje">
                                    <UndoIcon />
                                </button>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    );
});

// ─── Iconos SVG (idénticos al original) ────────────────────

function PaperclipIcon() {
    return (
        <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21.44 11.05l-9.19 9.19a6 6 0 0 1-8.49-8.49l9.19-9.19a4 4 0 0 1 5.66 5.66l-9.2 9.19a2 2 0 0 1-2.83-2.83l8.49-8.48" />
        </svg>
    );
}

function SendIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
            <path d="M2.01 21 23 12 2.01 3 2 10l15 2-15 2z" />
        </svg>
    );
}

function StopIcon() {
    return (
        <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor">
            <rect x="6" y="6" width="12" height="12" rx="2" />
        </svg>
    );
}

function GearIcon() {
    return (
        <svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
        </svg>
    );
}

function SpinnerIcon() {
    return (
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round">
            <path d="M21 12a9 9 0 1 1-6.219-8.56" />
        </svg>
    );
}

function CheckIcon() {
    return (
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
            <path d="M20 6 9 17l-5-5" />
        </svg>
    );
}

function CrossIcon() {
    return (
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round">
            <path d="M18 6 6 18M6 6l12 12" />
        </svg>
    );
}

function CopyIcon() {
    return (
        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="9" y="9" width="13" height="13" rx="2" />
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
        </svg>
    );
}

function RetryIcon() {
    return (
        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M1 4v6h6" />
            <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10" />
        </svg>
    );
}

function CodeIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="m8 6-6 6 6 6M16 6l6 6-6 6" />
        </svg>
    );
}

function ArchIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M12 2 3 7l9 5 9-5z" />
            <path d="m3 12 9 5 9-5" />
            <path d="m3 17 9 5 9-5" />
        </svg>
    );
}

function PlanIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M9 6 11 8l4-4" />
            <path d="M9 13l2 2 4-4" />
            <path d="M9 20l2 2 4-4" />
            <path d="M3 4h18M3 11h18M3 18h18" />
        </svg>
    );
}

function ResearchIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="7" />
            <path d="m21 21-4.35-4.35" />
        </svg>
    );
}

function GlobeIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="12" cy="12" r="9" />
            <path d="M3 12h18M12 3a14 14 0 0 1 0 18 14 14 0 0 1 0-18z" />
        </svg>
    );
}

function FlaskIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M9 3h6M10 3v6l-5 9a2 2 0 0 0 1.8 3h10.4a2 2 0 0 0 1.8-3l-5-9V3" />
            <path d="M7.5 14h9" />
        </svg>
    );
}

function SkillIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M14.7 6.3a4.5 4.5 0 0 0-6.4 6.4L3 18v3h3l5.3-5.3a4.5 4.5 0 0 0 6.4-6.4L14 12.6l-2.6-2.6 3.3-3.7z" />
        </svg>
    );
}

function ArrowUpIcon({ size = 16 }: { size?: number }) {
    return (
        <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round">
            <line x1="12" y1="19" x2="12" y2="5" />
            <polyline points="5 12 12 5 19 12" />
        </svg>
    );
}

function PlusIcon({ size = 15 }: { size?: number }) {
    return (
        <svg viewBox="0 0 24 24" width={size} height={size} fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
    );
}

function ArrowDownIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M12 5v14M5 12l7 7 7-7" />
        </svg>
    );
}

function MenuIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
            <path d="M3 6h18M3 12h18M3 18h18" />
        </svg>
    );
}

function FolderIcon() {
    return (
        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
        </svg>
    );
}

function UndoIcon() {
    return (
        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M9 14L4 9l5-5" />
            <path d="M4 9h10.5a5.5 5.5 0 0 1 5.5 5.5v0a5.5 5.5 0 0 1-5.5 5.5H11" />
        </svg>
    );
}

function GitIcon() {
    return (
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="18" cy="18" r="3" />
            <circle cx="6" cy="6" r="3" />
            <path d="M6 9v12" />
            <path d="M18 15a9 9 0 0 0-9-9" />
        </svg>
    );
}

// ─── Boundaries y helpers ───────────────────────────────────

interface BoundaryProps {
    lang: Lang;
    children: React.ReactNode;
}

class ErrorBoundary extends Component<BoundaryProps, {error: Error | null}> {
    constructor(props: BoundaryProps) {
        super(props);
        this.state = {error: null};
    }

    static getDerivedStateFromError(error: Error) {
        return {error};
    }

    componentDidCatch(error: Error) {
        console.error("ErrorBoundary:", error);
    }

    render() {
        if (this.state.error) {
            return (
                <div id="error-fallback">
                    <h2>{t(this.props.lang, "settings.load_error")}</h2>
                    <pre>{this.state.error.message}</pre>
                    <button type="button" className="btn primary" onClick={() => this.setState({error: null})}>
                        {t(this.props.lang, "settings.cancel")}
                    </button>
                </div>
            );
        }
        return this.props.children;
    }
}

class ItemBoundary extends Component<{children: React.ReactNode}, {error: string | null}> {
    constructor(props: {children: React.ReactNode}) {
        super(props);
        this.state = {error: null};
    }

    static getDerivedStateFromError(error: Error) {
        return {error: error.message || String(error)};
    }

    componentDidCatch(error: Error) {
        console.error("ItemBoundary:", error);
    }

    render() {
        if (this.state.error) {
            return <div className="item-error">No se pudo mostrar este elemento: {this.state.error}</div>;
        }
        return this.props.children;
    }
}

interface CmdDef {
    name: string;
    args: string;
    descKey: string;
    hasArgs: boolean;
}

const COMMANDS: CmdDef[] = [
    {name: "/help", args: "", descKey: "cmd.desc_help", hasArgs: false},
    {name: "/clear", args: "", descKey: "cmd.desc_clear", hasArgs: false},
    {name: "/diff", args: "", descKey: "cmd.desc_diff", hasArgs: false},
    {name: "/review", args: "", descKey: "cmd.desc_review", hasArgs: false},
    {name: "/undo", args: "", descKey: "cmd.desc_undo", hasArgs: false},
    {name: "/redo", args: "", descKey: "cmd.desc_redo", hasArgs: false},
    {name: "/cost", args: "", descKey: "cmd.desc_cost", hasArgs: false},
    {name: "/compact", args: "", descKey: "cmd.desc_compact", hasArgs: false},
    {name: "/skills", args: "", descKey: "cmd.desc_skills", hasArgs: false},
    {name: "/tasks", args: "", descKey: "cmd.desc_tasks", hasArgs: false},
    {name: "/terminal", args: "", descKey: "cmd.desc_terminal", hasArgs: false},
    {name: "/share", args: "", descKey: "cmd.desc_share", hasArgs: false},
    {name: "/init", args: "", descKey: "cmd.desc_init", hasArgs: false},
    {name: "/login", args: "<api_key>", descKey: "cmd.desc_login", hasArgs: true},
    {name: "/model", args: "<modelo>", descKey: "cmd.desc_model", hasArgs: true},
    {name: "/rename", args: "<nuevo_nombre>", descKey: "cmd.desc_rename", hasArgs: true},
    {name: "/lang", args: "es|en", descKey: "cmd.desc_lang", hasArgs: true},
    {name: "/theme", args: "letsgo|dark|light", descKey: "cmd.desc_theme", hasArgs: true},
];

function basename(p: string): string {
    return p.split(/[\\/]/).filter(Boolean).pop() || p;
}

let msgId = 1000;
const nextId = () => ++msgId;

let tabSeq = 0;
const nextTabId = () => `tab-${++tabSeq}`;

const WORK_MODES = [
    {id: "planning", label: "Plan", icon: PlanIcon, i18n: "mode.planning", hint: "mode.planning_hint"},
    {id: "code", label: "Código", icon: CodeIcon, i18n: "mode.code", hint: "mode.code_hint"},
    {id: "architecture", label: "Arch", icon: ArchIcon, i18n: "mode.architecture", hint: "mode.architecture_hint"},
    {id: "research", label: "Research", icon: ResearchIcon, i18n: "mode.research", hint: "mode.research_hint"},
] as const;

function summarizeArgs(args: string): string {
    const one = args.replace(/\s+/g, " ").trim();
    return one.length > 64 ? one.slice(0, 64) + "…" : one || "…";
}

function fmtSize(bytes: number): string {
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function fmtDuration(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
}

function fmtDate(iso: string): string {
    if (!iso) return "";
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "";
    return d.toLocaleString(undefined, {month: "short", day: "numeric", hour: "2-digit", minute: "2-digit"});
}

function formatLastAccess(dateStr?: string): string {
    if (!dateStr) return "(0m)";
    const diffMs = Math.max(0, Date.now() - new Date(dateStr).getTime());
    const totalMinutes = Math.floor(diffMs / (60 * 1000));
    if (totalMinutes < 60) {
        return `(${totalMinutes}m)`;
    }
    const totalHours = Math.floor(totalMinutes / 60);
    if (totalHours < 24) {
        return `(${totalHours}h)`;
    }
    const totalDays = Math.floor(totalHours / 24);
    return `(${totalDays}d)`;
}

function parseStoredHistory(rawMessages: any[]): ChatMessage[] {
    const result: ChatMessage[] = [];
    if (!Array.isArray(rawMessages)) return result;

    for (const m of rawMessages) {
        const dbId = m.id || m.ID || undefined;
        let rawContent = m.content;
        if (typeof rawContent === "string" && (rawContent.startsWith("[") || rawContent.startsWith("{"))) {
            try {
                rawContent = JSON.parse(rawContent);
            } catch {}
        }

        const role = m.role || "assistant";

        // If it's an array of Anthropic content blocks
        if (Array.isArray(rawContent)) {
            for (const block of rawContent) {
                if (!block) continue;
                if (block.type === "text" && block.text) {
                    result.push({
                        id: nextId(),
                        dbId,
                        role: role === "user" ? "user" : "assistant",
                        content: block.text,
                    });
                } else if (block.type === "tool_use" || block.name) {
                    const toolName = block.name || "tool";
                    const argsStr = typeof block.input === "object" ? JSON.stringify(block.input) : String(block.input || "");
                    result.push({
                        id: nextId(),
                        dbId,
                        role: "assistant",
                        kind: "tool",
                        toolStatus: "done",
                        toolArgs: argsStr,
                        content: `${toolName}: ${summarizeArgs(argsStr)}`,
                    });
                } else if (block.type === "tool_result" || block.tool_result) {
                    const tr = block.tool_result || block;
                    const toolName = tr.tool_name || tr.name || "tool";
                    const isErr = !!tr.is_error;
                    const output = typeof tr.content === "string" ? tr.content : JSON.stringify(tr.content || "");
                    result.push({
                        id: nextId(),
                        dbId,
                        role: "assistant",
                        kind: "tool",
                        toolStatus: isErr ? "error" : "done",
                        content: `${toolName}: ${summarizeArgs(output)}`,
                    });
                }
            }
        } else if (typeof rawContent === "object" && rawContent !== null) {
            if (rawContent.type === "text" && rawContent.text) {
                result.push({
                    id: nextId(),
                    dbId,
                    role: role === "user" ? "user" : "assistant",
                    content: rawContent.text,
                });
            } else if (rawContent.type === "tool_result" || rawContent.tool_result) {
                const tr = rawContent.tool_result || rawContent;
                const toolName = tr.tool_name || tr.name || "tool";
                const isErr = !!tr.is_error;
                const output = typeof tr.content === "string" ? tr.content : JSON.stringify(tr.content || "");
                result.push({
                    id: nextId(),
                    dbId,
                    role: "assistant",
                    kind: "tool",
                    toolStatus: isErr ? "error" : "done",
                    content: `${toolName}: ${summarizeArgs(output)}`,
                });
            } else if (rawContent.type === "tool_use") {
                const toolName = rawContent.name || "tool";
                const argsStr = typeof rawContent.input === "object" ? JSON.stringify(rawContent.input) : String(rawContent.input || "");
                result.push({
                    id: nextId(),
                    dbId,
                    role: "assistant",
                    kind: "tool",
                    toolStatus: "done",
                    toolArgs: argsStr,
                    content: `${toolName}: ${summarizeArgs(argsStr)}`,
                });
            } else {
                result.push({
                    id: nextId(),
                    dbId,
                    role: role === "user" ? "user" : "assistant",
                    content: JSON.stringify(rawContent),
                });
            }
        } else {
            const text = String(rawContent || "").trim();
            if (text) {
                result.push({
                    id: nextId(),
                    dbId,
                    role: role === "user" ? "user" : "assistant",
                    content: text,
                });
            }
        }
    }

    return result;
}

function themeFor(configTheme?: string): string {
    if (configTheme === "light") return "light";
    if (configTheme === "dark") return "dark";
    return "letsgo";
}

// Elimina emojis del render del panel de actividad (las tools los incluyen).
function stripEmojis(text: string): string {
    return text.replace(
        /[\u{1F000}-\u{1FAFF}\u{2600}-\u{27BF}\u{FE0F}\u{2B00}-\u{2BFF}\u{2190}-\u{21FF}\u{2B50}\u{2705}\u{274C}\u{2714}\u{2716}\u{2728}]/gu,
        "",
    );
}

// ─── Tipos de estado del shell ─────────────────────────────

interface ChatTab {
    id: string;
    title: string;
    messages: ChatMessage[];
    tools: ToolActivity[];
    streaming: {msgId: number; text: string; usage: UsageInfo | null; started: number} | null;
    saved: boolean;
    sessionId: string | null;
}

function newTab(id: string): ChatTab {
    return {id, title: "", messages: [], tools: [], streaming: null, saved: false, sessionId: null};
}

interface BrowserPane {
    messages: ChatMessage[];
    tools: ToolActivity[];
    streaming: {msgId: number; text: string; usage: UsageInfo | null; started: number} | null;
    busy: boolean;
    status: {provider: string; model: string; account: string} | null;
    sessionId: string | null;
    title: string;
}

function newBrowser(): BrowserPane {
    return {messages: [], tools: [], streaming: null, busy: false, status: null, sessionId: null, title: ""};
}

function SpecialPaneView(props: {
    kind: "test" | "skills";
    pane: BrowserPane;
    lang: Lang;
    copiedId: number | null;
    onCopyContent: (content: string) => void;
    onCopied: (id: number) => void;
    onRetry: (m: ChatMessage) => void;
}) {
    const {kind, pane, lang, copiedId, onCopyContent, onCopied, onRetry} = props;
    const isTest = kind === "test";
    const title = t(lang, isTest ? "test.tab" : "skills.tab");
    const welcome = t(lang, isTest ? "test.welcome_title" : "skills.welcome_title");
    const subtitle = t(lang, isTest ? "test.welcome_subtitle" : "skills.welcome_subtitle");
    return (
        <>
            <div className={`browser-bar ${kind}-bar`}>
                <span className="browser-bar-title">
                    {isTest ? <FlaskIcon /> : <SkillIcon />} {title}
                </span>
                <div className="browser-bar-controls">
                    <span className="browser-bar-label">
                        {pane.status?.provider && pane.status.model
                            ? `${pane.status.provider} / ${pane.status.model}`
                            : t(lang, "browser.none")}
                    </span>
                </div>
            </div>
            {pane.messages.length === 0 && (
                <div id="chat-empty" className="browser-empty">
                    <div className="browser-empty-icon">{isTest ? <FlaskIcon /> : <SkillIcon />}</div>
                    <h1>{welcome}</h1>
                    <p>{subtitle}</p>
                </div>
            )}
            {pane.messages.map((m) => (
                <MessageRow
                    key={m.id}
                    m={m}
                    lang={lang}
                    isStreaming={false}
                    copied={copiedId === m.id}
                    onCopyContent={onCopyContent}
                    onCopied={onCopied}
                    onRetry={onRetry}
                />
            ))}
            {pane.busy && !pane.streaming && !pane.tools.some((tl) => tl.status === "running") && (
                <div className="thinking">
                    <span className="thinking-dots"><i /><i /><i /></span>
                    <span>{t(lang, "status.thinking_bubble")}</span>
                </div>
            )}
        </>
    );
}

const savedTheme = typeof window !== "undefined" ? localStorage.getItem("app_theme") || "letsgo" : "letsgo";
if (typeof document !== "undefined") {
    document.documentElement.dataset.theme = themeFor(savedTheme);
}

const defaultStatus: AppStatus = {
    provider: "anthropic",
    model: "claude-sonnet-4-20250514",
    language: "es",
    theme: savedTheme,
    active_account: "default",
    is_running: false,
    has_config: true,
    project_dir: "",
    work_mode: "code",
    session_tokens: 0,
    session_cost_usd: 0,
    account_cumulative_tokens: 0,
};

const defaultConfig: ConfigView = {
    accounts: [
        {name: "default", provider: "anthropic", model: "claude-sonnet-4-20250514", has_key: true, key_masked: "sk-ant-...", active: true, tokens: 0},
    ],
    providers: [
        {id: "anthropic", display_name: "Anthropic", default_model: "claude-sonnet-4-20250514", base_url: "https://api.anthropic.com"},
        {id: "openai", display_name: "OpenAI", default_model: "gpt-4o", base_url: "https://api.openai.com"},
        {id: "openrouter", display_name: "OpenRouter", default_model: "anthropic/claude-sonnet-4", base_url: "https://openrouter.ai/api"},
        {id: "ollama", display_name: "Ollama", default_model: "llama3.2", base_url: "http://localhost:11434"},
    ],
    budget: {max_tokens: 100000, max_cost_usd: 10, warn_at: 80},
    auto_approve: {},
};
function detectProvider(model: string, cfg: any): string {
    const m = (model || "").toLowerCase();
    if (m.includes("gemini")) return "gemini";
    if (m.includes("gpt") || m.includes("o1") || m.includes("o3")) return "openai";
    if (m.includes("claude")) return "anthropic";
    if (m.includes("deepseek")) return "deepseek";
    if (m.includes("groq") || m.includes("llama-3.3") || m.includes("llama-3.1")) return "groq";
    if (m.includes("llama") || m.includes("qwen") || m.includes("mistral")) return cfg?.ollama_base_url ? "ollama" : "openrouter";
    if (cfg?.gemini_api_key || cfg?.GeminiAPIKey) return "gemini";
    if (cfg?.groq_api_key || cfg?.GroqAPIKey) return "groq";
    if (cfg?.openai_api_key || cfg?.OpenAIAPIKey) return "openai";
    if (cfg?.openrouter_api_key || cfg?.OpenRouterAPIKey) return "openrouter";
    if (cfg?.deepseek_api_key || cfg?.DeepSeekAPIKey) return "deepseek";
    if (cfg?.anthropic_api_key || cfg?.AnthropicAPIKey) return "anthropic";
    if (cfg?.ollama_base_url || cfg?.OllamaBaseURL) return "ollama";
    return "gemini";
}

function buildAccounts(cfg: any): AccountInfo[] {
    const accs: AccountInfo[] = [];
    if (cfg.gemini_api_key || cfg.GeminiAPIKey) {
        accs.push({name: "Google Gemini", provider: "gemini", model: cfg.model?.includes("gemini") ? cfg.model : "gemini-2.0-flash", has_key: true, key_masked: "AIza••••", active: false, tokens: 0});
    }
    if (cfg.anthropic_api_key || cfg.AnthropicAPIKey) {
        accs.push({name: "Anthropic", provider: "anthropic", model: cfg.model?.includes("claude") ? cfg.model : "claude-3-5-sonnet-20241022", has_key: true, key_masked: "sk-ant••••", active: false, tokens: 0});
    }
    if (cfg.openai_api_key || cfg.OpenAIAPIKey) {
        accs.push({name: "OpenAI", provider: "openai", model: cfg.model?.includes("gpt") ? cfg.model : "gpt-4o", has_key: true, key_masked: "sk-••••", active: false, tokens: 0});
    }
    if (cfg.groq_api_key || cfg.GroqAPIKey) {
        accs.push({name: "Groq", provider: "groq", model: cfg.model?.includes("versatile") ? cfg.model : "llama-3.3-70b-versatile", has_key: true, key_masked: "gsk_••••", active: false, tokens: 0});
    }
    if (cfg.openrouter_api_key || cfg.OpenRouterAPIKey) {
        accs.push({name: "OpenRouter", provider: "openrouter", model: cfg.model?.includes("/") ? cfg.model : "anthropic/claude-3.5-sonnet", has_key: true, key_masked: "sk-or••••", active: false, tokens: 0});
    }
    if (cfg.deepseek_api_key || cfg.DeepSeekAPIKey) {
        accs.push({name: "DeepSeek", provider: "deepseek", model: cfg.model?.includes("deepseek") ? cfg.model : "deepseek-chat", has_key: true, key_masked: "sk-••••", active: false, tokens: 0});
    }
    if (cfg.ollama_base_url || cfg.OllamaBaseURL) {
        accs.push({name: "Ollama (Local)", provider: "ollama", model: "llama3.2", has_key: true, key_masked: "local", active: false, tokens: 0});
    }
    if (accs.length === 0) {
        accs.push({name: "Default", provider: "anthropic", model: "claude-3-5-sonnet-20241022", has_key: false, key_masked: "none", active: true, tokens: 0});
    }
    return accs;
}

// ─── App ────────────────────────────────────────────────────

function App() {
    const [status, setStatus] = useState<AppStatus>(defaultStatus);
    const [config, setConfig] = useState<ConfigView>(defaultConfig);
    const [tabs, setTabs] = useState<ChatTab[]>(() => [newTab(nextTabId())]);
    const [activeTabId, setActiveTabId] = useState(tabs[0]?.id ?? "tab-1");

    const applyMessages = useCallback((fn: (msgs: ChatMessage[]) => ChatMessage[]) => {
        setTabs((prev) => prev.map((tb) => (tb.id === activeTabId ? {...tb, messages: fn(tb.messages)} : tb)));
    }, [activeTabId]);

    const applyTools = useCallback((fn: (tools: ToolActivity[]) => ToolActivity[]) => {
        setTabs((prev) => prev.map((tb) => (tb.id === activeTabId ? {...tb, tools: fn(tb.tools)} : tb)));
    }, [activeTabId]);

    const [input, setInput] = useState("");
    const [attachments, setAttachments] = useState<Attachment[]>([]);
    const [busy, setBusy] = useState(false);
    const [streaming, setStreaming] = useState(false);
    const [statusLabel, setStatusLabel] = useState(() => t("es", "status.ready"));
    const [errorState, setErrorState] = useState(false);
    const [lastError, setLastError] = useState<string | null>(null);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [onboardingOpen, setOnboardingOpen] = useState(false);
    const [spendOpen, setSpendOpen] = useState(false);
    const [pendingToolApproval, setPendingToolApproval] = useState<{callId: string; name: string; input: any} | null>(null);
    const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
    const [recentProjects, setRecentProjects] = useState<RecentProject[]>(() => {
        try {
            const raw = localStorage.getItem("recent_projects");
            return raw ? JSON.parse(raw) : [];
        } catch {
            return [];
        }
    });
    const [sessions, setSessions] = useState<GUISessionInfo[]>([]);
    const [projectMenuOpen, setProjectMenuOpen] = useState(false);
    const [explorerOpen, setExplorerOpen] = useState(false);
    const [explorerEntries, setExplorerEntries] = useState<FileEntry[]>([]);
    const [explorerPath, setExplorerPath] = useState("");
    const [headerModelMenu, setHeaderModelMenu] = useState(false);
    const [headerModels, setHeaderModels] = useState<string[]>([]);
    const [headerAccountMenu, setHeaderAccountMenu] = useState(false);
    const [modeMenuOpen, setModeMenuOpen] = useState(false);
    const [paletteOpen, setPaletteOpen] = useState(false);
    const [paletteQuery, setPaletteQuery] = useState("");
    const [paletteSel, setPaletteSel] = useState(0);
    const viewportRef = useRef<HTMLDivElement>(null);
    const activityRef = useRef<HTMLDivElement>(null);
    const fileRef = useRef<HTMLInputElement>(null);
    const composerRef = useRef<HTMLTextAreaElement>(null);
    const paletteRef = useRef<HTMLInputElement>(null);
    const [cmdSel, setCmdSel] = useState(0);
    const [paletteDismissed, setPaletteDismissed] = useState(false);
    const [activePane, setActivePane] = useState<"none" | "browser" | "test" | "skills">("none");
    const browserActive = activePane === "browser";
    const testActive = activePane === "test";
    const skillsActive = activePane === "skills";
    const specialActive = browserActive || testActive || skillsActive;
    const [browser, setBrowser] = useState<BrowserPane>(newBrowser());
    const [browserInput, setBrowserInput] = useState("");
    const [browserModelMenu, setBrowserModelMenu] = useState(false);
    const [browserModels, setBrowserModels] = useState<string[]>([]);
    const [browserHistory, setBrowserHistory] = useState<BrowserHistoryEntry[]>([]);
    const [test, setTest] = useState<BrowserPane>(newBrowser());
    const [testInput, setTestInput] = useState("");
    const [skills, setSkills] = useState<BrowserPane>(newBrowser());
    const [skillsInput, setSkillsInput] = useState("");
    const [sessionUsage, setSessionUsage] = useState({tokens: 0, cost: 0, accum: 0});
    const tabTokensRef = useRef<Map<string, number>>(new Map());
    const [stickToBottom, setStickToBottom] = useState(true);
    const [showJumpDown, setShowJumpDown] = useState(false);
    const [activityStick, setActivityStick] = useState(true);
    const toolLineRef = useRef<Map<string, number>>(new Map());
    const [zoom, setZoom] = useState(100);
    const zoomRef = useRef(100);
    const [copiedId, setCopiedId] = useState<number | null>(null);

    const [activityInspectorOpen, setActivityInspectorOpen] = useState(false);
    const [activityTab, setActivityTab] = useState<ActivityTab>("overview");
    const [activityMaximized, setActivityMaximized] = useState(false);
    const [commandLogs, setCommandLogs] = useState<CommandLogItem[]>([]);
    const [backgroundTasks, setBackgroundTasks] = useState<BackgroundTask[]>([]);
    const [skillsUsed, setSkillsUsed] = useState<string[]>(["speckit-plan", "speckit-implement"]);

    const [gitSummary, setGitSummary] = useState<GitDiffSummary | null>(null);
    const [gitSelectedFile, setGitSelectedFile] = useState<string | null>(null);
    const [gitFileDiff, setGitFileDiff] = useState<string | null>(null);
    const [gitLoading, setGitLoading] = useState(false);

    const refreshGitDiff = useCallback(async () => {
        setGitLoading(true);
        try {
            const summary = await GitService.DiffSummary();
            setGitSummary(summary);
            if (summary.files && summary.files.length > 0) {
                const target = summary.files.find((f) => f.path === gitSelectedFile) ? gitSelectedFile! : summary.files[0].path;
                setGitSelectedFile(target);
                const fileDiff = await GitService.DiffFile(target);
                setGitFileDiff(fileDiff);
            } else {
                setGitSelectedFile(null);
                setGitFileDiff(null);
            }
        } catch (e) {
            console.error("Failed to load git diff:", e);
        } finally {
            setGitLoading(false);
        }
    }, [gitSelectedFile]);

    const selectGitFile = useCallback(async (filePath: string) => {
        setGitSelectedFile(filePath);
        setGitLoading(true);
        try {
            const diff = await GitService.DiffFile(filePath);
            setGitFileDiff(diff);
        } catch (e) {
            console.error("Failed to load file diff:", e);
        } finally {
            setGitLoading(false);
        }
    }, []);

    const handleStageFile = useCallback(async (filePath: string) => {
        try {
            await GitService.StageFile(filePath);
            await refreshGitDiff();
        } catch (e) {
            console.error("Failed to stage file:", e);
        }
    }, [refreshGitDiff]);

    const handleUnstageFile = useCallback(async (filePath: string) => {
        try {
            await GitService.UnstageFile(filePath);
            await refreshGitDiff();
        } catch (e) {
            console.error("Failed to unstage file:", e);
        }
    }, [refreshGitDiff]);

    const handleStageAll = useCallback(async () => {
        try {
            await GitService.StageAll();
            await refreshGitDiff();
        } catch (e) {
            console.error("Failed to stage all files:", e);
        }
    }, [refreshGitDiff]);

    const handleUnstageAll = useCallback(async () => {
        try {
            await GitService.UnstageAll();
            await refreshGitDiff();
        } catch (e) {
            console.error("Failed to unstage all files:", e);
        }
    }, [refreshGitDiff]);

    const handleCommit = useCallback(async (msg: string) => {
        try {
            await GitService.Commit(msg);
            await refreshGitDiff();
            applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✓ Cambios confirmados con commit: "${msg}"`}]);
        } catch (e: any) {
            console.error("Failed to commit:", e);
            applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✖ Error al confirmar cambios: ${e}`}]);
        }
    }, [refreshGitDiff, applyMessages]);

    useEffect(() => {
        if (activityInspectorOpen) {
            refreshGitDiff();
        }
    }, [activityInspectorOpen, refreshGitDiff]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if ((e.ctrlKey || e.metaKey) && (e.key.toLowerCase() === "d" || e.key.toLowerCase() === "b")) {
                e.preventDefault();
                setActivityInspectorOpen((v) => !v);
            }
        };
        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, []);

    useEffect(() => {
        const theme = status ? themeFor(status.theme) : savedTheme;
        document.documentElement.dataset.theme = theme;
    }, [status]);

    useEffect(() => {
        try {
            localStorage.setItem("recent_projects", JSON.stringify(recentProjects));
        } catch {}
    }, [recentProjects]);

    useEffect(() => {
        document.documentElement.style.zoom = "";
        const root = document.getElementById("root");
        if (root) {
            root.style.height = "100%";
            root.style.width = "100%";
        }
    }, [zoom]);

    const appZoomStyle = useMemo(() => {
        const scale = zoom / 100;
        return {
            zoom: scale,
            width: `${100 / scale}%`,
            height: `${100 / scale}%`,
        };
    }, [zoom]);

    const lang: Lang = status?.language === "es" ? "es" : "en";
    const currentMode = WORK_MODES.find((m) => m.id === (status?.work_mode ?? "code")) ?? WORK_MODES[0];
    const activeTab = tabs.find((tb) => tb.id === activeTabId) ?? tabs[0];
    const messages = activeTab?.messages ?? [];
    const tools = activeTab?.tools ?? [];

    const handleRollback = useCallback(
        async (targetMsg: ChatMessage) => {
            const sid = activeTab?.sessionId || "";
            if (!sid) {
                const idx = activeTab?.messages ? activeTab.messages.findIndex((m) => m.id === targetMsg.id) : -1;
                if (idx >= 0 && activeTab) {
                    setInput(targetMsg.content || "");
                    setTabs((prev) =>
                        prev.map((t) => (t.id === activeTab.id ? {...t, messages: t.messages.slice(0, idx)} : t))
                    );
                    composerRef.current?.focus();
                }
                return;
            }

            try {
                const targetId = targetMsg.dbId ?? 0;
                const restoredPrompt = await SessionsService.Rollback(sid, targetId);
                if (restoredPrompt) {
                    setInput(restoredPrompt);
                } else if (targetMsg.content) {
                    setInput(targetMsg.content);
                }
                const history = await SessionsService.Open(sid);
                const chatMsgs = parseStoredHistory(history);
                setTabs((prev) =>
                    prev.map((t) => (t.id === activeTab?.id ? {...t, messages: chatMsgs} : t))
                );
                composerRef.current?.focus();
            } catch (err) {
                console.error("Failed to rollback message:", err);
            }
        },
        [activeTab]
    );

    // ─── Scroll ──────────────────────────────────────────────
    const viewMessages = browserActive ? browser.messages : testActive ? test.messages : skillsActive ? skills.messages : messages;
    const viewTools = browserActive ? browser.tools : testActive ? test.tools : skillsActive ? skills.tools : tools;
    useEffect(() => {
        if (viewportRef.current && stickToBottom) {
            viewportRef.current.scrollTop = viewportRef.current.scrollHeight;
        }
    }, [viewMessages, viewTools, stickToBottom]);

    useEffect(() => {
        if (activityRef.current && activityStick) {
            activityRef.current.scrollTop = activityRef.current.scrollHeight;
        }
    }, [viewTools, activityStick]);

    const onViewportScroll = useCallback(() => {
        const el = viewportRef.current;
        if (!el) return;
        const near = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
        setStickToBottom(near);
        setShowJumpDown(!near);
    }, []);

    const onActivityScroll = useCallback(() => {
        const el = activityRef.current;
        if (!el) return;
        setActivityStick(el.scrollHeight - el.scrollTop - el.clientHeight < 40);
    }, []);

    const jumpToBottom = useCallback(() => {
        const el = viewportRef.current;
        if (!el) return;
        el.scrollTo({top: el.scrollHeight, behavior: "smooth"});
    }, []);

    // ─── Cuenta / modelo ─────────────────────────────────────
    const changeHeaderModel = useCallback((model: string) => {
        setStatus((s) => s ? {...s, model} : s);
        setHeaderModelMenu(false);
        SettingsService.SaveConfig({model, Model: model}).catch(() => {});
    }, []);

    const refreshModels = useCallback((prov?: string) => {
        const provider = prov || status?.provider || "gemini";
        SettingsService.DiscoverModels(provider).then((models: any) => {
            if (Array.isArray(models) && models.length > 0) {
                const list = models.map((m: any) => typeof m === "string" ? m : (m.id || m.name || String(m)));
                setHeaderModels(list);
            }
        }).catch(() => {});
    }, [status?.provider]);

    useEffect(() => {
        if (status?.provider) {
            refreshModels(status.provider);
        }
    }, [status?.provider, refreshModels]);

    const changeHeaderAccount = useCallback((name: string) => {
        const acc = config?.accounts.find((a) => a.name === name);
        if (!acc) return;
        setStatus((s) => s ? {...s, active_account: name, provider: acc.provider, model: acc.model} : s);
        setHeaderAccountMenu(false);
        setHeaderModelMenu(false);
        refreshModels(acc.provider);
        SettingsService.SaveConfig({model: acc.model, Model: acc.model}).catch(() => {});
    }, [config, refreshModels]);

    // ─── Carga inicial de backend ──────────────────────────
    useEffect(() => {
        SettingsService.GetConfig().then((cfg: any) => {
            if (cfg) {
                const accounts = buildAccounts(cfg);
                const activeModel = cfg.model || cfg.Model || "gemini-2.0-flash";
                const activeProvider = detectProvider(activeModel, cfg);
                const activeAcc = accounts.find((a) => a.provider === activeProvider) || accounts[0];

                const th = cfg.theme || cfg.ThemeVariant || savedTheme;
                localStorage.setItem("app_theme", th);
                document.documentElement.dataset.theme = themeFor(th);
                setConfig((prev) => ({...prev, ...cfg, accounts, theme: th}));
                setStatus((s) => ({
                    ...s,
                    model: activeModel,
                    provider: activeProvider,
                    active_account: activeAcc ? activeAcc.name : s.active_account,
                    language: cfg.language || s.language,
                    theme: th,
                }));

                const hasKey = cfg.anthropic_api_key || cfg.AnthropicAPIKey ||
                    cfg.openai_api_key || cfg.OpenAIAPIKey ||
                    cfg.groq_api_key || cfg.GroqAPIKey ||
                    cfg.openrouter_api_key || cfg.OpenRouterAPIKey ||
                    cfg.gemini_api_key || cfg.GeminiAPIKey ||
                    cfg.deepseek_api_key || cfg.DeepSeekAPIKey ||
                    cfg.api_key || cfg.APIKey ||
                    cfg.ollama_base_url || cfg.OllamaBaseURL;
                if (!hasKey) {
                    setOnboardingOpen(true);
                }
            }
        }).catch(() => {});

        SessionsService.List().then((list: any) => {
            if (Array.isArray(list) && list.length > 0) {
                setSessions(list.map((s: any) => ({
                    id: s.id || s.ID,
                    title: s.name || s.Name || "Sesión",
                    created_at: s.created_at || s.CreatedAt || "",
                    updated_at: s.updated_at || s.UpdatedAt || "",
                    message_count: s.message_count || s.MessageCount || 0,
                    project_path: s.project_path || s.ProjectPath || "",
                })));
            }
        }).catch(() => {});

        SessionsService.ListProjects().then((projs: any) => {
            if (Array.isArray(projs) && projs.length > 0) {
                setRecentProjects((prev) => {
                    const existing = new Set(prev.map((p) => p.path));
                    const toAdd = projs
                        .filter((p: string) => p && !existing.has(p))
                        .map((p: string) => ({name: basename(p), path: p, last_opened: new Date().toISOString()}));
                    return [...prev, ...toAdd];
                });
            }
        }).catch(() => {});

        SessionsService.GetCurrentProject().then((proj: string) => {
            if (proj) {
                setStatus((s) => s ? ({...s, project_dir: proj}) : s);
                setRecentProjects((prev) => {
                    const base = basename(proj);
                    const filtered = prev.filter((p) => p.path !== proj);
                    return [{name: base, path: proj, last_opened: new Date().toISOString()}, ...filtered].slice(0, 10);
                });
            }
        }).catch(() => {});
    }, []);

    // ─── Wails event listeners ──────────────────────────────
    useEffect(() => {
        const unsubs = [
            Events.On("status:update", (ev: any) => setStatus(ev.data)),
            Events.On("config:update", (ev: any) => setConfig(ev.data)),
            Events.On("project:changed", (ev: any) => {
                const data = ev.data?.[0] || ev.data || ev;
                const proj = data?.project_path ?? data?.ProjectPath ?? "";
                setStatus((s) => s ? ({...s, project_dir: proj}) : s);
                if (proj) {
                    setRecentProjects((prev) => {
                        const base = basename(proj);
                        const filtered = prev.filter((p) => p.path !== proj);
                        return [{name: base, path: proj, last_opened: new Date().toISOString()}, ...filtered].slice(0, 10);
                    });
                }
            }),
            Events.On("theme:changed", (ev: any) => {
                const data = ev.data?.[0] || ev.data;
                const th = data?.theme;
                if (th) {
                    localStorage.setItem("app_theme", th);
                    setStatus((s) => s ? ({...s, theme: th}) : s);
                    document.documentElement.dataset.theme = themeFor(th);
                }
            }),
            Events.On("session:loaded", (ev: any) => {
                const data = ev.data?.[0] || ev.data || ev;
                const sid = data?.session_id || data?.SessionID;
                const msgs = data?.messages || data?.Messages;
                const ppath = data?.project_path || data?.ProjectPath;
                if (ppath) {
                    setStatus((s) => s ? ({...s, project_dir: ppath}) : s);
                }
                if (sid && Array.isArray(msgs)) {
                    const parsed = parseStoredHistory(msgs);
                    setTabs((prev) => prev.map((tb) => tb.sessionId === sid ? {...tb, messages: parsed} : tb));
                }
            }),
            Events.On("session:list", (ev: any) => {
                const data = ev.data?.[0] || ev.data;
                if (Array.isArray(data)) {
                    setSessions(data.map((s: any) => ({
                        id: s.id || s.ID,
                        title: s.name || s.Name || "Sesión",
                        created_at: s.created_at || s.CreatedAt || "",
                        updated_at: s.updated_at || s.UpdatedAt || "",
                        message_count: s.message_count || s.MessageCount || 0,
                        project_path: s.project_path || s.ProjectPath || "",
                    })));
                }
            }),
            Events.On("usage:update", (ev: any) => {
                const data = ev.data?.[0] || ev.data;
                if (data) {
                    setSessionUsage((prev) => ({
                        tokens: (data.input || 0) + (data.output || 0),
                        cost: data.cost || prev.cost,
                        accum: prev.accum + (data.cost || 0),
                    }));
                }
            }),
            Events.On("chat:start", (_ev: any) => {
                setBusy(true);
                setStreaming(true);
                setStatusLabel(t(lang, "status.thinking"));
            }),
            Events.On("chat:delta", (ev: any) => {
                const data = ev.data?.[0] || ev.data || ev;
                const text = typeof data === "string" ? data : (data?.text || data?.Text || "");
                if (!text) return;
                setStreaming(true);
                applyMessages((prev) => {
                    const last = prev[prev.length - 1];
                    if (last && last.role === "assistant" && !last.kind) {
                        return [...prev.slice(0, -1), {...last, content: last.content + text}];
                    }
                    return [...prev, {id: nextId(), role: "assistant", content: text}];
                });
            }),
            Events.On("chat:end", (_ev: any) => {
                setStreaming(false);
                setStatusLabel(t(lang, "status.ready"));
            }),
            Events.On("chat:idle", () => {
                setBusy(false);
                setStreaming(false);
                setStatusLabel(t(lang, "status.ready"));
            }),
            Events.On("chat:cancelled", () => {
                setBusy(false);
                setStreaming(false);
                setStatusLabel(t(lang, "status.cancelled"));
            }),
            Events.On("chat:error", (ev: any) => {
                const data = ev.data?.[0] || ev.data || ev;
                const msg = data?.message || "Unknown error";
                setErrorState(true);
                setLastError(msg);
                setBusy(false);
                setStreaming(false);
                setStatusLabel(t(lang, "status.error"));
                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✖ ${msg}`}]);
            }),
            Events.On("tool:start", (ev: any) => {
                const data = ev.data?.[0] || ev.data || ev;
                const callId = data?.call_id || nextId().toString();
                const name = data?.name || "tool";
                const kind = name === "bash" ? "bash" : name.startsWith("mcp_") ? "mcp" : "tool";
                applyMessages((prev) => [
                    ...prev,
                    {id: nextId(), role: "assistant", kind: "tool", content: `Executing tool ${name} (${callId})`, toolStatus: "running"},
                ]);
                setCommandLogs((prev) => [
                    {
                        id: callId,
                        name,
                        kind,
                        args: typeof data?.input === "object" ? JSON.stringify(data.input) : String(data?.input || ""),
                        status: "running",
                        timestamp: Date.now(),
                    },
                    ...prev.slice(0, 49),
                ]);
            }),
            Events.On("tool:end", (ev: any) => {
                const data = ev.data?.[0] || ev.data || ev;
                const content = data?.content || "done";
                const isErr = !!data?.is_error;
                const callId = data?.call_id || "";
                const name = data?.name || "tool";
                applyMessages((prev) => [
                    ...prev,
                    {
                        id: nextId(),
                        role: "assistant",
                        kind: "tool",
                        content: isErr ? `Error (${name}): ${content}` : `Result (${name}): ${content}`,
                        toolStatus: isErr ? "error" : "done",
                    },
                ]);
                setCommandLogs((prev) =>
                    prev.map((c) =>
                        c.id === callId || (c.name === name && c.status === "running")
                            ? {
                                  ...c,
                                  status: isErr ? "error" : "done",
                                  output: typeof content === "string" ? content : JSON.stringify(content),
                                  duration_ms: Date.now() - c.timestamp,
                              }
                            : c
                    )
                );
                if (name === "write" || name === "edit" || name === "bash") {
                    refreshGitDiff();
                }
            }),
            Events.On("tool:request", (ev: any) => {
                const data = ev.data?.[0] || ev.data || ev;
                const callId = data?.call_id;
                const name = data?.name || "herramienta";
                const input = data?.input;
                if (callId) {
                    setPendingToolApproval({callId, name, input});
                }
            }),
        ];
        return () => unsubs.forEach((u) => { if (typeof u === "function") u(); });
    }, [lang, applyMessages]);

    // ─── Tabs ────────────────────────────────────────────────
    const selectChatTab = useCallback(
        (id: string) => {
            setActiveTabId(id);
            setActivePane("none");
            const tabTokens = tabTokensRef.current.get(id) ?? 0;
            setSessionUsage((u) => ({...u, tokens: tabTokens}));
        },
        [],
    );

    const selectWebTab = useCallback(() => setActivePane("browser"), []);
    const selectTestTab = useCallback(() => setActivePane("test"), []);
    const selectSkillsTab = useCallback(() => setActivePane("skills"), []);

    const addTab = useCallback(async () => {
        const id = nextTabId();
        tabTokensRef.current.set(id, 0);
        setSessionUsage((u) => ({...u, tokens: 0}));
        const currentProject = status?.project_dir || "";
        const dateStr = new Date().toISOString().slice(0, 10);
        const base = currentProject ? basename(currentProject) : "General";
        const sessionName = `[${dateStr}] ${base}`;
        let sessId: string | null = null;
        try {
            const newSess = await SessionsService.Create(sessionName, currentProject);
            sessId = (newSess as any)?.id || (newSess as any)?.ID || null;
            const list = await SessionsService.List();
            if (Array.isArray(list)) {
                setSessions(list.map((s: any) => ({
                    id: s.id || s.ID,
                    title: s.name || s.Name || "Sesión",
                    created_at: s.created_at || s.CreatedAt || "",
                    updated_at: s.updated_at || s.UpdatedAt || "",
                    message_count: s.message_count || s.MessageCount || 0,
                    project_path: s.project_path || s.ProjectPath || "",
                })));
            }
        } catch (e) {
            console.error("addTab create session error:", e);
        }
        setTabs((prev) => [...prev, {
            ...newTab(id),
            title: sessionName,
            sessionId: sessId,
        }]);
        setActiveTabId(id);
        setActivePane("none");
    }, [status?.project_dir]);

    const closeTab = useCallback(
        (id: string) => {
            tabTokensRef.current.delete(id);
            const idx = tabs.findIndex((tb) => tb.id === id);
            if (idx < 0) return;
            const next = tabs.filter((tb) => tb.id !== id);
            if (next.length === 0) {
                addTab();
                return;
            }
            setTabs(next);
            if (id === activeTabId) {
                const target = next[Math.max(0, idx - 1)].id;
                setActiveTabId(target);
                setSessionUsage((u) => ({...u, tokens: tabTokensRef.current.get(target) ?? 0}));
            }
        },
        [tabs, activeTabId, addTab],
    );

    // ─── Explorador ──────────────────────────────────────────
    const loadExplorerDir = useCallback(async (sub: string) => {
        // TODO: const entries = await ListDirectory(sub);
        // setExplorerEntries(entries);
        // setExplorerPath(sub);
        // setExplorerOpen(true);
        setExplorerPath(sub);
        setExplorerOpen(true);
        return true;
    }, []);

    const openExplorer = useCallback(async () => {
        if (explorerOpen) {
            setExplorerOpen(false);
            return;
        }
        await loadExplorerDir("");
    }, [explorerOpen, loadExplorerDir]);

    const explorerOpenDir = useCallback(
        (entry: FileEntry) => {
            if (!entry.is_dir) return;
            const sub = explorerPath ? `${explorerPath}/${entry.name}` : entry.name;
            loadExplorerDir(sub);
        },
        [explorerPath, loadExplorerDir],
    );

    const explorerGoUp = useCallback(() => {
        if (!explorerPath) return;
        const idx = explorerPath.lastIndexOf("/");
        loadExplorerDir(idx < 0 ? "" : explorerPath.slice(0, idx));
    }, [explorerPath, loadExplorerDir]);

    const send = useCallback(async () => {
        const raw = input.trim();
        if (!raw || busy) return;
        const isCommand = raw.startsWith("/") && !raw.startsWith("//");
        const text = isCommand ? raw : raw.startsWith("//") ? raw.slice(1) : raw;
        setInput("");
        setPaletteDismissed(false);
        const userMsg: ChatMessage = {id: nextId(), role: "user", content: text};
        applyMessages((prev) => [...prev, userMsg]);
        setTabs((prev) =>
            prev.map((tb) => {
                if (tb.id !== activeTabId || tb.title) return tb;
                const runes = Array.from(userMsg.content.trim());
                const title = runes.length > 40 ? runes.slice(0, 40).join("") + "…" : runes.join("");
                return {...tb, title: title || t(lang, "app.new_chat")};
            }),
        );
        if (isCommand) {
            if (raw.startsWith("/clear")) {
                setTabs((prev) =>
                    prev.map((tb) => (tb.id === activeTabId ? {...tb, messages: [], tools: [], streaming: null} : tb)),
                );
                tabTokensRef.current.set(activeTabId, 0);
                setSessionUsage((u) => ({...u, tokens: 0}));
                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: t(lang, "cmd.cleared")}]);
                return;
            }
            if (raw.startsWith("/theme")) {
                const v = raw.split(/\s+/)[1] || "";
                const nextTheme = v === "light" ? "light" : v === "dark" ? "dark" : "letsgo";
                localStorage.setItem("app_theme", nextTheme);
                setStatus((s) => s ? {...s, theme: nextTheme} : s);
                document.documentElement.dataset.theme = themeFor(nextTheme);
                ThemeService.Set(nextTheme).catch(() => {});
                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✓ Tema cambiado a: **${nextTheme}**`}]);
                return;
            }
            if (raw.startsWith("/model") || raw.startsWith("/models")) {
                const parts = raw.split(/\s+/);
                if (parts.length > 1 && parts[1]) {
                    const newModel = parts[1];
                    changeHeaderModel(newModel);
                    applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✓ Modelo activo cambiado a: \`${newModel}\``}]);
                } else {
                    const modelsList = headerModels.length > 0
                        ? headerModels.map((m) => `• \`${m}\``).join("\n")
                        : "No se encontraron modelos. Verifica tus claves en Ajustes.";
                    applyMessages((prev) => [...prev, {
                        id: nextId(),
                        role: "assistant",
                        content: `### 🤖 Modelos disponibles (${status?.provider || "activo"}):\n\n${modelsList}\n\n*Usa \`/model <nombre>\` o presiona Tab en el compositor para autocompletar.*`,
                    }]);
                }
                return;
            }
            if (raw.startsWith("/rename")) {
                const newName = raw.replace(/^\/rename\s*/, "").trim();
                if (!newName) {
                    applyMessages((prev) => [...prev, {
                        id: nextId(),
                        role: "assistant",
                        content: "Uso: `/rename <nuevo nombre>` para cambiar el nombre de la sesión activa.",
                    }]);
                    return;
                }
                const targetSid = activeTab?.sessionId;
                if (targetSid) {
                    try {
                        await SessionsService.Rename(targetSid, newName);
                    } catch (e) {
                        console.error("Failed to rename session:", e);
                    }
                }
                setTabs((prev) => prev.map((tb) => tb.id === activeTabId ? {...tb, title: newName} : tb));
                setSessions((prev) => prev.map((s) => s.id === targetSid ? {...s, title: newName} : s));
                applyMessages((prev) => [...prev, {
                    id: nextId(),
                    role: "assistant",
                    content: `✓ Sesión renombrada a: **${newName}**`,
                }]);
                return;
            }
            if (raw.startsWith("/diff")) {
                setActivityInspectorOpen(true);
                setActivityTab("git");
                refreshGitDiff();
                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: "✓ Panel de Git Review abierto."}]);
                return;
            }
            if (raw.startsWith("/cost")) {
                setSpendOpen(true);
                return;
            }
            if (raw.startsWith("/skills")) {
                setActivityInspectorOpen(true);
                setActivityTab("overview");
                return;
            }
            if (raw.startsWith("/tasks")) {
                setActivityInspectorOpen(true);
                setActivityTab("overview");
                return;
            }
            if (raw.startsWith("/terminal")) {
                setActivityInspectorOpen(true);
                setActivityTab("commands");
                return;
            }
            if (raw.startsWith("/share")) {
                const transcript = messages
                    .filter((m) => m.role === "user" || m.role === "assistant")
                    .map((m) => `**${m.role === "user" ? "User" : "LetsGo"}**:\n${m.content}`)
                    .join("\n\n---\n\n");
                navigator.clipboard.writeText(transcript);
                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: "✓ Conversación copiada al portapapeles en formato Markdown."}]);
                return;
            }
            if (raw.startsWith("/review")) {
                setActivityInspectorOpen(true);
                setActivityTab("git");
                refreshGitDiff();
                setBusy(true);
                setStreaming(true);
                setStatusLabel(t(lang, "status.thinking"));
                ChatService.Send("Por favor revisa detalladamente los cambios pendientes en el repositorio (git diff), identificando posibles bugs, mejoras de diseño y consistencia con las reglas del proyecto.").catch((err) => {
                    setBusy(false);
                    setStreaming(false);
                    applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✖ Error al solicitar revisión: ${err}`}]);
                });
                return;
            }
            if (raw.startsWith("/init")) {
                setBusy(true);
                setStreaming(true);
                setStatusLabel(t(lang, "status.thinking"));
                ChatService.Send("Inicializa la configuración del proyecto generando un archivo AGENTS.md completo con arquitectura, comandos de instalación/compilación/test/lint y reglas de estilo.").catch((err) => {
                    setBusy(false);
                    setStreaming(false);
                    applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✖ Error al inicializar proyecto: ${err}`}]);
                });
                return;
            }
            if (raw.startsWith("/undo")) {
                const lastAssist = [...messages].reverse().find((m) => m.role === "assistant" && !m.kind);
                if (lastAssist) {
                    handleRollback(lastAssist);
                    applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: "✓ Último turno revertido."}]);
                } else {
                    applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: "No hay turnos previos para deshacer."}]);
                }
                return;
            }
            if (raw.startsWith("/redo")) {
                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: "No hay turnos deshechos para rehacer."}]);
                return;
            }
            if (raw.startsWith("/compact")) {
                if (messages.length > 4) {
                    const keepMsgs = messages.slice(-4);
                    setTabs((prev) =>
                        prev.map((tb) => (tb.id === activeTabId ? {...tb, messages: [
                            {id: nextId(), role: "assistant", content: "*(Contexto anterior compactado)*"},
                            ...keepMsgs
                        ]} : tb))
                    );
                    applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: "✓ Contexto compactado exitosamente."}]);
                } else {
                    applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: "La conversación ya es compacta."}]);
                }
                return;
            }
            applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: t(lang, "cmd.help")}]);
            return;
        }
        let targetSid = activeTab?.sessionId;
        if (!targetSid) {
            try {
                const currentProject = status?.project_dir || "";
                const dateStr = new Date().toISOString().slice(0, 10);
                const base = currentProject ? basename(currentProject) : "General";
                const sessionName = `[${dateStr}] ${base}`;
                const newSess = await SessionsService.Create(sessionName, currentProject);
                targetSid = (newSess as any)?.id || (newSess as any)?.ID || null;
                if (targetSid) {
                    setTabs((prev) => prev.map((tb) => tb.id === activeTabId ? {...tb, sessionId: targetSid, title: sessionName} : tb));
                    const list = await SessionsService.List();
                    if (Array.isArray(list)) {
                        setSessions(list.map((s: any) => ({
                            id: s.id || s.ID,
                            title: s.name || s.Name || "Sesión",
                            created_at: s.created_at || s.CreatedAt || "",
                            updated_at: s.updated_at || s.UpdatedAt || "",
                            message_count: s.message_count || s.MessageCount || 0,
                            project_path: s.project_path || s.ProjectPath || "",
                        })));
                    }
                }
            } catch (e) {
                console.error("Failed to auto-create session in send:", e);
            }
        }
        setBusy(true);
        setStreaming(true);
        setStatusLabel(t(lang, "status.thinking"));
        setErrorState(false);
        setLastError(null);
        try {
            if ((ChatService as any).SendInSession && targetSid) {
                await (ChatService as any).SendInSession(text, targetSid);
            } else {
                await ChatService.Send(text);
            }
        } catch (err: any) {
            setBusy(false);
            setStreaming(false);
            setErrorState(true);
            setLastError(String(err));
            applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✖ ${err}`}]);
        }
    }, [input, busy, lang, activeTabId, activeTab, applyMessages, status]);

    const onFilesPicked = useCallback(
        async (e: React.ChangeEvent<HTMLInputElement>) => {
            const files = Array.from(e.target.files || []);
            e.target.value = "";
            if (files.length === 0 || busy) return;
            const loaded = await Promise.all(
                files.map(async (f): Promise<Attachment> => {
                    let content = "";
                    try {
                        content = await f.text();
                    } catch {}
                    return {id: nextId(), name: f.name, size: f.size, content};
                }),
            );
            setAttachments((prev) => [...prev, ...loaded]);
        },
        [busy],
    );

    const removeAttachment = useCallback((id: number) => {
        setAttachments((prev) => prev.filter((a) => a.id !== id));
    }, []);

    // ─── Comandos de la paleta del compositor (filtrado) ────
    const isModelSubmenu = input.startsWith("/model ") || input === "/model" || input.startsWith("/models");
    const filteredModelSuggestions = useMemo(() => {
        if (!isModelSubmenu) return [];
        const q = input.startsWith("/model ") ? input.slice(7).trim().toLowerCase() : "";
        if (!q) return headerModels.slice(0, 10);
        return headerModels.filter((m) => m.toLowerCase().includes(q)).slice(0, 10);
    }, [isModelSubmenu, input, headerModels]);

    const filteredCmds = useMemo(() => {
        if (isModelSubmenu) return [];
        return COMMANDS.filter((c) => c.name.startsWith(input.trim().toLowerCase()));
    }, [input, isModelSubmenu]);

    const paletteOpenComposer = input.startsWith("/") && !paletteDismissed;

    const changeWorkMode = useCallback((mode: string) => {
        setStatus((s) => s ? {...s, work_mode: mode} : s);
        setModeMenuOpen(false);
    }, []);

    const onKeyDown = useCallback(
        (e: React.KeyboardEvent) => {
            if (paletteOpenComposer) {
                if (isModelSubmenu && filteredModelSuggestions.length > 0) {
                    if (e.key === "ArrowDown") {
                        e.preventDefault();
                        setCmdSel((s) => (s + 1) % filteredModelSuggestions.length);
                        return;
                    }
                    if (e.key === "ArrowUp") {
                        e.preventDefault();
                        setCmdSel((s) => (s - 1 + filteredModelSuggestions.length) % filteredModelSuggestions.length);
                        return;
                    }
                    if (e.key === "Tab" || (e.key === "Enter" && !e.shiftKey)) {
                        e.preventDefault();
                        const chosen = filteredModelSuggestions[cmdSel] || filteredModelSuggestions[0];
                        if (chosen) {
                            changeHeaderModel(chosen);
                            setInput("");
                            applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✓ Modelo activo cambiado a: \`${chosen}\``}]);
                            return;
                        }
                    }
                } else if (filteredCmds.length > 0) {
                    if (e.key === "ArrowDown") {
                        e.preventDefault();
                        setCmdSel((s) => (s + 1) % filteredCmds.length);
                        return;
                    }
                    if (e.key === "ArrowUp") {
                        e.preventDefault();
                        setCmdSel((s) => (s - 1 + filteredCmds.length) % filteredCmds.length);
                        return;
                    }
                    if (e.key === "Tab") {
                        e.preventDefault();
                        const chosen = filteredCmds[cmdSel] || filteredCmds[0];
                        if (chosen) {
                            setInput(chosen.name + (chosen.hasArgs ? " " : ""));
                            setCmdSel(0);
                            return;
                        }
                    }
                }
            } else {
                const isMention = /@\S*$/.test(input);
                if (e.key === "Tab" && !isMention) {
                    e.preventDefault();
                    const modes = WORK_MODES.map((m) => m.id);
                    const currentModeId = status?.work_mode ?? "planning";
                    const currentIdx = modes.indexOf(currentModeId as any);
                    const validIdx = currentIdx >= 0 ? currentIdx : 0;
                    const nextIdx = e.shiftKey
                        ? (validIdx - 1 + modes.length) % modes.length
                        : (validIdx + 1) % modes.length;
                    changeWorkMode(modes[nextIdx]);
                    return;
                }
            }
            if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                send();
            }
        },
        [send, paletteOpenComposer, isModelSubmenu, filteredModelSuggestions, filteredCmds, cmdSel, changeHeaderModel, applyMessages, input, status?.work_mode, changeWorkMode],
    );

    // ─── Paleta de comandos global (Ctrl+K) ──────────────────
    const paletteItems = useMemo(() => {
        const q = paletteQuery.trim().toLowerCase();
        const items: {key: string; group: string; label: string; hint: string; run: () => void}[] = [
            {
                key: "new-chat",
                group: "pal.group_actions",
                label: t(lang, "pal.new_chat"),
                hint: "",
                run: addTab,
            },
            {
                key: "clear-chat",
                group: "pal.group_actions",
                label: t(lang, "pal.clear"),
                hint: "",
                run: () => {
                    setTabs((prev) =>
                        prev.map((tb) => (tb.id === activeTabId ? {...tb, messages: [], tools: [], streaming: null} : tb)),
                    );
                    setActivePane("none");
                },
            },
            {
                key: "theme-toggle",
                group: "pal.group_actions",
                label: (status?.theme) === "light" ? t(lang, "pal.theme_dark") : t(lang, "pal.theme_light"),
                hint: "",
                run: () => setStatus(s => s ? {...s, theme: s.theme === "light" ? "letsgo" : "light"} : s),
            },
            {
                key: "settings",
                group: "pal.group_actions",
                label: t(lang, "pal.settings"),
                hint: "",
                run: () => setSettingsOpen(true),
            },
            {
                key: "onboarding",
                group: "pal.group_actions",
                label: t(lang, "nav.onboarding"),
                hint: "",
                run: () => setOnboardingOpen(true),
            },
            {
                key: "toggle-sidebar",
                group: "pal.group_actions",
                label: sidebarCollapsed ? t(lang, "pal.show_sidebar") : t(lang, "pal.toggle_sidebar"),
                hint: "",
                run: () => setSidebarCollapsed((v) => !v),
            },
            {
                key: "open-project",
                group: "pal.group_projects",
                label: t(lang, "pal.open_project"),
                hint: "",
                run: () => {},
            },
            {
                key: "new-project",
                group: "pal.group_projects",
                label: t(lang, "pal.new_project"),
                hint: "",
                run: () => {},
            },
            {
                key: "tab-web",
                group: "pal.group_tabs",
                label: t(lang, "browser.tab"),
                hint: "",
                run: () => setActivePane("browser"),
            },
            {
                key: "tab-test",
                group: "pal.group_tabs",
                label: t(lang, "test.tab"),
                hint: "",
                run: () => setActivePane("test"),
            },
            {
                key: "tab-skills",
                group: "pal.group_tabs",
                label: t(lang, "skills.tab"),
                hint: "",
                run: () => setActivePane("skills"),
            },
        ];
        if (!q) return items;
        return items.filter((it) => it.label.toLowerCase().includes(q));
    }, [lang, qpaletteDeps(), status?.theme, sidebarCollapsed, addTab, activeTabId]);

    function qpaletteDeps() {
        return 0;
    }

    const changeBrowserAccount = useCallback((name: string) => {
        setBrowser((b) => ({...b, status: {...(b.status ?? {provider: "", model: "", account: ""}), account: name}}));
        setBrowserModelMenu(false);
    }, []);

    const changeBrowserModel = useCallback((model: string) => {
        setBrowser((b) => ({...b, status: {...(b.status ?? {provider: "", model: "", account: ""}), model}}));
        setBrowserModelMenu(false);
    }, []);

    const sendBrowserPrompt = useCallback(() => {
        const text = browserInput.trim();
        if (!text) return;
        setBrowserInput("");
        setBrowser((prev) => ({...prev, busy: true, messages: [...prev.messages, {id: nextId(), role: "user", content: text}]}));
        // TODO: await SendBrowser(text);
    }, [browserInput]);

    const sendTestPrompt = useCallback(() => {
        const text = testInput.trim();
        if (!text) return;
        setTestInput("");
        setTest((prev) => ({...prev, busy: true, messages: [...prev.messages, {id: nextId(), role: "user", content: text}]}));
        // TODO: await SendTest(text);
    }, [testInput]);

    const sendSkillsPrompt = useCallback(() => {
        const text = skillsInput.trim();
        if (!text) return;
        setSkillsInput("");
        setSkills((prev) => ({...prev, busy: true, messages: [...prev.messages, {id: nextId(), role: "user", content: text}]}));
        // TODO: await SendSkills(text);
    }, [skillsInput]);

    const cancel = useCallback(async () => {
        try {
            await ChatService.Cancel();
        } catch {}
        setBusy(false);
        setStreaming(false);
        setStatusLabel(t(lang, "status.cancelled"));
    }, [lang]);

    const copyMessage = useCallback(async (content: string) => {
        try {
            await navigator.clipboard.writeText(content);
        } catch {}
    }, []);

    const handleCopied = useCallback((id: number) => {
        setCopiedId(id);
        setTimeout(() => setCopiedId(null), 1500);
    }, []);

    const retryMessage = useCallback(
        async (m: ChatMessage) => {
            if (m.role !== "user" || busy) return;
            applyMessages((prev) => [...prev, m]);
            setBusy(true);
            setStreaming(true);
            setStatusLabel(t(lang, "status.thinking"));
            try {
                await ChatService.Send(m.content);
            } catch (err: any) {
                setBusy(false);
                setStreaming(false);
                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✖ ${err}`}]);
            }
        },
        [busy, applyMessages, lang],
    );

    const openSavedSession = useCallback(
        async (sessionId: string) => {
            try {
                const history = await SessionsService.Open(sessionId);
                const found = sessions.find((s) => s.id === sessionId);
                const title = found?.title || "Sesión";
                const chatMsgs: ChatMessage[] = parseStoredHistory(history);
                const existing = tabs.find((t) => t.sessionId === sessionId);
                if (existing) {
                    setTabs((prev) => prev.map((t) => t.id === existing.id ? {...t, messages: chatMsgs} : t));
                    setActiveTabId(existing.id);
                    setActivePane("none");
                    return;
                }
                const id = nextTabId();
                setTabs((prev) => [...prev, {
                    id,
                    title,
                    messages: chatMsgs,
                    tools: [],
                    streaming: null,
                    saved: true,
                    sessionId,
                }]);
                setActiveTabId(id);
                setActivePane("none");
                if (found?.project_path) {
                    setStatus((s) => s ? ({...s, project_dir: found.project_path!}) : s);
                }
            } catch (e) {
                console.error("Failed to open session:", e);
            }
        },
        [sessions, tabs],
    );

    const deleteSavedSession = useCallback(
        async (sessionId: string) => {
            try {
                await SessionsService.Delete(sessionId);
            } catch (e) {
                console.error("Failed to delete session:", e);
            }
            setSessions((prev) => prev.filter((s) => s.id !== sessionId));
            setTabs((prev) => prev.filter((t) => t.sessionId !== sessionId));
        },
        [],
    );

    const renameSessionById = useCallback(
        async (sessionId: string, currentTitle: string) => {
            const newName = window.prompt("Nuevo nombre para la conversación:", currentTitle);
            if (!newName || newName.trim() === "" || newName === currentTitle) return;
            const trimmed = newName.trim();
            try {
                await SessionsService.Rename(sessionId, trimmed);
            } catch (e) {
                console.error("Failed to rename session:", e);
            }
            setSessions((prev) => prev.map((s) => s.id === sessionId ? {...s, title: trimmed} : s));
            setTabs((prev) => prev.map((t) => t.sessionId === sessionId ? {...t, title: trimmed} : t));
        },
        [],
    );

    const handleOpenProject = useCallback(async () => {
        setProjectMenuOpen(false);
        try {
            const dir = await SessionsService.SelectProjectFolder();
            if (dir) {
                setStatus((s) => s ? ({...s, project_dir: dir}) : s);
                const name = basename(dir);
                setRecentProjects((prev) => {
                    const filtered = prev.filter((p) => p.path !== dir);
                    return [{name, path: dir, last_opened: new Date().toISOString()}, ...filtered].slice(0, 10);
                });
                const dateStr = new Date().toISOString().slice(0, 10);
                const sessionName = `[${dateStr}] ${name}`;
                const newSess = await SessionsService.Create(sessionName, dir);
                const sessId = (newSess as any)?.id || (newSess as any)?.ID;
                if (sessId) {
                    const id = nextTabId();
                    setTabs((prev) => [...prev, {
                        ...newTab(id),
                        title: sessionName,
                        sessionId: sessId,
                    }]);
                    setActiveTabId(id);
                    setActivePane("none");
                }
                const list = await SessionsService.List();
                if (Array.isArray(list)) {
                    setSessions(list.map((s: any) => ({
                        id: s.id || s.ID,
                        title: s.name || s.Name || "Sesión",
                        created_at: s.created_at || s.CreatedAt || "",
                        updated_at: s.updated_at || s.UpdatedAt || "",
                        message_count: s.message_count || s.MessageCount || 0,
                        project_path: s.project_path || s.ProjectPath || "",
                    })));
                }
            }
        } catch (err) {
            console.error("SelectProjectFolder error:", err);
        }
    }, []);

    const [collapsedFolders, setCollapsedFolders] = useState<Record<string, boolean>>({});

    const toggleFolder = useCallback((folderKey: string) => {
        setCollapsedFolders((prev) => ({...prev, [folderKey]: !prev[folderKey]}));
    }, []);

    const createSessionForProject = useCallback(async (projectPath: string) => {
        try {
            const dateStr = new Date().toISOString().slice(0, 10);
            const base = projectPath ? basename(projectPath) : "General";
            const sessionName = `[${dateStr}] ${base}`;
            const newSess = await SessionsService.Create(sessionName, projectPath);
            const sessId = (newSess as any)?.id || (newSess as any)?.ID;
            if (sessId) {
                const id = nextTabId();
                tabTokensRef.current.set(id, 0);
                setSessionUsage((u) => ({...u, tokens: 0}));
                setTabs((prev) => [...prev, {
                    ...newTab(id),
                    title: sessionName,
                    sessionId: sessId,
                }]);
                setActiveTabId(id);
                setActivePane("none");
                if (projectPath) {
                    setStatus((s) => s ? ({...s, project_dir: projectPath}) : s);
                    setRecentProjects((prev) => {
                        const filtered = prev.filter((p) => p.path !== projectPath);
                        return [{name: base, path: projectPath, last_opened: new Date().toISOString()}, ...filtered].slice(0, 10);
                    });
                }
            }
            const list = await SessionsService.List();
            if (Array.isArray(list)) {
                setSessions(list.map((s: any) => ({
                    id: s.id || s.ID,
                    title: s.name || s.Name || "Sesión",
                    created_at: s.created_at || s.CreatedAt || "",
                    updated_at: s.updated_at || s.UpdatedAt || "",
                    message_count: s.message_count || s.MessageCount || 0,
                    project_path: s.project_path || s.ProjectPath || "",
                })));
            }
        } catch (err) {
            console.error("createSessionForProject error:", err);
        }
    }, []);

    const groupedSessions = useMemo(() => {
        const map = new Map<string, {name: string; path: string; items: GUISessionInfo[]}>();
        if (status?.project_dir) {
            const p = status.project_dir;
            map.set(p, {
                name: basename(p),
                path: p,
                items: [],
            });
        }
        for (const rp of recentProjects) {
            if (rp.path && !map.has(rp.path)) {
                map.set(rp.path, {
                    name: rp.name || basename(rp.path),
                    path: rp.path,
                    items: [],
                });
            }
        }
        for (const s of sessions) {
            const p = s.project_path || "";
            const key = p || "__general__";
            if (!map.has(key)) {
                map.set(key, {
                    name: p ? basename(p) : "General",
                    path: p,
                    items: [],
                });
            }
            map.get(key)!.items.push(s);
        }
        return Array.from(map.entries());
    }, [sessions, recentProjects, status?.project_dir]);

    const removeRecent = useCallback((path: string) => {
        setRecentProjects((prev) => prev.filter((rp) => rp.path !== path));
    }, []);

    const applyZoom = useCallback((next: number) => {
        const clamped = Math.max(70, Math.min(150, Math.round(next)));
        if (clamped === zoomRef.current) return;
        zoomRef.current = clamped;
        setZoom(clamped);
    }, []);

    const onSettingsSaved = useCallback(() => {}, []);

    // ─── Render ──────────────────────────────────────────────
    if (!status || !config) {
        return (
            <div id="app" style={appZoomStyle}>
                <div className="loading-screen" style={{display: "flex", alignItems: "center", justifyContent: "center", height: "100vh"}}>
                    <div className="spinner" />
                </div>
            </div>
        );
    }

    return (
        <div id="app" style={appZoomStyle}>
            <header id="app-header">
                <button type="button"
                    className={`icon-btn${sidebarCollapsed ? " pressed" : ""}`}
                    onClick={() => setSidebarCollapsed((v) => !v)}
                    aria-pressed={!sidebarCollapsed}
                    title={sidebarCollapsed ? t(lang, "pal.show_sidebar") : t(lang, "pal.toggle_sidebar")}
                >
                    <MenuIcon />
                </button>
                <div className="brand">
                    <img className="brand-logo" src={goulmLogo} alt="LetsGo" />
                    LetsGo
                </div>
                <button type="button"
                    className="project-trigger"
                    onClick={() => setProjectMenuOpen((v) => !v)}
                    title={status?.project_dir || t(lang, "app.empty_project")}
                >
                    <FolderIcon />
                    <span className="project-name">
                        {status?.project_dir ? basename(status.project_dir) : t(lang, "app.empty_project")}
                    </span>
                </button>
                {projectMenuOpen && (
                    <div id="project-menu">
                        {recentProjects.length > 0 && (
                            <>
                                <div className="sidebar-section-title">{t(lang, "app.recent")}</div>
                                {recentProjects.map((rp) => (
                                    <button type="button"
                                        key={rp.path}
                                        className="project-menu-item"
                                        onClick={async () => {
                                            setProjectMenuOpen(false);
                                            setStatus((s) => s ? ({...s, project_dir: rp.path}) : s);
                                            try {
                                                await SessionsService.SetCurrentProject(rp.path);
                                            } catch {}
                                        }}
                                    >
                                        <FolderIcon />
                                        <span className="item-title">{rp.name}</span>
                                    </button>
                                ))}
                            </>
                        )}
                        <div className="project-menu-actions">
                            <button type="button" className="btn small" onClick={handleOpenProject}>
                                {t(lang, "app.open_project")}
                            </button>
                            <button type="button" className="btn small" onClick={handleOpenProject}>
                                {t(lang, "app.create_project")}
                            </button>
                        </div>
                    </div>
                )}
                <div className="account-select">
                    {config?.accounts && config.accounts.length > 0 ? (
                        <>
                            <button type="button"
                                className="account-select-trigger"
                                onClick={() => setHeaderAccountMenu((v) => !v)}
                                title={t(lang, "app.account")}
                            >
                                <span className="account-select-name">{status?.active_account || "…"}</span>
                                <span className="mode-caret">▾</span>
                            </button>
                            {headerAccountMenu && (
                                <div id="header-account-menu">
                                    {config.accounts.map((acc) => {
                                        const active = acc.name === status?.active_account;
                                        return (
                                            <button type="button"
                                                key={acc.name}
                                                className={`browser-acc-item ${active ? "active" : ""}`}
                                                onClick={() => changeHeaderAccount(acc.name)}
                                            >
                                                <span className="browser-acc-item-name">{acc.name}</span>
                                                <span className="browser-acc-item-sub">{acc.provider} · {acc.model}</span>
                                            </button>
                                        );
                                    })}
                                </div>
                            )}
                        </>
                    ) : (
                        <div className="account-select-static">
                            <span className="account-select-name">{status?.active_account || "…"}</span>
                        </div>
                    )}
                </div>
                <div className="usage-chip" role="button" tabIndex={0}
                    title={t(lang, "spend.title")}
                    onClick={() => setSpendOpen(true)}
                    onKeyDown={(e) => { if (e.key === "Enter" || e.key === " ") { e.preventDefault(); setSpendOpen(true); } }}
                >
                    {t(lang, "usage.totals", {tokens: sessionUsage.tokens, cost: sessionUsage.cost.toFixed(4)})}
                </div>
                <button type="button"
                    className={`icon-btn ${explorerOpen ? "active" : ""}`}
                    onClick={openExplorer}
                    title={t(lang, "explorer.title")}
                >
                    <FolderIcon />
                </button>
                <button type="button"
                    className={`icon-btn ${activityInspectorOpen && activityTab === "git" ? "active" : ""}`}
                    onClick={() => {
                        if (activityInspectorOpen && activityTab === "git") {
                            setActivityInspectorOpen(false);
                        } else {
                            setActivityInspectorOpen(true);
                            setActivityTab("git");
                        }
                    }}
                    title="Git Review (Ctrl+D)"
                >
                    <GitIcon />
                </button>
                <button type="button"
                    className={`icon-btn ${activityInspectorOpen && activityTab === "overview" ? "active" : ""}`}
                    onClick={() => {
                        if (activityInspectorOpen && activityTab === "overview") {
                            setActivityInspectorOpen(false);
                        } else {
                            setActivityInspectorOpen(true);
                            setActivityTab("overview");
                        }
                    }}
                    title="Activity Inspector (Ctrl+B)"
                >
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <rect width="7" height="9" x="3" y="3" rx="1"/>
                        <rect width="7" height="5" x="14" y="3" rx="1"/>
                        <rect width="7" height="9" x="14" y="12" rx="1"/>
                        <rect width="7" height="5" x="3" y="16" rx="1"/>
                    </svg>
                </button>
                <button type="button"
                    className="palette-btn"
                    onClick={() => setPaletteOpen(true)}
                    title={t(lang, "app.palette")}
                >
                    <span>Ctrl K</span>
                </button>
            </header>

            <div id="app-main">
                <aside id="sidebar" className={sidebarCollapsed ? "collapsed" : ""}>
                    <div className="sidebar-section">
                        <div className="sidebar-section-title">{t(lang, "app.recent")}</div>
                        {recentProjects.length === 0 && (
                            <div className="sidebar-empty">{t(lang, "app.no_recent")}</div>
                        )}
                        {recentProjects.map((rp) => (
                            <div key={rp.path} className="sidebar-item" title={rp.path}>
                                <button type="button"
                                    className="sidebar-item-open"
                                    onClick={async () => {
                                        setStatus((s) => s ? ({...s, project_dir: rp.path}) : s);
                                        try {
                                            await SessionsService.SetCurrentProject(rp.path);
                                        } catch {}
                                    }}
                                    title={rp.path}
                                >
                                    <FolderIcon />
                                    <span className="item-title">{rp.name}</span>
                                </button>
                                <button type="button"
                                    className="item-delete"
                                    onClick={() => removeRecent(rp.path)}
                                    title={t(lang, "app.remove_recent")}
                                >
                                    ✕
                                </button>
                            </div>
                        ))}
                        {recentProjects.length > 0 && (
                            <button type="button"
                                className="btn small"
                                style={{width: "100%", marginTop: 6}}
                                onClick={() => setRecentProjects([])}
                            >
                                {t(lang, "app.clear_recent")}
                            </button>
                        )}
                    </div>
                    <div className="sidebar-section" style={{flex: 1, display: "flex", flexDirection: "column", minHeight: 0}}>
                        <button
                            type="button"
                            className="sidebar-new-chat-btn"
                            onClick={addTab}
                            title={t(lang, "pal.new_chat")}
                        >
                            <PlusIcon size={14} />
                            <span>{t(lang, "pal.new_chat")}</span>
                        </button>
                        <div className="sidebar-section-title">{t(lang, "app.tabs")}</div>
                        <div className="sidebar-list">
                            {sessions.length === 0 && (
                                <div className="sidebar-empty">{t(lang, "app.no_sessions")}</div>
                            )}
                            {groupedSessions.map(([key, group]) => {
                                const isCollapsed = !!collapsedFolders[key];
                                return (
                                    <div key={key} className="folder-group">
                                        <div className="folder-group-header-row">
                                            <button
                                                type="button"
                                                className="folder-group-header"
                                                onClick={async () => {
                                                    toggleFolder(key);
                                                    if (group.path) {
                                                        setStatus((s) => s ? ({...s, project_dir: group.path}) : s);
                                                        try {
                                                            await SessionsService.SetCurrentProject(group.path);
                                                        } catch {}
                                                    }
                                                }}
                                                title={group.path || "General"}
                                            >
                                                <span className="folder-chevron">{isCollapsed ? "▶" : "▼"}</span>
                                                <FolderIcon />
                                                <span className="folder-title">{group.name}</span>
                                                <span className="folder-count">{group.items.length}</span>
                                            </button>
                                            <button
                                                type="button"
                                                className="folder-add-btn"
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    createSessionForProject(group.path);
                                                }}
                                                title={t(lang, "app.new_chat")}
                                            >
                                                +
                                            </button>
                                        </div>
                                        {!isCollapsed && (
                                            <div className="folder-group-items">
                                                {group.items.length === 0 ? (
                                                    <div className="folder-empty-hint">Sin conversaciones aún</div>
                                                ) : (
                                                    group.items.map((s) => (
                                                        <div
                                                            key={s.id}
                                                            className={`sidebar-item ${activeTab?.sessionId === s.id ? "active" : ""}`}
                                                        >
                                                            <button
                                                                type="button"
                                                                className="sidebar-item-open"
                                                                onClick={() => openSavedSession(s.id)}
                                                                onDoubleClick={() => renameSessionById(s.id, s.title)}
                                                                title={s.title}
                                                            >
                                                                <span className="item-title">
                                                                    {s.title || t(lang, "app.new_chat")}
                                                                    <span className="last-access-badge" title="Último acceso">
                                                                        {formatLastAccess(s.updated_at || s.created_at)}
                                                                    </span>
                                                                </span>
                                                                {s.kind === "test" && <span className="session-badge test">{t(lang, "test.tab")}</span>}
                                                                {s.kind === "skills" && <span className="session-badge skills">{t(lang, "skills.tab")}</span>}
                                                            </button>
                                                            <button
                                                                type="button"
                                                                className="item-rename"
                                                                onClick={() => renameSessionById(s.id, s.title)}
                                                                title="Renombrar conversación"
                                                            >
                                                                ✎
                                                            </button>
                                                            <button
                                                                type="button"
                                                                className="item-delete"
                                                                onClick={() => deleteSavedSession(s.id)}
                                                                title={t(lang, "app.close_tab")}
                                                            >
                                                                ✕
                                                            </button>
                                                        </div>
                                                    ))
                                                )}
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                    <div id="sidebar-footer">
                        <button
                            type="button"
                            className="sidebar-footer-btn"
                            onClick={() => setSettingsOpen(true)}
                            title={t(lang, "app.config")}
                        >
                            <GearIcon />
                            <span>{t(lang, "app.config")}</span>
                        </button>
                    </div>
                </aside>

                <main id="chat-pane">
                    <div id="tabs-bar">
                        {tabs.map((tb) => (
                            <div
                                key={tb.id}
                                className={`tab-item ${tb.id === activeTabId && !specialActive ? "active" : ""}`}
                            >
                                <button type="button"
                                    className="tab-open"
                                    role="tab"
                                    aria-selected={tb.id === activeTabId && !specialActive}
                                    onClick={() => selectChatTab(tb.id)}
                                    onDoubleClick={() => {
                                        if (tb.sessionId) {
                                            renameSessionById(tb.sessionId, tb.title);
                                        } else {
                                            const n = window.prompt("Renombrar pestaña:", tb.title);
                                            if (n && n.trim()) {
                                                setTabs((prev) => prev.map((t) => t.id === tb.id ? {...t, title: n.trim()} : t));
                                            }
                                        }
                                    }}
                                >
                                    <span className="tab-title">{tb.title || t(lang, "app.new_chat")}</span>
                                </button>
                                {tb.id === activeTabId && !specialActive && (busy || streaming) && (
                                    <span className="tab-dot" title={t(lang, "status.busy")} />
                                )}
                                <button type="button"
                                    className="tab-close"
                                    onClick={(e) => { e.stopPropagation(); closeTab(tb.id); }}
                                    title={t(lang, "app.close_tab")}
                                >
                                    ✕
                                </button>
                            </div>
                        ))}
                        <span className="tabs-bar-spacer" />
                        <div className={`tab-item web-tab ${browserActive ? "active" : ""}`} role="tab" aria-selected={browserActive}>
                            <button type="button" className="tab-open web-tab-open" onClick={selectWebTab} title={t(lang, "browser.tab")}>
                                <span className="web-tab-icon"><GlobeIcon /></span>
                                <span className="tab-title">{t(lang, "browser.tab")}</span>
                            </button>
                            {browserActive && (
                                <button type="button" className="tab-close" onClick={() => setActivePane("none")} title={t(lang, "app.close_tab")}>
                                    ✕
                                </button>
                            )}
                        </div>
                        <div className={`tab-item special-tab test-tab ${testActive ? "active" : ""}`} role="tab" aria-selected={testActive}>
                            <button type="button" className="tab-open special-tab-open" onClick={selectTestTab} title={t(lang, "test.tab")}>
                                <span className="special-tab-icon"><FlaskIcon /></span>
                                <span className="tab-title">{t(lang, "test.tab")}</span>
                            </button>
                            {testActive && (
                                <button type="button" className="tab-close" onClick={() => setActivePane("none")} title={t(lang, "app.close_tab")}>
                                    ✕
                                </button>
                            )}
                        </div>
                        <div className={`tab-item special-tab skills-tab ${skillsActive ? "active" : ""}`} role="tab" aria-selected={skillsActive}>
                            <button type="button" className="tab-open special-tab-open" onClick={selectSkillsTab} title={t(lang, "skills.tab")}>
                                <span className="special-tab-icon"><SkillIcon /></span>
                                <span className="tab-title">{t(lang, "skills.tab")}</span>
                            </button>
                            {skillsActive && (
                                <button type="button" className="tab-close" onClick={() => setActivePane("none")} title={t(lang, "app.close_tab")}>
                                    ✕
                                </button>
                            )}
                        </div>
                    </div>

                    <div id="chat-viewport" ref={viewportRef} onScroll={onViewportScroll}>
                        {errorState && lastError && (
                            <div className="error-banner" role="alert">
                                <span><strong>{t(lang, "msg.error_title")}:</strong> {lastError}</span>
                                <button type="button" onClick={() => setErrorState(false)} title={t(lang, "msg.close")}>
                                    ✕
                                </button>
                            </div>
                        )}
                        {browserActive && (
                            <div className="browser-bar">
                                <span className="browser-bar-title"><GlobeIcon /> {t(lang, "browser.tab")}</span>
                                <div className="browser-bar-controls">
                                    <div className="browser-model-select">
                                        <button type="button"
                                            className="browser-model-trigger"
                                            onClick={() => setBrowserModelMenu((v) => !v)}
                                            title={t(lang, "browser.model")}
                                        >
                                            <span className="browser-model-provider">{browser.status?.provider || t(lang, "browser.none")}</span>
                                            <span className="browser-model-name">{browser.status?.model || t(lang, "browser.none")}</span>
                                            <span className="mode-caret">▾</span>
                                        </button>
                                        {browserModelMenu && (
                                            <div id="browser-model-menu">
                                                <div className="browser-menu-group">{t(lang, "browser.account")}</div>
                                                {(config?.accounts ?? []).map((a) => {
                                                    const active = a.name === browser.status?.account;
                                                    return (
                                                        <button type="button"
                                                            key={a.name}
                                                            className={`browser-acc-item ${active ? "active" : ""}`}
                                                            onClick={() => changeBrowserAccount(a.name)}
                                                        >
                                                            <span className="browser-acc-item-name">{a.name}</span>
                                                            <span className="browser-acc-item-model">{a.provider} / {a.model}</span>
                                                        </button>
                                                    );
                                                })}
                                                <div className="browser-menu-divider" />
                                                <div className="browser-menu-group">{t(lang, "browser.model")}</div>
                                                {["anthropic/claude-3-5-sonnet", "anthropic/claude-3-5-haiku", "openai/gpt-4o-mini"].map((m) => {
                                                    const active = m === browser.status?.model;
                                                    return (
                                                        <button type="button"
                                                            key={m}
                                                            className={`browser-acc-item ${active ? "active" : ""}`}
                                                            onClick={() => changeBrowserModel(m)}
                                                        >
                                                            <span className="browser-acc-item-name">{m}</span>
                                                        </button>
                                                    );
                                                })}
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        )}
                        {testActive ? (
                            <SpecialPaneView
                                kind="test"
                                pane={test}
                                lang={lang}
                                copiedId={copiedId}
                                onCopyContent={copyMessage}
                                onCopied={handleCopied}
                                onRetry={retryMessage}
                            />
                        ) : skillsActive ? (
                            <SpecialPaneView
                                kind="skills"
                                pane={skills}
                                lang={lang}
                                copiedId={copiedId}
                                onCopyContent={copyMessage}
                                onCopied={handleCopied}
                                onRetry={retryMessage}
                            />
                        ) : browserActive ? (
                            <>
                                {browser.messages.length === 0 && (
                                    <div id="chat-empty" className="browser-empty">
                                        <div className="browser-empty-icon"><GlobeIcon /></div>
                                        <h1>{t(lang, "browser.welcome_title")}</h1>
                                        <p>{t(lang, "browser.welcome_subtitle")}</p>
                                        {browserHistory.length > 0 && (
                                            <div className="browser-history">
                                                <div className="browser-history-title">{t(lang, "browser.history")}</div>
                                                {browserHistory.slice(0, 6).map((h) => (
                                                    <button type="button"
                                                        key={h.query}
                                                        className="browser-history-item"
                                                        onClick={() => { setBrowserInput(h.query); composerRef.current?.focus(); }}
                                                        title={h.query}
                                                    >
                                                        <span className="browser-history-icon"><GlobeIcon /></span>
                                                        <span className="item-title">{h.query}</span>
                                                    </button>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                )}
                                {browser.messages.map((m) => (
                                    <MessageRow
                                        key={m.id}
                                        m={m}
                                        lang={lang}
                                        isStreaming={false}
                                        copied={copiedId === m.id}
                                        onCopyContent={copyMessage}
                                        onCopied={handleCopied}
                                        onRetry={retryMessage}
                                        onRollback={handleRollback}
                                    />
                                ))}
                                {browser.busy && !browser.streaming && !browser.tools.some((tl) => tl.status === "running") && (
                                    <div className="thinking">
                                        <span className="thinking-dots"><i /><i /><i /></span>
                                        <span>{t(lang, "status.thinking_bubble")}</span>
                                    </div>
                                )}
                            </>
                        ) : (
                            <>
                                {!status?.project_dir && messages.length > 0 && (
                                    <div className="unscoped-project-banner">
                                        <div className="banner-text">
                                            <FolderIcon />
                                            <span>{t(lang, "app.no_project_banner")}</span>
                                        </div>
                                        <button type="button" className="btn small primary" onClick={handleOpenProject}>
                                            {t(lang, "app.open_project")}
                                        </button>
                                    </div>
                                )}
                                {messages.length === 0 && (
                                    <div id="chat-empty">
                                        <h1>{t(lang, "app.welcome_title")}</h1>
                                        <p>
                                            {!status?.project_dir
                                                ? t(lang, "app.welcome_no_project")
                                                : t(lang, "app.welcome_subtitle")}
                                        </p>
                                        <div className="welcome-actions">
                                            <button type="button" className="btn primary" onClick={handleOpenProject}>
                                                {t(lang, "app.open_project")}
                                            </button>
                                            {recentProjects.length > 0 && (
                                                <button type="button" className="btn" onClick={() => setProjectMenuOpen(true)}>
                                                    {t(lang, "app.recent")}
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                )}
                                {messages.map((m) => (
                                    <MessageRow
                                        key={m.id}
                                        m={m}
                                        lang={lang}
                                        isStreaming={streaming && activeTab?.streaming?.msgId === m.id}
                                        copied={copiedId === m.id}
                                        onCopyContent={copyMessage}
                                        onCopied={handleCopied}
                                        onRetry={retryMessage}
                                        onRollback={handleRollback}
                                    />
                                ))}
                                {busy && !streaming && !tools.some((tl) => tl.status === "running") && (
                                    <div className="thinking">
                                        <span className="thinking-dots"><i /><i /><i /></span>
                                        <span>{t(lang, "status.thinking_bubble")}</span>
                                    </div>
                                )}
                            </>
                        )}
                    </div>

                    <div className="disclaimer">{t(lang, "app.disclaimer")}</div>

                    {showJumpDown && (
                        <button type="button" className="jump-down" onClick={jumpToBottom} title={t(lang, "msg.close")}>
                            <ArrowDownIcon />
                        </button>
                    )}

                    <div id="composer">
                        {!specialActive && paletteOpenComposer && (isModelSubmenu ? filteredModelSuggestions.length > 0 : filteredCmds.length > 0) && (
                            <div id="cmd-palette" role="listbox">
                                <div className="cmd-palette-title">
                                    <span>{isModelSubmenu ? `Modelos disponibles (${status?.provider || "activo"})` : t(lang, "composer.commands")}</span>
                                    <span className="cmd-palette-hint">Tab ⇥ o Enter ↵</span>
                                </div>
                                {isModelSubmenu ? (
                                    filteredModelSuggestions.map((m, i) => (
                                        <button type="button"
                                            key={m}
                                            className={`cmd-palette-item ${i === cmdSel ? "selected" : ""}`}
                                            onMouseEnter={() => setCmdSel(i)}
                                            onClick={() => {
                                                changeHeaderModel(m);
                                                setInput("");
                                                applyMessages((prev) => [...prev, {id: nextId(), role: "assistant", content: `✓ Modelo activo cambiado a: \`${m}\``}]);
                                                composerRef.current?.focus();
                                            }}
                                        >
                                            <div className="cmd-palette-left">
                                                <ProviderIcon provider={status?.provider} model={m} size={15} />
                                                <span className="cmd-palette-name">{m}</span>
                                            </div>
                                            <span className="cmd-palette-desc">Seleccionar modelo</span>
                                        </button>
                                    ))
                                ) : (
                                    filteredCmds.map((c, i) => (
                                        <button type="button"
                                            key={c.name}
                                            className={`cmd-palette-item ${i === cmdSel ? "selected" : ""}`}
                                            onMouseEnter={() => setCmdSel(i)}
                                            onClick={() => { setInput(c.name + (c.hasArgs ? " " : "")); setCmdSel(0); composerRef.current?.focus(); }}
                                        >
                                            <div className="cmd-palette-left">
                                                <span className="cmd-palette-icon">/</span>
                                                <span className="cmd-palette-name">{c.name.slice(1)}</span>
                                                {c.args && <span className="cmd-palette-args">{c.args}</span>}
                                            </div>
                                            <span className="cmd-palette-desc">{t(lang, c.descKey)}</span>
                                        </button>
                                    ))
                                )}
                            </div>
                        )}
                        {!specialActive && attachments.length > 0 && (
                            <div id="attachment-list">
                                {attachments.map((a) => (
                                    <div key={a.id} className="attachment-chip">
                                        <span className="att-name" title={a.name}>{a.name}</span>
                                        <span className="att-size">{fmtSize(a.size)}</span>
                                        <button type="button"
                                            className="att-remove"
                                            onClick={() => removeAttachment(a.id)}
                                            title={t(lang, "composer.remove")}
                                        >
                                            ✕
                                        </button>
                                    </div>
                                ))}
                            </div>
                        )}
                        <div id="composer-box">
                            {!specialActive && (
                                <div className="composer-modes-header">
                                    <div className="composer-modes-bar" title="Usa Tab para alternar modo">
                                        {WORK_MODES.map((m) => {
                                            const active = m.id === (status?.work_mode ?? "planning");
                                            const Icon = m.icon;
                                            return (
                                                <div
                                                    key={m.id}
                                                    className={`composer-mode-tab ${active ? "active" : ""}`}
                                                    onClick={() => changeWorkMode(m.id)}
                                                    role="tab"
                                                    aria-selected={active}
                                                >
                                                    <Icon />
                                                    <span>{m.label}</span>
                                                </div>
                                            );
                                        })}
                                    </div>
                                    <div className="composer-modes-shortcut">
                                        <span>Tab ⇥</span>
                                    </div>
                                </div>
                            )}
                            {browserActive ? (
                                <textarea
                                    ref={composerRef}
                                    className="browser-input"
                                    value={browserInput}
                                    onChange={(e) => setBrowserInput(e.target.value)}
                                    onKeyDown={(e) => {
                                        if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); sendBrowserPrompt(); }
                                    }}
                                    placeholder={t(lang, "browser.composer_ph")}
                                    rows={2}
                                />
                            ) : testActive ? (
                                <textarea
                                    ref={composerRef}
                                    className="browser-input"
                                    value={testInput}
                                    onChange={(e) => setTestInput(e.target.value)}
                                    onKeyDown={(e) => {
                                        if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); sendTestPrompt(); }
                                    }}
                                    placeholder={t(lang, "test.composer_ph")}
                                    rows={2}
                                />
                            ) : skillsActive ? (
                                <textarea
                                    ref={composerRef}
                                    className="browser-input"
                                    value={skillsInput}
                                    onChange={(e) => setSkillsInput(e.target.value)}
                                    onKeyDown={(e) => {
                                        if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); sendSkillsPrompt(); }
                                    }}
                                    placeholder={t(lang, "skills.composer_ph")}
                                    rows={2}
                                />
                            ) : (
                                <textarea
                                    ref={composerRef}
                                    value={input}
                                    onChange={(e) => { setInput(e.target.value); setPaletteDismissed(false); }}
                                    onKeyDown={onKeyDown}
                                    placeholder={t(lang, "composer.placeholder")}
                                    rows={2}
                                />
                            )}
                            <div className="composer-toolbar">
                                <div className="composer-toolbar-left">
                                    {!specialActive && (
                                        <button
                                            type="button"
                                            className="composer-icon-btn attach-btn"
                                            onClick={() => fileRef.current?.click()}
                                            title={t(lang, "composer.attach")}
                                            disabled={busy}
                                        >
                                            <PlusIcon size={16} />
                                        </button>
                                    )}

                                    <div className="composer-model-wrapper" style={{position: "relative"}}>
                                        <button
                                            type="button"
                                            className="composer-model-btn"
                                            onClick={() => {
                                                setHeaderModelMenu((v) => !v);
                                                if (headerModels.length === 0) {
                                                    refreshModels(status?.provider);
                                                }
                                            }}
                                            title={status ? `${status.provider} / ${status.model}` : "Modelo"}
                                        >
                                            <ProviderIcon provider={status?.provider} model={status?.model} size={15} />
                                            <span className="composer-model-label">
                                                {status?.model || "Modelo"}
                                            </span>
                                            <span className="mode-caret">▾</span>
                                        </button>
                                        {headerModelMenu && (
                                            <div className="composer-pill-menu" role="listbox">
                                                {headerModels.length === 0 ? (
                                                    <div style={{padding: "8px 12px", fontSize: "12px", color: "var(--fg-muted)"}}>
                                                        Cargando modelos...
                                                    </div>
                                                ) : (
                                                    headerModels.map((m) => (
                                                        <button
                                                            type="button"
                                                            key={m}
                                                            className={`browser-acc-item ${m === status?.model ? "active" : ""}`}
                                                            onClick={() => changeHeaderModel(m)}
                                                        >
                                                            <ProviderIcon provider={status?.provider} model={m} size={14} />
                                                            <span className="browser-acc-item-name">{m}</span>
                                                        </button>
                                                    ))
                                                )}
                                            </div>
                                        )}
                                    </div>
                                </div>

                                <div className="composer-toolbar-right">
                                    {browserActive ? (
                                        browser.busy ? (
                                            <button type="button" className="composer-icon-btn stop-btn" onClick={() => setBrowser((b) => ({...b, busy: false}))} title={t(lang, "composer.stop")}>
                                                <StopIcon />
                                            </button>
                                        ) : (
                                            <button type="button" className="composer-icon-btn send-btn" onClick={sendBrowserPrompt} disabled={!browserInput.trim()} title={t(lang, "composer.send")}>
                                                <ArrowUpIcon size={16} />
                                            </button>
                                        )
                                    ) : testActive ? (
                                        test.busy ? (
                                            <button type="button" className="composer-icon-btn stop-btn" onClick={() => setTest((p) => ({...p, busy: false}))} title={t(lang, "composer.stop")}>
                                                <StopIcon />
                                            </button>
                                        ) : (
                                            <button type="button" className="composer-icon-btn send-btn" onClick={sendTestPrompt} disabled={!testInput.trim()} title={t(lang, "composer.send")}>
                                                <ArrowUpIcon size={16} />
                                            </button>
                                        )
                                    ) : skillsActive ? (
                                        skills.busy ? (
                                            <button type="button" className="composer-icon-btn stop-btn" onClick={() => setSkills((p) => ({...p, busy: false}))} title={t(lang, "composer.stop")}>
                                                <StopIcon />
                                            </button>
                                        ) : (
                                            <button type="button" className="composer-icon-btn send-btn" onClick={sendSkillsPrompt} disabled={!skillsInput.trim()} title={t(lang, "composer.send")}>
                                                <ArrowUpIcon size={16} />
                                            </button>
                                        )
                                    ) : busy ? (
                                        <button type="button" className="composer-icon-btn stop-btn" onClick={cancel} title={t(lang, "composer.stop")}>
                                            <StopIcon />
                                        </button>
                                    ) : (
                                        <button
                                            type="button"
                                            className="composer-icon-btn send-btn"
                                            onClick={send}
                                            disabled={!input.trim() && attachments.length === 0}
                                            title={t(lang, "composer.send")}
                                        >
                                            <ArrowUpIcon size={16} />
                                        </button>
                                    )}
                                </div>
                            </div>
                        </div>
                        <input ref={fileRef} type="file" multiple hidden onChange={onFilesPicked} />
                        <div id="composer-hint">
                            <span>
                                {t(lang, "composer.hint")}
                                {attachments.length > 0 ? ` · ${t(lang, "composer.files", {n: attachments.length})}` : ""}
                            </span>
                        </div>
                    </div>
                </main>

                {explorerOpen && (
                    <aside id="explorer-pane" style={{width: 240, borderLeft: "1px solid var(--border)", background: "var(--bg-panel)", overflowY: "auto"}}>
                        <div className="explorer-panel">
                            <div className="pane-title">
                                {t(lang, "explorer.title")}
                                {explorerPath !== "" && (
                                    <button type="button" className="explorer-up" onClick={explorerGoUp} title={t(lang, "explorer.up")}>
                                        ←
                                    </button>
                                )}
                            </div>
                            {explorerEntries.length === 0 && <div className="pane-empty">{t(lang, "explorer.empty")}</div>}
                            {explorerEntries.map((entry) => (
                                <button type="button"
                                    key={entry.path}
                                    className={`explorer-item ${entry.is_dir ? "dir" : "file"}`}
                                    onClick={() => explorerOpenDir(entry)}
                                    title={entry.path}
                                >
                                    {entry.is_dir ? <FolderIcon /> : <span className="explorer-file-dot" />}
                                    <span className="explorer-item-name">{entry.name}</span>
                                </button>
                            ))}
                        </div>
                    </aside>
                )}

                <ActivityInspector
                    open={activityInspectorOpen}
                    activeTab={activityTab}
                    onTabChange={setActivityTab}
                    maximized={activityMaximized}
                    onToggleMaximize={() => setActivityMaximized((v) => !v)}
                    onClose={() => setActivityInspectorOpen(false)}
                    gitSummary={gitSummary}
                    selectedFile={gitSelectedFile}
                    fileDiff={gitFileDiff}
                    gitLoading={gitLoading}
                    onSelectFile={selectGitFile}
                    onStageFile={handleStageFile}
                    onUnstageFile={handleUnstageFile}
                    onStageAll={handleStageAll}
                    onUnstageAll={handleUnstageAll}
                    onCommit={handleCommit}
                    onRefreshGit={refreshGitDiff}
                    commandLogs={commandLogs}
                    backgroundTasks={backgroundTasks}
                    skillsUsed={skillsUsed}
                    lang={lang}
                />
            </div>

            {paletteOpen && (
                <div id="palette-overlay" onClick={() => setPaletteOpen(false)}>
                    <div id="palette" onClick={(e) => e.stopPropagation()}>
                        <input
                            ref={paletteRef}
                            id="palette-input"
                            value={paletteQuery}
                            onChange={(e) => setPaletteQuery(e.target.value)}
                            onKeyDown={(e) => {
                                if (e.key === "Escape") { e.preventDefault(); setPaletteOpen(false); return; }
                                if (e.key === "ArrowDown") { e.preventDefault(); setPaletteSel((s) => (s + 1) % Math.max(1, paletteItems.length)); return; }
                                if (e.key === "ArrowUp") { e.preventDefault(); setPaletteSel((s) => (s - 1 + paletteItems.length) % Math.max(1, paletteItems.length)); return; }
                                if (e.key === "Enter") {
                                    const it = paletteItems[Math.min(paletteSel, paletteItems.length - 1)];
                                    if (it) { e.preventDefault(); setPaletteOpen(false); setPaletteQuery(""); it.run(); }
                                }
                            }}
                            placeholder={t(lang, "app.palette_hint")}
                        />
                        <div id="palette-groups">
                            {paletteItems.length === 0 && (
                                <div className="palette-empty">{t(lang, "app.no_sessions")}</div>
                            )}
                            {(() => {
                                let lastGroup = "";
                                return paletteItems.map((it, i) => {
                                    const showGroup = it.group !== lastGroup;
                                    lastGroup = it.group;
                                    return (
                                        <div key={it.key}>
                                            {showGroup && <div className="palette-group">{t(lang, it.group)}</div>}
                                            <button type="button"
                                                className={`palette-item ${i === paletteSel ? "selected" : ""}`}
                                                onMouseEnter={() => setPaletteSel(i)}
                                                onClick={() => { setPaletteOpen(false); setPaletteQuery(""); it.run(); }}
                                            >
                                                <span className="palette-icon">
                                                    {it.key.startsWith("proj-") || it.key.startsWith("open-project") || it.key.startsWith("new-project") ? (
                                                        <FolderIcon />
                                                    ) : it.key.startsWith("tab-") ? (
                                                        <span style={{fontSize: 11}}>◧</span>
                                                    ) : (
                                                        <span style={{fontSize: 11}}>›</span>
                                                    )}
                                                </span>
                                                <span className="item-title">{it.label}</span>
                                                {it.hint && <span className="palette-desc">{it.hint}</span>}
                                            </button>
                                        </div>
                                    );
                                });
                            })()}
                        </div>
                    </div>
                </div>
            )}

            {settingsOpen && (
                <ErrorBoundary lang={lang}>
                    <SettingsView
                        lang={lang}
                        zoom={zoom}
                        initialConfig={config}
                        onApplyTheme={(th) => {
                            localStorage.setItem("app_theme", th);
                            setStatus((s) => s ? ({...s, theme: th}) : s);
                            document.documentElement.dataset.theme = themeFor(th);
                            ThemeService.Set(th).catch(() => {});
                        }}
                        onApplyZoom={applyZoom}
                        onClose={() => setSettingsOpen(false)}
                        onSaved={() => {
                            SettingsService.GetConfig().then((c: any) => {
                                if (c) {
                                    setConfig((prev) => ({...prev, ...c}));
                                    if (c.model) setStatus((s) => ({...s, model: c.model}));
                                    if (c.language) setStatus((s) => ({...s, language: c.language}));
                                    const th = c.theme || c.ThemeVariant;
                                    if (th) {
                                        localStorage.setItem("app_theme", th);
                                        setStatus((s) => ({...s, theme: th}));
                                        document.documentElement.dataset.theme = themeFor(th);
                                    }
                                }
                            });
                        }}
                    />
                </ErrorBoundary>
            )}

            {onboardingOpen && (
                <OnboardingView
                    lang={lang}
                    onLangChange={(l) => {
                        setStatus((s) => ({...s, language: l}));
                        SettingsService.SaveConfig({language: l}).catch(() => {});
                    }}
                    onComplete={(updated) => {
                        setConfig((prev) => ({...prev, ...updated}));
                        if (updated.language) setStatus((s) => ({...s, language: updated.language as string}));
                        if (updated.model) setStatus((s) => ({...s, model: updated.model as string}));
                        setOnboardingOpen(false);
                    }}
                    onClose={() => setOnboardingOpen(false)}
                />
            )}

            {spendOpen && (
                <SpendView lang={lang} onClose={() => setSpendOpen(false)} />
            )}

            {pendingToolApproval && (
                <div className="permission-modal-overlay" style={{
                    position: "fixed",
                    inset: 0,
                    backgroundColor: "rgba(10, 10, 15, 0.8)",
                    backdropFilter: "blur(6px)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    zIndex: 9999,
                    padding: "20px",
                }}>
                    <div className="permission-modal-card" style={{
                        backgroundColor: "var(--card, #1e1e2e)",
                        borderRadius: "14px",
                        border: "1px solid var(--border, #313244)",
                        boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
                        width: "100%",
                        maxWidth: "520px",
                        overflow: "hidden",
                    }}>
                        <div style={{padding: "20px 24px", borderBottom: "1px solid #313244", display: "flex", alignItems: "center", gap: "10px"}}>
                            <span style={{fontSize: "1.4rem"}}>🛡️</span>
                            <h3 style={{margin: 0, fontSize: "1.15rem", color: "#cdd6f4"}}>
                                Autorizar ejecución de herramienta
                            </h3>
                        </div>
                        <div style={{padding: "20px 24px"}}>
                            <div style={{marginBottom: "12px", display: "flex", alignItems: "center", gap: "8px"}}>
                                <span style={{fontSize: "0.85rem", color: "#a6adc8"}}>Herramienta:</span>
                                <code style={{backgroundColor: "#11111b", padding: "2px 8px", borderRadius: "4px", color: "#89b4fa", fontWeight: 600}}>
                                    {pendingToolApproval.name}
                                </code>
                            </div>
                            {pendingToolApproval.input && (
                                <div style={{backgroundColor: "#11111b", padding: "12px", borderRadius: "8px", border: "1px solid #313244", maxHeight: "160px", overflowY: "auto", fontFamily: "monospace", fontSize: "0.82rem", color: "#cdd6f4", marginBottom: "16px"}}>
                                    {typeof pendingToolApproval.input === "string" ? pendingToolApproval.input : JSON.stringify(pendingToolApproval.input, null, 2)}
                                </div>
                            )}
                            <p style={{fontSize: "0.85rem", color: "#a6adc8", margin: "0 0 16px"}}>
                                Elige el nivel de autorización (estilo Gemini CLI):
                            </p>
                            <div style={{display: "flex", flexDirection: "column", gap: "8px"}}>
                                <button
                                    type="button"
                                    onClick={() => {
                                        ChatService.ApproveScope(pendingToolApproval.callId, "once", pendingToolApproval.name);
                                        setPendingToolApproval(null);
                                    }}
                                    style={{
                                        padding: "10px 16px",
                                        borderRadius: "8px",
                                        border: "1px solid #89b4fa",
                                        backgroundColor: "rgba(137, 180, 250, 0.15)",
                                        color: "#89b4fa",
                                        fontWeight: 600,
                                        cursor: "pointer",
                                        textAlign: "left",
                                    }}
                                >
                                    ✔ Permitir solo esta vez
                                </button>
                                <button
                                    type="button"
                                    onClick={() => {
                                        ChatService.ApproveScope(pendingToolApproval.callId, "chat", pendingToolApproval.name);
                                        setPendingToolApproval(null);
                                    }}
                                    style={{
                                        padding: "10px 16px",
                                        borderRadius: "8px",
                                        border: "1px solid #a6e3a1",
                                        backgroundColor: "rgba(166, 227, 161, 0.15)",
                                        color: "#a6e3a1",
                                        fontWeight: 600,
                                        cursor: "pointer",
                                        textAlign: "left",
                                    }}
                                >
                                    💬 Permitir siempre para este chat
                                </button>
                                <button
                                    type="button"
                                    onClick={() => {
                                        ChatService.ApproveScope(pendingToolApproval.callId, "project", pendingToolApproval.name);
                                        setPendingToolApproval(null);
                                    }}
                                    style={{
                                        padding: "10px 16px",
                                        borderRadius: "8px",
                                        border: "1px solid #cba6f7",
                                        backgroundColor: "rgba(203, 166, 247, 0.15)",
                                        color: "#cba6f7",
                                        fontWeight: 600,
                                        cursor: "pointer",
                                        textAlign: "left",
                                    }}
                                >
                                    📁 Permitir siempre para este proyecto
                                </button>
                                <button
                                    type="button"
                                    onClick={() => {
                                        ChatService.ApproveScope(pendingToolApproval.callId, "deny", pendingToolApproval.name);
                                        setPendingToolApproval(null);
                                    }}
                                    style={{
                                        padding: "10px 16px",
                                        borderRadius: "8px",
                                        border: "1px solid #f38ba8",
                                        backgroundColor: "rgba(243, 139, 168, 0.15)",
                                        color: "#f38ba8",
                                        fontWeight: 600,
                                        cursor: "pointer",
                                        textAlign: "left",
                                    }}
                                >
                                    ✖ Denegar
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

export default App;