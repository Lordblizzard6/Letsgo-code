import React from "react";

interface IconProps {
    size?: number;
    className?: string;
    style?: React.CSSProperties;
}

// 1. Anthropic / Claude - Official 12-point sunburst
export function ClaudeIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="currentColor"
            className={`provider-icon provider-claude ${className}`}
            style={{ color: "#D97757", ...style }}
            aria-label="Claude"
        >
            <path d="M13.67 2.5a.75.75 0 0 0-1.34 0l-1.37 2.92a.75.75 0 0 1-.44.43l-3.08.97a.75.75 0 0 0-.42 1.28l2.36 2.2a.75.75 0 0 1 .2.62l-.65 3.17a.75.75 0 0 0 1.1.8l2.8-1.58a.75.75 0 0 1 .68 0l2.8 1.58a.75.75 0 0 0 1.1-.8l-.65-3.17a.75.75 0 0 1 .2-.62l2.36-2.2a.75.75 0 0 0-.42-1.28l-3.08-.97a.75.75 0 0 1-.44-.43L13.67 2.5Z" />
            <circle cx="12" cy="12" r="3.2" />
        </svg>
    );
}

// 2. OpenAI - Official 6-fold spiral
export function OpenAIIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={`provider-icon provider-openai ${className}`}
            style={{ color: "#10A37F", ...style }}
            aria-label="OpenAI"
        >
            <path d="M12 2a4 4 0 0 1 3.8 2.8l.2 1 .9-.4a4 4 0 0 1 5.2 2.3 4 4 0 0 1-.3 4.2l-.7.8.6.8a4 4 0 0 1-.4 5.3 4 4 0 0 1-4.1 1.2l-1-.3v1a4 4 0 0 1-3.8 2.8 4 4 0 0 1-3.8-2.8l-.2-1-.9.4a4 4 0 0 1-5.2-2.3 4 4 0 0 1 .3-4.2l.7-.8-.6-.8a4 4 0 0 1 .4-5.3 4 4 0 0 1 4.1-1.2l1 .3v-1A4 4 0 0 1 12 2Z" />
            <path d="M12 8.5v7" />
            <path d="M9 10.5 15 14" />
            <path d="m9 14 6-3.5" />
        </svg>
    );
}

// 3. Google Gemini - Official 4-point sparkle star
export function GeminiIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="currentColor"
            className={`provider-icon provider-gemini ${className}`}
            style={{ color: "#4E8BF5", ...style }}
            aria-label="Gemini"
        >
            <path d="M12 2C12 7.52 7.52 12 2 12c5.48 0 10 4.48 10 10 0-5.52 4.48-10 10-10-5.52 0-10-4.48-10-10Z" />
        </svg>
    );
}

// 4. Groq - Official Groq speed 'g'
export function GroqIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="currentColor"
            className={`provider-icon provider-groq ${className}`}
            style={{ color: "#F55036", ...style }}
            aria-label="Groq"
        >
            <circle cx="12" cy="12" r="9" fill="none" stroke="currentColor" strokeWidth="2.4" />
            <path d="M15 12h-4a2 2 0 0 0-2 2v1a2 2 0 0 0 2 2h3v-3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" fill="none" />
        </svg>
    );
}

// 5. Ollama - Official Ollama llama
export function OllamaIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="currentColor"
            className={`provider-icon provider-ollama ${className}`}
            style={{ color: "#E0E0E0", ...style }}
            aria-label="Ollama"
        >
            <path d="M8 3a2 2 0 0 0-2 2v2a2 2 0 0 0 2 2h1v4H8a3 3 0 0 0-3 3v4a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-2h6v2a1 1 0 0 0 1 1h2a1 1 0 0 0 1-1v-4a3 3 0 0 0-3-3h-1V9h1a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2h-1V2a1 1 0 0 0-2 0v1h-2V2a1 1 0 0 0-2 0v1H8Z" />
            <circle cx="9.5" cy="5.5" r="0.75" fill="#121212" />
            <circle cx="14.5" cy="5.5" r="0.75" fill="#121212" />
        </svg>
    );
}

// 6. DeepSeek - Official DeepSeek wave / fin
export function DeepSeekIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="currentColor"
            className={`provider-icon provider-deepseek ${className}`}
            style={{ color: "#1D75F0", ...style }}
            aria-label="DeepSeek"
        >
            <path d="M3 13.5C4.5 9 8.5 6 13 6c5.5 0 8 4 8 8 0 4-3.5 6.5-8 6.5C8 20.5 4.5 17.5 3 13.5Z" opacity="0.4" />
            <path d="M7 13.5c1-3 3.5-5 6.5-5 3.5 0 5.5 2.5 5.5 5.5 0 3-2.5 5-5.5 5-3 0-5.5-2.5-6.5-5.5Z" />
            <circle cx="14.5" cy="11.5" r="1.2" fill="#FFFFFF" />
        </svg>
    );
}

// 7. OpenRouter - Official geometric routes
export function OpenRouterIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={`provider-icon provider-openrouter ${className}`}
            style={{ color: "#6366F1", ...style }}
            aria-label="OpenRouter"
        >
            <circle cx="6" cy="6" r="3" fill="#6366F1" />
            <circle cx="18" cy="6" r="3" />
            <circle cx="12" cy="18" r="3" fill="#6366F1" />
            <path d="m8.5 7.5 7 0" />
            <path d="m7.5 8.5 3 7" />
            <path d="m16.5 8.5-3 7" />
        </svg>
    );
}

// Generic Bot / Brain Fallback
export function BotDefaultIcon({ size = 16, className = "", style = {} }: IconProps) {
    return (
        <svg
            width={size}
            height={size}
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className={`provider-icon provider-default ${className}`}
            style={style}
            aria-label="AI Model"
        >
            <rect x="3" y="11" width="18" height="10" rx="2" />
            <circle cx="8.5" cy="15.5" r="1.5" fill="currentColor" />
            <circle cx="15.5" cy="15.5" r="1.5" fill="currentColor" />
            <path d="M12 2v4" />
            <path d="M8 6h8" />
        </svg>
    );
}

// Resolves provider/model string to the correct real AI logo
export function ProviderIcon({
    provider = "",
    model = "",
    size = 16,
    className = "",
    style = {},
}: {
    provider?: string;
    model?: string;
    size?: number;
    className?: string;
    style?: React.CSSProperties;
}) {
    const p = (provider || "").toLowerCase();
    const m = (model || "").toLowerCase();

    if (p === "anthropic" || m.includes("claude") || m.includes("sonnet") || m.includes("haiku") || m.includes("opus")) {
        return <ClaudeIcon size={size} className={className} style={style} />;
    }
    if (p === "openai" || m.includes("gpt") || m.includes("o1") || m.includes("o3") || m.includes("chatgpt")) {
        return <OpenAIIcon size={size} className={className} style={style} />;
    }
    if (p === "gemini" || m.includes("gemini") || m.includes("gemma")) {
        return <GeminiIcon size={size} className={className} style={style} />;
    }
    if (p === "groq" || m.includes("groq")) {
        return <GroqIcon size={size} className={className} style={style} />;
    }
    if (p === "ollama" || m.includes("llama") || m.includes("mistral") || m.includes("qwen") || m.includes("phi")) {
        return <OllamaIcon size={size} className={className} style={style} />;
    }
    if (p === "deepseek" || m.includes("deepseek")) {
        return <DeepSeekIcon size={size} className={className} style={style} />;
    }
    if (p === "openrouter" || m.includes("/")) {
        return <OpenRouterIcon size={size} className={className} style={style} />;
    }

    return <BotDefaultIcon size={size} className={className} style={style} />;
}
