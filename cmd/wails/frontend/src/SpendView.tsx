import {useCallback, useEffect, useMemo, useState} from "react";
import type {SpendCategory, SpendEntry} from "./types";
import {t, type Lang} from "./i18n";

interface Props {
    lang: Lang;
    onClose: () => void;
}

type RangeKey = 7 | 30 | 90;

const CAT_COLORS: Record<SpendCategory, string> = {
    normal: "var(--accent, #8b5cf6)",
    web: "#2fbf71",
    test: "#f5a623",
    skills: "#ff6b5f",
};

const DAY_KEYS: RangeKey[] = [7, 30, 90];

export function SpendView({lang, onClose}: Props) {
    const [days, setDays] = useState<RangeKey>(30);
    const [entries, setEntries] = useState<SpendEntry[]>([]);

    const load = useCallback((d: RangeKey) => {
        // TODO: const data = await GetSpendEntries(d);
        // setEntries(data);
    }, []);

    useEffect(() => {
        load(days);
    }, [days, load]);

    useEffect(() => {
        const onKey = (e: KeyboardEvent) => {
            if (e.key === "Escape") onClose();
        };
        window.addEventListener("keydown", onKey);
        return () => window.removeEventListener("keydown", onKey);
    }, [onClose]);

    const gridDays = useMemo(() => {
        const out: {day: string; items: SpendEntry[]; total: number}[] = [];
        const now = new Date();
        const idx = new Map<string, number>();
        for (let i = days - 1; i >= 0; i--) {
            const d = new Date(now);
            d.setDate(now.getDate() - i);
            const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
            idx.set(key, out.length);
            out.push({day: key, items: [], total: 0});
        }
        for (const e of entries) {
            const d = new Date(e.ts);
            const key = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
            const gi = idx.get(key);
            if (gi === undefined) continue;
            out[gi].items.push(e);
            out[gi].total += e.cost_usd;
        }
        return out;
    }, [entries, days]);

    const clear = useCallback(() => {
        // TODO: await ClearSpend();
        setEntries([]);
    }, []);

    const totalTokens = entries.reduce((s, e) => s + e.tokens, 0);
    const totalCost = entries.reduce((s, e) => s + e.cost_usd, 0);

    return (
        <div id="spend-overlay" onClick={onClose}>
            <div
                id="spend-modal"
                role="dialog"
                aria-modal="true"
                aria-label={t(lang, "spend.title")}
                onClick={(e) => e.stopPropagation()}
            >
                <div id="spend-header">
                    <h2>{t(lang, "spend.title")}</h2>
                    <button type="button" className="btn icon-btn" onClick={onClose} title={t(lang, "settings.close")}>
                        ✕
                    </button>
                </div>

                <div id="spend-tabs">
                    {DAY_KEYS.map((d) => (
                        <button key={d}
                            type="button"
                            className={`tab-btn ${days === d ? "active" : ""}`}
                            onClick={() => setDays(d)}
                        >
                            {d === 7 ? t(lang, "spend.days_7") : d === 30 ? t(lang, "spend.days_30") : t(lang, "spend.days_90")}
                        </button>
                    ))}
                    <span className="spend-tabs-spacer" />
                    <button type="button" className="spend-clear" onClick={clear} title={t(lang, "spend.clear")}>
                        {t(lang, "spend.clear")}
                    </button>
                </div>

                <div id="spend-body">
                    <div className="spend-summary">
                        <div className="sum-cell">
                            <span className="sum-num">{totalTokens.toLocaleString(lang === "es" ? "es-ES" : "en-US")}</span>
                            <span className="sum-lbl">{t(lang, "spend.tokens", {n: totalTokens})}</span>
                        </div>
                        <div className="sum-cell">
                            <span className="sum-num">${totalCost.toFixed(4)}</span>
                            <span className="sum-lbl">{t(lang, "spend.cost", {cost: totalCost.toFixed(4)})}</span>
                        </div>
                    </div>

                    {entries.length === 0 && (
                        <div className="spend-empty">{t(lang, "spend.empty")}</div>
                    )}

                    {entries.length > 0 && (
                        <>
                            <div className="spend-legend">
                                <span className="legend-item"><i style={{background: CAT_COLORS.normal}} /> {t(lang, "spend.legend_normal")}</span>
                                <span className="legend-item"><i style={{background: CAT_COLORS.web}} /> {t(lang, "spend.legend_web")}</span>
                                <span className="legend-item"><i style={{background: CAT_COLORS.test}} /> {t(lang, "spend.legend_test")}</span>
                            </div>
                            <div className="spend-heatmap">
                                {gridDays.map((gd) => (
                                    <div key={gd.day} className="heat-col" title={`${gd.day} · ${gd.items.length} · $${gd.total.toFixed(4)}`}>
                                        {gd.items.length === 0 ? (
                                            <div className="heat-empty" />
                                        ) : (
                                            gd.items.slice(0, 4).map((e, i) => (
                                                <div key={i}
                                                    className="heat-cell"
                                                    style={{background: palette(e.category, e.cost_usd)}}
                                                    title={`${gd.day} · ${CAT_LABELS[e.category] ?? e.category} · ${e.tokens} tok · $${e.cost_usd.toFixed(4)}`}
                                                />
                                            ))
                                        )}
                                        {gd.items.length > 4 && <div className="heat-more">+{gd.items.length - 4}</div>}
                                    </div>
                                ))}
                            </div>
                        </>
                    )}

                    {entries.length > 0 && (
                        <div className="spend-list">
                            {[...entries].reverse().slice(0, 60).map((e, i) => (
                                <div key={`${e.ts}-${i}`} className="spend-item">
                                    <i className="spend-item-dot" style={{background: CAT_COLORS[e.category] ?? "#888"}} />
                                    <span className="spend-item-cat">{CAT_LABELS[e.category] ?? e.category}</span>
                                    <span className="spend-item-model">{e.model}</span>
                                    <span className="spend-item-tok">{e.tokens}</span>
                                    <span className="spend-item-cost">${e.cost_usd.toFixed(4)}</span>
                                    <time className="spend-item-date">{fmtTime(e.ts, lang)}</time>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}

const CAT_LABELS: Record<SpendCategory, string> = {
    normal: "Chat",
    web: "Web",
    test: "Tests",
    skills: "Skills",
};

function palette(cat: SpendCategory, cost: number): string {
    const base = CAT_COLORS[cat] ?? "#888";
    const inten = cost <= 0 ? 40 : Math.min(100, 40 + Math.round(Math.log1p(cost * 1000) * 18));
    return `color-mix(in srgb, ${base} ${inten}%, transparent)`;
}

function fmtTime(ts: number, lang: Lang): string {
    const d = new Date(ts);
    return d.toLocaleDateString(lang === "es" ? "es-ES" : "en-US") + " " + d.toLocaleTimeString([], {hour: "2-digit", minute: "2-digit"});
}