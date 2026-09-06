import {useEffect, useState} from "react";
import {t, type Lang} from "./i18n";
import {SettingsService, ThemeService} from "./bindings";
import type {ConfigView, SettingsTab} from "./types";

interface Props {
    lang: Lang;
    zoom: number;
    initialConfig?: ConfigView;
    onApplyTheme?: (next: string) => void;
    onApplyZoom: (next: number) => void;
    onClose: () => void;
    onSaved: () => void;
}

const TABS: {id: SettingsTab; labelKey: string; icon: string}[] = [
    {id: "providers", labelKey: "tab.providers", icon: "⚡"},
    {id: "general", labelKey: "tab.general", icon: "⚙️"},
    {id: "safety", labelKey: "tab.safety", icon: "🛡️"},
    {id: "budget", labelKey: "tab.budget", icon: "📊"},
];

function maskKey(key?: string): string {
    if (!key) return "";
    const trimmed = key.trim();
    if (trimmed.length <= 8) return "••••••••";
    return trimmed.slice(0, 4) + "••••••••" + trimmed.slice(-4);
}

export default function SettingsView({lang, zoom, initialConfig, onApplyTheme, onApplyZoom, onClose, onSaved}: Props) {
    const [tab, setTab] = useState<SettingsTab>("providers");
    const [cfg, setCfg] = useState<ConfigView>(() => initialConfig ?? {
        theme: "letsgo",
        language: lang,
        auto_approve: { "file-edit": false, "bash": false, "web": true },
        max_tokens: 4096,
        temperature: 0.7,
        ollama_base_url: "http://localhost:11434",
    });
    const [showKey, setShowKey] = useState<Record<string, boolean>>({});
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<{text: string; isError?: boolean} | null>(null);

function normalizeCfg(data: any): ConfigView {
    if (!data) return {} as ConfigView;
    return {
        ...data,
        anthropic_api_key: data.anthropic_api_key || data.AnthropicAPIKey || "",
        openai_api_key: data.openai_api_key || data.OpenAIAPIKey || "",
        groq_api_key: data.groq_api_key || data.GroqAPIKey || "",
        openrouter_api_key: data.openrouter_api_key || data.OpenRouterAPIKey || "",
        gemini_api_key: data.gemini_api_key || data.GeminiAPIKey || "",
        deepseek_api_key: data.deepseek_api_key || data.DeepSeekAPIKey || "",
        ollama_base_url: data.ollama_base_url || data.OllamaBaseURL || "http://localhost:11434",
        model: data.model || data.Model || "",
        plan_model: data.plan_model || data.PlanModel || "",
        theme: ((data.theme === "goulm" || data.theme === "lets-go") ? "letsgo" : data.theme) || ((data.ThemeVariant === "goulm" || data.ThemeVariant === "lets-go") ? "letsgo" : data.ThemeVariant) || "letsgo",
        language: data.language || "es",
        auto_approve: data.auto_approve || data.AutoApprove || {},
        max_tokens: data.max_tokens || data.MaxTokens || 4096,
        temperature: data.temperature ?? data.Temperature ?? 0.7,
    };
}

    // Cargar config real de Go si no se pasó o para refrescar
    useEffect(() => {
        SettingsService.GetConfig().then((data: any) => {
            if (data) {
                setCfg((prev) => ({...prev, ...normalizeCfg(data)}));
            }
        }).catch(() => {});
    }, []);

    const flash = (text: string, isError = false) => {
        setMessage({text, isError});
        setTimeout(() => setMessage(null), 3000);
    };

    const toggleShowKey = (provider: string) => {
        setShowKey((prev) => ({...prev, [provider]: !prev[provider]}));
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            const payload: any = {
                ...cfg,
                AnthropicAPIKey: cfg.anthropic_api_key,
                OpenAIAPIKey: cfg.openai_api_key,
                GroqAPIKey: cfg.groq_api_key,
                OpenRouterAPIKey: cfg.openrouter_api_key,
                GeminiAPIKey: cfg.gemini_api_key,
                DeepSeekAPIKey: cfg.deepseek_api_key,
                OllamaBaseURL: cfg.ollama_base_url,
                Model: cfg.model,
                PlanModel: cfg.plan_model,
                ThemeVariant: cfg.theme,
                theme: cfg.theme,
                Language: cfg.language,
                language: cfg.language,
            };
            if (cfg.theme) {
                localStorage.setItem("app_theme", cfg.theme);
                ThemeService.Set(cfg.theme).catch(() => {});
            }
            const updated = await SettingsService.SaveConfig(payload);
            if (updated) {
                setCfg((prev) => ({...prev, ...normalizeCfg(updated)}));
            }
            setSaving(false);
            flash(t(lang, "providers.saved"));
            onSaved();
        } catch (err) {
            setSaving(false);
            flash(String(err), true);
        }
    };

    return (
        <div id="settings-overlay" data-theme={(cfg.theme === "goulm" || cfg.theme === "lets-go") ? "letsgo" : (cfg.theme || "letsgo")} style={{
            position: "fixed",
            inset: 0,
            backgroundColor: "rgba(10, 10, 15, 0.8)",
            backdropFilter: "blur(6px)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 9000,
            padding: "20px",
        }}>
            <div id="settings-modal" style={{
                backgroundColor: "var(--bg-panel)",
                borderRadius: "16px",
                border: "1px solid var(--border)",
                boxShadow: "0 25px 50px -12px var(--shadow)",
                width: "100%",
                maxWidth: "850px",
                maxHeight: "90vh",
                display: "flex",
                flexDirection: "column",
                overflow: "hidden",
            }}>
                {/* Header */}
                <div id="settings-header" style={{
                    padding: "20px 24px",
                    borderBottom: "1px solid var(--border)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                }}>
                    <div style={{display: "flex", alignItems: "center", gap: "12px"}}>
                        <h2 style={{margin: 0, fontSize: "1.3rem", fontWeight: 700, color: "var(--text)"}}>
                            {t(lang, "app.config")}
                        </h2>
                        <span style={{
                            fontSize: "0.75rem",
                            padding: "4px 10px",
                            borderRadius: "12px",
                            backgroundColor: "rgba(166, 227, 161, 0.15)",
                            color: "#a6e3a1",
                            fontWeight: 600,
                        }}>
                            🔒 AES-256 Encrypted
                        </span>
                    </div>
                    <button
                        type="button"
                        className="btn icon-btn"
                        onClick={onClose}
                        style={{
                            background: "transparent",
                            border: "none",
                            color: "var(--text-soft)",
                            fontSize: "1.2rem",
                            cursor: "pointer",
                            padding: "4px 8px",
                            borderRadius: "6px",
                        }}
                        title={t(lang, "settings.close")}
                    >
                        ✕
                    </button>
                </div>

                {/* Tabs Navigation */}
                <div id="settings-tabs" style={{
                    display: "flex",
                    gap: "8px",
                    padding: "12px 24px",
                    backgroundColor: "var(--bg)",
                    borderBottom: "1px solid var(--border)",
                }}>
                    {TABS.map((tb) => (
                        <button
                            type="button"
                            key={tb.id}
                            onClick={() => setTab(tb.id)}
                            style={{
                                padding: "8px 16px",
                                borderRadius: "8px",
                                border: "none",
                                backgroundColor: tab === tb.id ? "var(--accent)" : "transparent",
                                color: tab === tb.id ? "var(--on-accent)" : "var(--text-soft)",
                                fontWeight: tab === tb.id ? 700 : 500,
                                cursor: "pointer",
                                display: "flex",
                                alignItems: "center",
                                gap: "6px",
                                fontSize: "0.9rem",
                            }}
                        >
                            <span>{tb.icon}</span>
                            <span>{t(lang, tb.labelKey)}</span>
                        </button>
                    ))}
                </div>

                {/* Flash message */}
                {message && (
                    <div style={{
                        padding: "10px 24px",
                        backgroundColor: message.isError ? "rgba(243, 139, 168, 0.2)" : "rgba(166, 227, 161, 0.2)",
                        color: message.isError ? "#f38ba8" : "#a6e3a1",
                        fontSize: "0.88rem",
                        display: "flex",
                        alignItems: "center",
                        gap: "8px",
                    }}>
                        <span>{message.isError ? "✖" : "✔"}</span>
                        <span>{message.text}</span>
                    </div>
                )}

                {/* Tab Content */}
                <div id="settings-body" style={{
                    padding: "24px",
                    overflowY: "auto",
                    flex: 1,
                    display: "flex",
                    flexDirection: "column",
                    gap: "20px",
                }}>
                    {/* TAB: PROVEEDORES (P0) */}
                    {tab === "providers" && (
                        <div>
                            <div style={{marginBottom: "20px"}}>
                                <h3 style={{margin: "0 0 6px", fontSize: "1.1rem", color: "var(--text)"}}>
                                    {t(lang, "providers.title")}
                                </h3>
                                <p style={{margin: 0, fontSize: "0.85rem", color: "var(--text-soft)"}}>
                                    {t(lang, "providers.subtitle")}
                                </p>
                                <div style={{
                                    marginTop: "10px",
                                    padding: "8px 12px",
                                    backgroundColor: "rgba(137, 180, 250, 0.1)",
                                    borderRadius: "8px",
                                    border: "1px solid rgba(137, 180, 250, 0.2)",
                                    fontSize: "0.8rem",
                                    color: "#89b4fa",
                                    display: "flex",
                                    alignItems: "center",
                                    gap: "8px",
                                }}>
                                    <span>🔒</span>
                                    <span>{t(lang, "providers.encryption_notice")}</span>
                                </div>
                            </div>

                            <div style={{display: "flex", flexDirection: "column", gap: "16px"}}>
                                {/* OpenRouter */}
                                <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px"}}>
                                        <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                                            <span style={{fontSize: "1.2rem"}}>🌐</span>
                                            <strong style={{color: "var(--text)"}}>OpenRouter</strong>
                                            <span style={{fontSize: "0.75rem", color: cfg.openrouter_api_key ? "#a6e3a1" : "#6c7086"}}>
                                                {cfg.openrouter_api_key ? t(lang, "providers.configured_badge") : t(lang, "providers.not_configured")}
                                            </span>
                                        </div>
                                    </div>
                                    <p style={{margin: "0 0 10px", fontSize: "0.8rem", color: "var(--text-soft)"}}>
                                        {t(lang, "providers.openrouter_desc")}
                                    </p>
                                    <div style={{display: "flex", gap: "8px"}}>
                                        <input
                                            type={showKey.openrouter ? "text" : "password"}
                                            placeholder="sk-or-v1-..."
                                            value={cfg.openrouter_api_key || ""}
                                            onChange={(e) => setCfg({...cfg, openrouter_api_key: e.target.value})}
                                            style={{
                                                flex: 1,
                                                padding: "8px 12px",
                                                borderRadius: "6px",
                                                border: "1px solid var(--border)",
                                                backgroundColor: "var(--code-bg)",
                                                color: "var(--text)",
                                                fontFamily: "monospace",
                                                fontSize: "0.85rem",
                                            }}
                                        />
                                        <button
                                            type="button"
                                            onClick={() => toggleShowKey("openrouter")}
                                            style={{padding: "6px 12px", borderRadius: "6px", border: "1px solid var(--border)", backgroundColor: "var(--bg-elevated)", color: "var(--text)", cursor: "pointer"}}
                                        >
                                            {showKey.openrouter ? "🙈" : "👁️"}
                                        </button>
                                    </div>
                                </div>

                                {/* Anthropic */}
                                <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px"}}>
                                        <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                                            <span style={{fontSize: "1.2rem"}}>🟣</span>
                                            <strong style={{color: "var(--text)"}}>Anthropic</strong>
                                            <span style={{fontSize: "0.75rem", color: cfg.anthropic_api_key ? "#a6e3a1" : "#6c7086"}}>
                                                {cfg.anthropic_api_key ? t(lang, "providers.configured_badge") : t(lang, "providers.not_configured")}
                                            </span>
                                        </div>
                                    </div>
                                    <p style={{margin: "0 0 10px", fontSize: "0.8rem", color: "var(--text-soft)"}}>
                                        {t(lang, "providers.anthropic_desc")}
                                    </p>
                                    <div style={{display: "flex", gap: "8px"}}>
                                        <input
                                            type={showKey.anthropic ? "text" : "password"}
                                            placeholder="sk-ant-..."
                                            value={cfg.anthropic_api_key || ""}
                                            onChange={(e) => setCfg({...cfg, anthropic_api_key: e.target.value})}
                                            style={{
                                                flex: 1,
                                                padding: "8px 12px",
                                                borderRadius: "6px",
                                                border: "1px solid var(--border)",
                                                backgroundColor: "var(--code-bg)",
                                                color: "var(--text)",
                                                fontFamily: "monospace",
                                                fontSize: "0.85rem",
                                            }}
                                        />
                                        <button
                                            type="button"
                                            onClick={() => toggleShowKey("anthropic")}
                                            style={{padding: "6px 12px", borderRadius: "6px", border: "1px solid var(--border)", backgroundColor: "var(--bg-elevated)", color: "var(--text)", cursor: "pointer"}}
                                        >
                                            {showKey.anthropic ? "🙈" : "👁️"}
                                        </button>
                                    </div>
                                </div>

                                {/* OpenAI */}
                                <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px"}}>
                                        <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                                            <span style={{fontSize: "1.2rem"}}>🟢</span>
                                            <strong style={{color: "var(--text)"}}>OpenAI</strong>
                                            <span style={{fontSize: "0.75rem", color: cfg.openai_api_key ? "#a6e3a1" : "#6c7086"}}>
                                                {cfg.openai_api_key ? t(lang, "providers.configured_badge") : t(lang, "providers.not_configured")}
                                            </span>
                                        </div>
                                    </div>
                                    <p style={{margin: "0 0 10px", fontSize: "0.8rem", color: "var(--text-soft)"}}>
                                        {t(lang, "providers.openai_desc")}
                                    </p>
                                    <div style={{display: "flex", gap: "8px"}}>
                                        <input
                                            type={showKey.openai ? "text" : "password"}
                                            placeholder="sk-proj-..."
                                            value={cfg.openai_api_key || ""}
                                            onChange={(e) => setCfg({...cfg, openai_api_key: e.target.value})}
                                            style={{
                                                flex: 1,
                                                padding: "8px 12px",
                                                borderRadius: "6px",
                                                border: "1px solid var(--border)",
                                                backgroundColor: "var(--code-bg)",
                                                color: "var(--text)",
                                                fontFamily: "monospace",
                                                fontSize: "0.85rem",
                                            }}
                                        />
                                        <button
                                            type="button"
                                            onClick={() => toggleShowKey("openai")}
                                            style={{padding: "6px 12px", borderRadius: "6px", border: "1px solid var(--border)", backgroundColor: "var(--bg-elevated)", color: "var(--text)", cursor: "pointer"}}
                                        >
                                            {showKey.openai ? "🙈" : "👁️"}
                                        </button>
                                    </div>
                                </div>

                                {/* Groq */}
                                <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px"}}>
                                        <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                                            <span style={{fontSize: "1.2rem"}}>⚡</span>
                                            <strong style={{color: "var(--text)"}}>Groq</strong>
                                            <span style={{fontSize: "0.75rem", color: cfg.groq_api_key ? "#a6e3a1" : "#6c7086"}}>
                                                {cfg.groq_api_key ? t(lang, "providers.configured_badge") : t(lang, "providers.not_configured")}
                                            </span>
                                        </div>
                                    </div>
                                    <p style={{margin: "0 0 10px", fontSize: "0.8rem", color: "var(--text-soft)"}}>
                                        {t(lang, "providers.groq_desc")}
                                    </p>
                                    <div style={{display: "flex", gap: "8px"}}>
                                        <input
                                            type={showKey.groq ? "text" : "password"}
                                            placeholder="gsk_..."
                                            value={cfg.groq_api_key || ""}
                                            onChange={(e) => setCfg({...cfg, groq_api_key: e.target.value})}
                                            style={{
                                                flex: 1,
                                                padding: "8px 12px",
                                                borderRadius: "6px",
                                                border: "1px solid var(--border)",
                                                backgroundColor: "var(--code-bg)",
                                                color: "var(--text)",
                                                fontFamily: "monospace",
                                                fontSize: "0.85rem",
                                            }}
                                        />
                                        <button
                                            type="button"
                                            onClick={() => toggleShowKey("groq")}
                                            style={{padding: "6px 12px", borderRadius: "6px", border: "1px solid var(--border)", backgroundColor: "var(--bg-elevated)", color: "var(--text)", cursor: "pointer"}}
                                        >
                                            {showKey.groq ? "🙈" : "👁️"}
                                        </button>
                                    </div>
                                </div>

                                {/* Ollama (Local) */}
                                <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px"}}>
                                        <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                                            <span style={{fontSize: "1.2rem"}}>🚀</span>
                                            <strong style={{color: "var(--text)"}}>Ollama (Local)</strong>
                                            <span style={{fontSize: "0.75rem", color: "#a6e3a1"}}>
                                                Local / Privado
                                            </span>
                                        </div>
                                    </div>
                                    <p style={{margin: "0 0 10px", fontSize: "0.8rem", color: "var(--text-soft)"}}>
                                        {t(lang, "providers.ollama_desc")}
                                    </p>
                                    <div>
                                        <label style={{display: "block", fontSize: "0.8rem", color: "var(--text-soft)", marginBottom: "4px"}}>
                                            {t(lang, "providers.ollama_url")}
                                        </label>
                                        <input
                                            type="text"
                                            placeholder="http://localhost:11434"
                                            value={cfg.ollama_base_url || "http://localhost:11434"}
                                            onChange={(e) => setCfg({...cfg, ollama_base_url: e.target.value})}
                                            style={{
                                                width: "100%",
                                                padding: "8px 12px",
                                                borderRadius: "6px",
                                                border: "1px solid var(--border)",
                                                backgroundColor: "var(--code-bg)",
                                                color: "var(--text)",
                                                fontFamily: "monospace",
                                                fontSize: "0.85rem",
                                                boxSizing: "border-box",
                                            }}
                                        />
                                    </div>
                                </div>

                                {/* Google Gemini */}
                                <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px"}}>
                                        <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                                            <span style={{fontSize: "1.2rem"}}>🔵</span>
                                            <strong style={{color: "var(--text)"}}>Google Gemini</strong>
                                            <span style={{fontSize: "0.75rem", color: cfg.gemini_api_key ? "#a6e3a1" : "#6c7086"}}>
                                                {cfg.gemini_api_key ? t(lang, "providers.configured_badge") : t(lang, "providers.not_configured")}
                                            </span>
                                        </div>
                                    </div>
                                    <p style={{margin: "0 0 10px", fontSize: "0.8rem", color: "var(--text-soft)"}}>
                                        {t(lang, "providers.gemini_desc")}
                                    </p>
                                    <div style={{display: "flex", gap: "8px"}}>
                                        <input
                                            type={showKey.gemini ? "text" : "password"}
                                            placeholder="AIza..."
                                            value={cfg.gemini_api_key || ""}
                                            onChange={(e) => setCfg({...cfg, gemini_api_key: e.target.value})}
                                            style={{
                                                flex: 1,
                                                padding: "8px 12px",
                                                borderRadius: "6px",
                                                border: "1px solid var(--border)",
                                                backgroundColor: "var(--code-bg)",
                                                color: "var(--text)",
                                                fontFamily: "monospace",
                                                fontSize: "0.85rem",
                                            }}
                                        />
                                        <button
                                            type="button"
                                            onClick={() => toggleShowKey("gemini")}
                                            style={{padding: "6px 12px", borderRadius: "6px", border: "1px solid var(--border)", backgroundColor: "var(--bg-elevated)", color: "var(--text)", cursor: "pointer"}}
                                        >
                                            {showKey.gemini ? "🙈" : "👁️"}
                                        </button>
                                    </div>
                                </div>

                                {/* DeepSeek */}
                                <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "8px"}}>
                                        <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                                            <span style={{fontSize: "1.2rem"}}>🐋</span>
                                            <strong style={{color: "var(--text)"}}>DeepSeek</strong>
                                            <span style={{fontSize: "0.75rem", color: cfg.deepseek_api_key ? "#a6e3a1" : "#6c7086"}}>
                                                {cfg.deepseek_api_key ? t(lang, "providers.configured_badge") : t(lang, "providers.not_configured")}
                                            </span>
                                        </div>
                                    </div>
                                    <p style={{margin: "0 0 10px", fontSize: "0.8rem", color: "var(--text-soft)"}}>
                                        API Key para DeepSeek-V3 y DeepSeek-R1.
                                    </p>
                                    <div style={{display: "flex", gap: "8px"}}>
                                        <input
                                            type={showKey.deepseek ? "text" : "password"}
                                            placeholder="sk-..."
                                            value={cfg.deepseek_api_key || ""}
                                            onChange={(e) => setCfg({...cfg, deepseek_api_key: e.target.value})}
                                            style={{
                                                flex: 1,
                                                padding: "8px 12px",
                                                borderRadius: "6px",
                                                border: "1px solid var(--border)",
                                                backgroundColor: "var(--code-bg)",
                                                color: "var(--text)",
                                                fontFamily: "monospace",
                                                fontSize: "0.85rem",
                                            }}
                                        />
                                        <button
                                            type="button"
                                            onClick={() => toggleShowKey("deepseek")}
                                            style={{padding: "6px 12px", borderRadius: "6px", border: "1px solid var(--border)", backgroundColor: "var(--bg-elevated)", color: "var(--text)", cursor: "pointer"}}
                                        >
                                            {showKey.deepseek ? "🙈" : "👁️"}
                                        </button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* TAB: GENERAL */}
                    {tab === "general" && (
                        <div style={{display: "flex", flexDirection: "column", gap: "20px"}}>
                            <div>
                                <label style={{display: "block", fontSize: "0.9rem", fontWeight: 600, marginBottom: "8px", color: "var(--text)"}}>
                                    Idioma / Language
                                </label>
                                <div style={{display: "flex", gap: "12px"}}>
                                    <button
                                        type="button"
                                        onClick={() => setCfg({...cfg, language: "es"})}
                                        style={{
                                            padding: "10px 20px",
                                            borderRadius: "8px",
                                            border: (cfg.language || lang) === "es" ? "2px solid #89b4fa" : "1px solid #45475a",
                                            backgroundColor: (cfg.language || lang) === "es" ? "rgba(137, 180, 250, 0.15)" : "#181825",
                                            color: "var(--text)",
                                            cursor: "pointer",
                                        }}
                                    >
                                        🇪🇸 Español
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => setCfg({...cfg, language: "en"})}
                                        style={{
                                            padding: "10px 20px",
                                            borderRadius: "8px",
                                            border: (cfg.language || lang) === "en" ? "2px solid #89b4fa" : "1px solid #45475a",
                                            backgroundColor: (cfg.language || lang) === "en" ? "rgba(137, 180, 250, 0.15)" : "#181825",
                                            color: "var(--text)",
                                            cursor: "pointer",
                                        }}
                                    >
                                        🇺🇸 English
                                    </button>
                                </div>
                            </div>

                            <div>
                                <label style={{display: "block", fontSize: "0.9rem", fontWeight: 600, marginBottom: "8px", color: "var(--text)"}}>
                                    Tema / Theme
                                </label>
                                <div style={{display: "flex", gap: "12px"}}>
                                    {["letsgo", "dark", "light"].map((th) => (
                                        <button
                                            type="button"
                                            key={th}
                                            onClick={() => {
                                                setCfg({...cfg, theme: th});
                                                onApplyTheme?.(th);
                                                try {
                                                    localStorage.setItem("app_theme", th);
                                                    ThemeService.Set(th).catch(() => {});
                                                } catch {}
                                            }}
                                            style={{
                                                padding: "10px 20px",
                                                borderRadius: "8px",
                                                border: ((cfg.theme === "goulm" || cfg.theme === "lets-go") ? "letsgo" : (cfg.theme || "letsgo")) === th ? "2px solid var(--accent)" : "1px solid var(--border)",
                                                backgroundColor: ((cfg.theme === "goulm" || cfg.theme === "lets-go") ? "letsgo" : (cfg.theme || "letsgo")) === th ? "var(--accent-soft)" : "var(--bg-elevated)",
                                                color: "var(--text)",
                                                cursor: "pointer",
                                                textTransform: "capitalize",
                                            }}
                                        >
                                            {th === "letsgo" ? "LetsGo" : th}
                                        </button>
                                    ))}
                                </div>
                            </div>

                            <div>
                                <label style={{display: "block", fontSize: "0.9rem", fontWeight: 600, marginBottom: "8px", color: "var(--text)"}}>
                                    Zoom de la interfaz: {zoom}%
                                </label>
                                <div style={{display: "flex", gap: "10px"}}>
                                    {[80, 90, 100, 110, 125].map((z) => (
                                        <button
                                            type="button"
                                            key={z}
                                            onClick={() => onApplyZoom(z)}
                                            style={{
                                                padding: "6px 16px",
                                                borderRadius: "6px",
                                                border: zoom === z ? "2px solid var(--accent)" : "1px solid var(--border)",
                                                backgroundColor: zoom === z ? "var(--accent-soft)" : "var(--bg-elevated)",
                                                color: "var(--text)",
                                                cursor: "pointer",
                                            }}
                                        >
                                            {z}%
                                        </button>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    {/* TAB: SEGURIDAD Y PERMISOS */}
                    {tab === "safety" && (
                        <div style={{display: "flex", flexDirection: "column", gap: "16px"}}>
                            <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                <h4 style={{margin: "0 0 8px", color: "var(--text)"}}>Cifrado de Secretos en Disco</h4>
                                <p style={{margin: 0, fontSize: "0.85rem", color: "var(--text-soft)", lineHeight: 1.5}}>
                                    Todas las claves se almacenan automáticamente cifradas mediante <strong>AES-256-GCM</strong>.
                                    La clave de cifrado se deriva localmente en tu sistema operativo, protegiendo tus credenciales contra copias accidentales de archivos.
                                </p>
                            </div>

                            <div style={{backgroundColor: "var(--bg-elevated)", padding: "16px", borderRadius: "10px", border: "1px solid var(--border)"}}>
                                <h4 style={{margin: "0 0 12px", color: "var(--text)"}}>Auto-Aprobación de Herramientas</h4>
                                {[
                                    {id: "file-edit", label: "Edición de Archivos (write, edit, apply_patch)", desc: "Permite al modelo modificar archivos en el proyecto sin pedir confirmación previa"},
                                    {id: "bash", label: "Comandos Shell (terminal, bash, powershell)", desc: "Permite ejecutar comandos en la terminal automáticamente"},
                                    {id: "web", label: "Acceso Web (search, fetch)", desc: "Permite consultar páginas web y documentación"},
                                ].map((cat) => (
                                    <div key={cat.id} style={{display: "flex", justifyContent: "space-between", alignItems: "center", padding: "10px 0", borderBottom: "1px solid var(--border)"}}>
                                        <div>
                                            <strong style={{fontSize: "0.9rem", color: "var(--text)"}}>{cat.label}</strong>
                                            <p style={{margin: "2px 0 0", fontSize: "0.8rem", color: "var(--text-soft)"}}>{cat.desc}</p>
                                        </div>
                                        <input
                                            type="checkbox"
                                            checked={!!cfg.auto_approve?.[cat.id]}
                                            onChange={(e) => {
                                                const auto = {...(cfg.auto_approve || {})};
                                                auto[cat.id] = e.target.checked;
                                                setCfg({...cfg, auto_approve: auto});
                                            }}
                                            style={{cursor: "pointer", width: "18px", height: "18px"}}
                                        />
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}

                    {/* TAB: PRESUPUESTO Y LÍMITES */}
                    {tab === "budget" && (
                        <div style={{display: "flex", flexDirection: "column", gap: "16px"}}>
                            <div>
                                <label style={{display: "block", fontSize: "0.85rem", color: "var(--text)", marginBottom: "6px"}}>
                                    Max Tokens por Turno
                                </label>
                                <input
                                    type="number"
                                    value={cfg.max_tokens || 4096}
                                    onChange={(e) => setCfg({...cfg, max_tokens: parseInt(e.target.value, 10) || 4096})}
                                    style={{
                                        width: "100%",
                                        padding: "8px 12px",
                                        borderRadius: "6px",
                                        border: "1px solid var(--border)",
                                        backgroundColor: "var(--code-bg)",
                                        color: "var(--text)",
                                        boxSizing: "border-box",
                                    }}
                                />
                            </div>
                            <div>
                                <label style={{display: "block", fontSize: "0.85rem", color: "var(--text)", marginBottom: "6px"}}>
                                    Temperatura ({cfg.temperature ?? 0.7})
                                </label>
                                <input
                                    type="range"
                                    min="0"
                                    max="1"
                                    step="0.05"
                                    value={cfg.temperature ?? 0.7}
                                    onChange={(e) => setCfg({...cfg, temperature: parseFloat(e.target.value)})}
                                    style={{width: "100%", cursor: "pointer"}}
                                />
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer Actions */}
                <div id="settings-footer" style={{
                    padding: "16px 24px",
                    borderTop: "1px solid var(--border)",
                    display: "flex",
                    justifyContent: "flex-end",
                    gap: "12px",
                    backgroundColor: "var(--bg)",
                }}>
                    <button
                        type="button"
                        onClick={onClose}
                        style={{
                            padding: "8px 20px",
                            borderRadius: "8px",
                            border: "1px solid var(--border)",
                            backgroundColor: "var(--bg-elevated)",
                            color: "var(--text)",
                            cursor: "pointer",
                        }}
                    >
                        {t(lang, "settings.close")}
                    </button>
                    <button
                        type="button"
                        onClick={handleSave}
                        disabled={saving}
                        style={{
                            padding: "8px 24px",
                            borderRadius: "8px",
                            border: "none",
                            backgroundColor: "var(--accent)",
                            color: "var(--on-accent)",
                            fontWeight: 700,
                            cursor: "pointer",
                        }}
                    >
                        {saving ? "…" : t(lang, "providers.save_button")}
                    </button>
                </div>
            </div>
        </div>
    );
}