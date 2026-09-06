import {useState} from "react";
import {t, type Lang} from "./i18n";
import {SettingsService} from "./bindings";
import type {ConfigView} from "./types";

interface Props {
    lang: Lang;
    onLangChange: (lang: Lang) => void;
    onComplete: (updatedCfg: ConfigView) => void;
    onClose: () => void;
}

type ProviderChoice = "ollama" | "anthropic" | "openai" | "openrouter" | "groq" | "gemini" | "deepseek";

export function OnboardingView({lang, onLangChange, onComplete, onClose}: Props) {
    const [step, setStep] = useState<1 | 2 | 3>(1);
    const [selectedProvider, setSelectedProvider] = useState<ProviderChoice>("openrouter");
    const [apiKey, setApiKey] = useState("");
    const [ollamaUrl, setOllamaUrl] = useState("http://localhost:11434");
    const [saving, setSaving] = useState(false);

    const handleSaveAndFinish = async () => {
        setSaving(true);
        const partial: Record<string, any> = {
            language: lang,
        };

        if (selectedProvider === "ollama") {
            partial.ollama_base_url = ollamaUrl;
            partial.OllamaBaseURL = ollamaUrl;
            partial.model = "llama3.2";
            partial.Model = "llama3.2";
        } else if (selectedProvider === "anthropic") {
            partial.anthropic_api_key = apiKey.trim();
            partial.AnthropicAPIKey = apiKey.trim();
            partial.model = "claude-3-5-sonnet-20241022";
            partial.Model = "claude-3-5-sonnet-20241022";
        } else if (selectedProvider === "openai") {
            partial.openai_api_key = apiKey.trim();
            partial.OpenAIAPIKey = apiKey.trim();
            partial.model = "gpt-4o";
            partial.Model = "gpt-4o";
        } else if (selectedProvider === "openrouter") {
            partial.openrouter_api_key = apiKey.trim();
            partial.OpenRouterAPIKey = apiKey.trim();
            partial.model = "anthropic/claude-3.5-sonnet";
            partial.Model = "anthropic/claude-3.5-sonnet";
        } else if (selectedProvider === "groq") {
            partial.groq_api_key = apiKey.trim();
            partial.GroqAPIKey = apiKey.trim();
            partial.model = "llama-3.3-70b-versatile";
            partial.Model = "llama-3.3-70b-versatile";
        } else if (selectedProvider === "gemini") {
            partial.gemini_api_key = apiKey.trim();
            partial.GeminiAPIKey = apiKey.trim();
            partial.model = "gemini-2.0-flash";
            partial.Model = "gemini-2.0-flash";
        } else if (selectedProvider === "deepseek") {
            partial.deepseek_api_key = apiKey.trim();
            partial.DeepSeekAPIKey = apiKey.trim();
            partial.model = "deepseek-chat";
            partial.Model = "deepseek-chat";
        }

        try {
            const updated = await SettingsService.SaveConfig(partial);
            setSaving(false);
            onComplete(updated as any);
        } catch {
            setSaving(false);
            onClose();
        }
    };

    return (
        <div className="onboarding-overlay" style={{
            position: "fixed",
            inset: 0,
            backgroundColor: "rgba(10, 10, 15, 0.85)",
            backdropFilter: "blur(8px)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 9999,
            padding: "20px",
        }}>
            <div className="onboarding-card" style={{
                backgroundColor: "var(--card, #1e1e2e)",
                borderRadius: "16px",
                border: "1px solid var(--border, #313244)",
                boxShadow: "0 25px 50px -12px rgba(0, 0, 0, 0.5)",
                width: "100%",
                maxWidth: "680px",
                overflow: "hidden",
                display: "flex",
                flexDirection: "column",
            }}>
                {/* Header */}
                <div style={{
                    padding: "24px 32px",
                    borderBottom: "1px solid var(--border, #313244)",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                }}>
                    <div>
                        <h2 style={{margin: 0, fontSize: "1.35rem", fontWeight: 700, color: "var(--fg, #cdd6f4)"}}>
                            {t(lang, "onboarding.welcome")}
                        </h2>
                        <p style={{margin: "4px 0 0", fontSize: "0.85rem", color: "var(--text-muted, #a6adc8)"}}>
                            {t(lang, "onboarding.subtitle")}
                        </p>
                    </div>
                    <button
                        type="button"
                        onClick={onClose}
                        style={{
                            background: "transparent",
                            border: "none",
                            color: "var(--text-muted, #a6adc8)",
                            cursor: "pointer",
                            fontSize: "1.2rem",
                            padding: "4px 8px",
                            borderRadius: "6px",
                        }}
                        title={t(lang, "onboarding.skip")}
                    >
                        ✕
                    </button>
                </div>

                {/* Body */}
                <div style={{padding: "32px", minHeight: "320px"}}>
                    {step === 1 && (
                        <div>
                            <h3 style={{marginTop: 0, marginBottom: "16px", fontSize: "1.1rem"}}>
                                {t(lang, "onboarding.step1_title")}
                            </h3>
                            <div style={{display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px", marginTop: "24px"}}>
                                <button
                                    type="button"
                                    onClick={() => onLangChange("es")}
                                    style={{
                                        padding: "20px",
                                        borderRadius: "12px",
                                        border: lang === "es" ? "2px solid #89b4fa" : "1px solid #313244",
                                        backgroundColor: lang === "es" ? "rgba(137, 180, 250, 0.1)" : "#181825",
                                        color: "var(--fg, #cdd6f4)",
                                        cursor: "pointer",
                                        textAlign: "center",
                                        display: "flex",
                                        flexDirection: "column",
                                        alignItems: "center",
                                        gap: "8px",
                                    }}
                                >
                                    <span style={{fontSize: "2rem"}}>🇪🇸</span>
                                    <strong style={{fontSize: "1.05rem"}}>Español</strong>
                                    <span style={{fontSize: "0.8rem", color: "#a6adc8"}}>Interfaz en español</span>
                                </button>
                                <button
                                    type="button"
                                    onClick={() => onLangChange("en")}
                                    style={{
                                        padding: "20px",
                                        borderRadius: "12px",
                                        border: lang === "en" ? "2px solid #89b4fa" : "1px solid #313244",
                                        backgroundColor: lang === "en" ? "rgba(137, 180, 250, 0.1)" : "#181825",
                                        color: "var(--fg, #cdd6f4)",
                                        cursor: "pointer",
                                        textAlign: "center",
                                        display: "flex",
                                        flexDirection: "column",
                                        alignItems: "center",
                                        gap: "8px",
                                    }}
                                >
                                    <span style={{fontSize: "2rem"}}>🇺🇸</span>
                                    <strong style={{fontSize: "1.05rem"}}>English</strong>
                                    <span style={{fontSize: "0.8rem", color: "#a6adc8"}}>English interface</span>
                                </button>
                            </div>
                        </div>
                    )}

                    {step === 2 && (
                        <div>
                            <h3 style={{marginTop: 0, marginBottom: "8px", fontSize: "1.1rem"}}>
                                {t(lang, "onboarding.step2_title")}
                            </h3>
                            <p style={{margin: "0 0 16px", fontSize: "0.85rem", color: "#a6adc8"}}>
                                {t(lang, "providers.encryption_notice")}
                            </p>

                            <div style={{display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "10px", marginBottom: "20px"}}>
                                {[
                                    {id: "openrouter", name: "OpenRouter", icon: "🌐"},
                                    {id: "anthropic", name: "Anthropic", icon: "🟣"},
                                    {id: "openai", name: "OpenAI", icon: "🟢"},
                                    {id: "groq", name: "Groq", icon: "⚡"},
                                    {id: "gemini", name: "Gemini", icon: "🔵"},
                                    {id: "ollama", name: "Ollama (Local)", icon: "🚀"},
                                ].map((p) => (
                                    <button
                                        type="button"
                                        key={p.id}
                                        onClick={() => {
                                            setSelectedProvider(p.id as ProviderChoice);
                                            setApiKey("");
                                        }}
                                        style={{
                                            padding: "12px 10px",
                                            borderRadius: "8px",
                                            border: selectedProvider === p.id ? "2px solid #89b4fa" : "1px solid #313244",
                                            backgroundColor: selectedProvider === p.id ? "rgba(137, 180, 250, 0.1)" : "#181825",
                                            color: "var(--fg, #cdd6f4)",
                                            cursor: "pointer",
                                            display: "flex",
                                            alignItems: "center",
                                            gap: "8px",
                                            fontSize: "0.88rem",
                                        }}
                                    >
                                        <span>{p.icon}</span>
                                        <strong>{p.name}</strong>
                                    </button>
                                ))}
                            </div>

                            {selectedProvider === "ollama" ? (
                                <div style={{backgroundColor: "#181825", padding: "16px", borderRadius: "8px", border: "1px solid #313244"}}>
                                    <label style={{display: "block", fontSize: "0.85rem", marginBottom: "6px", color: "#cdd6f4"}}>
                                        {t(lang, "providers.ollama_url")}
                                    </label>
                                    <input
                                        type="text"
                                        value={ollamaUrl}
                                        onChange={(e) => setOllamaUrl(e.target.value)}
                                        style={{
                                            width: "100%",
                                            padding: "10px 12px",
                                            borderRadius: "6px",
                                            border: "1px solid #45475a",
                                            backgroundColor: "#11111b",
                                            color: "#cdd6f4",
                                            boxSizing: "border-box",
                                        }}
                                    />
                                    <small style={{display: "block", marginTop: "6px", color: "#a6adc8"}}>
                                        {t(lang, "providers.ollama_desc")}
                                    </small>
                                </div>
                            ) : (
                                <div style={{backgroundColor: "#181825", padding: "16px", borderRadius: "8px", border: "1px solid #313244"}}>
                                    <div style={{display: "flex", justifyContent: "space-between", marginBottom: "6px"}}>
                                        <label style={{fontSize: "0.85rem", color: "#cdd6f4"}}>
                                            API Key ({selectedProvider})
                                        </label>
                                        <span style={{fontSize: "0.75rem", color: "#a6e3a1"}}>
                                            {t(lang, "providers.encryption_badge", undefined) || "🔒 AES-256"}
                                        </span>
                                    </div>
                                    <input
                                        type="password"
                                        placeholder={t(lang, "providers.key_placeholder")}
                                        value={apiKey}
                                        onChange={(e) => setApiKey(e.target.value)}
                                        style={{
                                            width: "100%",
                                            padding: "10px 12px",
                                            borderRadius: "6px",
                                            border: "1px solid #45475a",
                                            backgroundColor: "#11111b",
                                            color: "#cdd6f4",
                                            boxSizing: "border-box",
                                        }}
                                    />
                                </div>
                            )}
                        </div>
                    )}

                    {step === 3 && (
                        <div style={{textAlign: "center", padding: "24px 0"}}>
                            <span style={{fontSize: "3.5rem"}}>🎉</span>
                            <h3 style={{margin: "16px 0 8px", fontSize: "1.3rem"}}>
                                {t(lang, "onboarding.step3_title")}
                            </h3>
                            <p style={{maxWidth: "460px", margin: "0 auto 24px", color: "#a6adc8", fontSize: "0.95rem"}}>
                                {t(lang, "onboarding.step3_desc")}
                            </p>
                            <div style={{
                                display: "inline-flex",
                                alignItems: "center",
                                gap: "8px",
                                padding: "8px 16px",
                                borderRadius: "20px",
                                backgroundColor: "rgba(166, 227, 161, 0.15)",
                                color: "#a6e3a1",
                                fontSize: "0.85rem",
                                fontWeight: 600,
                            }}>
                                🔒 {t(lang, "providers.encryption_notice")}
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer Controls */}
                <div style={{
                    padding: "20px 32px",
                    borderTop: "1px solid var(--border, #313244)",
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                    backgroundColor: "rgba(17, 17, 27, 0.4)",
                }}>
                    {step > 1 ? (
                        <button
                            type="button"
                            onClick={() => setStep((s) => (s - 1) as any)}
                            style={{
                                padding: "8px 18px",
                                borderRadius: "8px",
                                border: "1px solid #45475a",
                                backgroundColor: "transparent",
                                color: "#cdd6f4",
                                cursor: "pointer",
                            }}
                        >
                            {t(lang, "onboarding.back")}
                        </button>
                    ) : (
                        <button
                            type="button"
                            onClick={onClose}
                            style={{
                                background: "none",
                                border: "none",
                                color: "#6c7086",
                                cursor: "pointer",
                                fontSize: "0.85rem",
                            }}
                        >
                            {t(lang, "onboarding.skip")}
                        </button>
                    )}

                    {step < 3 ? (
                        <button
                            type="button"
                            onClick={() => setStep((s) => (s + 1) as any)}
                            style={{
                                padding: "8px 24px",
                                borderRadius: "8px",
                                border: "none",
                                backgroundColor: "#89b4fa",
                                color: "#11111b",
                                fontWeight: 700,
                                cursor: "pointer",
                            }}
                        >
                            {t(lang, "onboarding.next")}
                        </button>
                    ) : (
                        <button
                            type="button"
                            onClick={handleSaveAndFinish}
                            disabled={saving}
                            style={{
                                padding: "10px 28px",
                                borderRadius: "8px",
                                border: "none",
                                backgroundColor: "#a6e3a1",
                                color: "#11111b",
                                fontWeight: 700,
                                cursor: "pointer",
                            }}
                        >
                            {saving ? "…" : t(lang, "onboarding.finish")}
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
}
